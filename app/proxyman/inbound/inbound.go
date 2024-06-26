package inbound

//go:generate go run github.com/xtls/xray-core/common/errors/errorgen

import (
	"context"
	"sync"

	"github.com/xtls/xray-core/app/proxyman"
	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/inbound"
)

// Manager is to manage all inbound handlers.
type Manager struct {
	access          sync.RWMutex
	untaggedHandler []inbound.Handler
	taggedHandlers  map[string]inbound.Handler
	running         bool
}

// New returns a new Manager for inbound handlers.
func New(ctx context.Context, config *proxyman.InboundConfig) (*Manager, error) {
	m := &Manager{
		taggedHandlers: make(map[string]inbound.Handler),
	}
	var _ inbound.Manager = m
	return m, nil
}

// Type implements common.HasType.
func (*Manager) Type() interface{} {
	return inbound.ManagerType()
}

// GetAllHandlers returns all handlers.
func (m *Manager) GetAllHandlers(ctx context.Context) ([]inbound.Handler, error) {
	m.access.RLock()
	defer m.access.RUnlock()

	if size := len(m.untaggedHandler) + len(m.taggedHandlers); size > 0 {
		hs := make([]inbound.Handler, 0)
		hs = append(hs, m.untaggedHandler...)
		for _, h := range m.taggedHandlers {
			hs = append(hs, h)
		}
		return hs, nil
	}
	return nil, newError("no handler found")
}

// AddHandler implements inbound.Manager.
func (m *Manager) AddHandler(ctx context.Context, handler inbound.Handler) error {
	m.access.Lock()
	defer m.access.Unlock()

	tag := handler.Tag()
	if len(tag) > 0 {
		if _, found := m.taggedHandlers[tag]; found {
			return newError("existing tag found: " + tag)
		}
		m.taggedHandlers[tag] = handler
	} else {
		m.untaggedHandler = append(m.untaggedHandler, handler)
	}

	if m.running {
		return handler.Start()
	}

	return nil
}

// GetHandler implements inbound.Manager.
func (m *Manager) GetHandler(ctx context.Context, tag string) (inbound.Handler, error) {
	m.access.RLock()
	defer m.access.RUnlock()

	switch t := proxyman.ParseTag(tag).(type) {
	case int:
		if t >= 0 && t < len(m.untaggedHandler) {
			return m.untaggedHandler[t], nil
		}
	case string:
		if h, found := m.taggedHandlers[t]; found {
			return h, nil
		}
	}
	return nil, newError("handler not found: ", tag)
}

// RemoveHandler implements inbound.Manager.
func (m *Manager) RemoveHandler(ctx context.Context, tag string) error {
	m.access.Lock()
	defer m.access.Unlock()

	var handler inbound.Handler
	switch t := proxyman.ParseTag(tag).(type) {
	case int:
		if t >= 0 && t < len(m.untaggedHandler) {
			handler = m.untaggedHandler[t]
			uh := m.untaggedHandler
			m.untaggedHandler = append(uh[:t], uh[t+1:]...)
		} else {
			err := newError("handler #", t, ", index out of range").AtWarning()
			err.WriteToLog(session.ExportIDToError(ctx))
			return err
		}
	case string:
		if h, found := m.taggedHandlers[t]; found {
			handler = h
			delete(m.taggedHandlers, t)
		} else {
			err := newError("handler ", t, " not found").AtWarning()
			err.WriteToLog(session.ExportIDToError(ctx))
			return err
		}
	case error:
		return t
	}

	if handler != nil {
		if err := handler.Close(); err != nil {
			newError("failed to close handler ", tag).Base(err).AtWarning().WriteToLog(session.ExportIDToError(ctx))
		}
		return nil
	}
	return common.ErrNoClue
}

// Start implements common.Runnable.
func (m *Manager) Start() error {
	m.access.Lock()
	defer m.access.Unlock()

	m.running = true

	for _, handler := range m.taggedHandlers {
		if err := handler.Start(); err != nil {
			return err
		}
	}

	for _, handler := range m.untaggedHandler {
		if err := handler.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Close implements common.Closable.
func (m *Manager) Close() error {
	m.access.Lock()
	defer m.access.Unlock()

	m.running = false

	var errors []interface{}
	for _, handler := range m.taggedHandlers {
		if err := handler.Close(); err != nil {
			errors = append(errors, err)
		}
	}
	for _, handler := range m.untaggedHandler {
		if err := handler.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return newError("failed to close all handlers").Base(newError(serial.Concat(errors...)))
	}

	return nil
}

// NewHandler creates a new inbound.Handler based on the given config.
func NewHandler(ctx context.Context, config *core.InboundHandlerConfig) (inbound.Handler, error) {
	rawReceiverSettings, err := config.ReceiverSettings.GetInstance()
	if err != nil {
		return nil, err
	}
	proxySettings, err := config.ProxySettings.GetInstance()
	if err != nil {
		return nil, err
	}
	tag := config.Tag

	receiverSettings, ok := rawReceiverSettings.(*proxyman.ReceiverConfig)
	if !ok {
		return nil, newError("not a ReceiverConfig").AtError()
	}

	streamSettings := receiverSettings.StreamSettings
	if streamSettings != nil && streamSettings.SocketSettings != nil {
		ctx = session.ContextWithSockopt(ctx, &session.Sockopt{
			Mark: streamSettings.SocketSettings.Mark,
		})
	}

	allocStrategy := receiverSettings.AllocationStrategy
	if allocStrategy == nil || allocStrategy.Type == proxyman.AllocationStrategy_Always {
		return NewAlwaysOnInboundHandler(ctx, tag, receiverSettings, proxySettings)
	}

	if allocStrategy.Type == proxyman.AllocationStrategy_Random {
		return NewDynamicInboundHandler(ctx, tag, receiverSettings, proxySettings)
	}
	return nil, newError("unknown allocation strategy: ", receiverSettings.AllocationStrategy.Type).AtError()
}

func init() {
	common.Must(common.RegisterConfig((*proxyman.InboundConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*proxyman.InboundConfig))
	}))
	common.Must(common.RegisterConfig((*core.InboundHandlerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewHandler(ctx, config.(*core.InboundHandlerConfig))
	}))
}
