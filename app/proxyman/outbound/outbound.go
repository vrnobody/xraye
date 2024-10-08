package outbound

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/xtls/xray-core/app/proxyman"
	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/outbound"
)

// Manager is to manage all outbound handlers.
type Manager struct {
	access           sync.RWMutex
	defaultHandler   outbound.Handler
	taggedHandler    map[string]outbound.Handler
	untaggedHandlers []outbound.Handler
	running          bool
	tagsCache        *sync.Map
}

// New creates a new Manager.
func New(ctx context.Context, config *proxyman.OutboundConfig) (*Manager, error) {
	m := &Manager{
		taggedHandler: make(map[string]outbound.Handler),
		tagsCache:     &sync.Map{},
	}
	var _ outbound.Manager = m
	return m, nil
}

// Type implements common.HasType.
func (m *Manager) Type() interface{} {
	return outbound.ManagerType()
}

// Start implements core.Feature
func (m *Manager) Start() error {
	m.access.Lock()
	defer m.access.Unlock()

	m.running = true

	for _, h := range m.taggedHandler {
		if err := h.Start(); err != nil {
			return err
		}
	}

	for _, h := range m.untaggedHandlers {
		if err := h.Start(); err != nil {
			return err
		}
	}

	return nil
}

// Close implements core.Feature
func (m *Manager) Close() error {
	m.access.Lock()
	defer m.access.Unlock()

	m.running = false

	var errs []error
	for _, h := range m.taggedHandler {
		errs = append(errs, h.Close())
	}

	for _, h := range m.untaggedHandlers {
		errs = append(errs, h.Close())
	}

	return errors.Combine(errs...)
}

// GetDefaultHandler implements outbound.Manager.
func (m *Manager) GetDefaultHandler() outbound.Handler {
	m.access.RLock()
	defer m.access.RUnlock()

	if m.defaultHandler == nil {
		return nil
	}
	return m.defaultHandler
}

// GetHandler implements outbound.Manager.
func (m *Manager) GetHandler(tag string) outbound.Handler {
	m.access.RLock()
	defer m.access.RUnlock()
	switch t := proxyman.ParseTag(tag).(type) {
	case int:
		if t >= 0 && t < len(m.untaggedHandlers) {
			return m.untaggedHandlers[t]
		}
	case string:
		if h, ok := m.taggedHandler[t]; ok {
			return h
		}
	}
	return nil
}

// AddHandler implements outbound.Manager.
func (m *Manager) AddHandler(ctx context.Context, handler outbound.Handler) error {
	m.access.Lock()
	defer m.access.Unlock()

	m.tagsCache = &sync.Map{}

	if m.defaultHandler == nil {
		m.defaultHandler = handler
	}

	tag := handler.Tag()
	if len(tag) > 0 {
		if _, found := m.taggedHandler[tag]; found {
			return errors.New("existing tag found: " + tag)
		}
		m.taggedHandler[tag] = handler
	} else {
		m.untaggedHandlers = append(m.untaggedHandlers, handler)
	}

	if m.running {
		return handler.Start()
	}

	return nil
}

// GetAllHandlers returns all handlers.
func (m *Manager) GetAllHandlers(ctx context.Context) ([]outbound.Handler, error) {
	m.access.RLock()
	defer m.access.RUnlock()

	if size := len(m.untaggedHandlers) + len(m.taggedHandler); size > 0 {
		hs := make([]outbound.Handler, 0)
		hs = append(hs, m.untaggedHandlers...)
		for _, h := range m.taggedHandler {
			hs = append(hs, h)
		}
		return hs, nil
	}
	return nil, errors.New("no handler found")
}

// RemoveHandler implements outbound.Manager.
func (m *Manager) RemoveHandler(ctx context.Context, tag string) error {
	m.access.Lock()
	defer m.access.Unlock()

	m.tagsCache = &sync.Map{}

	var handler outbound.Handler
	switch t := proxyman.ParseTag(tag).(type) {
	case int:
		if t >= 0 && t < len(m.untaggedHandlers) {
			handler = m.untaggedHandlers[t]
			uh := m.untaggedHandlers
			m.untaggedHandlers = append(uh[:t], uh[t+1:]...)
		} else {
			emsg := fmt.Sprintf("handler #%d index out of range", t)
			errors.LogWarning(ctx, emsg)
			return errors.New(emsg)
		}
	case string:
		if h, ok := m.taggedHandler[t]; ok {
			handler = h
		} else {
			// do not return error here
			// app/commander/commander.go Start() will call RemoveHandler("api") and expceting nil.

			// suppress editor wraning
			_ = 0
		}
		delete(m.taggedHandler, t)
	case *errors.Error:
		return t
	}
	if handler != nil && handler == m.defaultHandler {
		m.defaultHandler = nil
	}
	return nil
}

// Select implements outbound.HandlerSelector.
func (m *Manager) Select(selectors []string) []string {

	key := strings.Join(selectors, ",")
	if cache, ok := m.tagsCache.Load(key); ok {
		return cache.([]string)
	}

	m.access.RLock()
	defer m.access.RUnlock()

	tags := make([]string, 0, len(selectors))

	for tag := range m.taggedHandler {
		for _, selector := range selectors {
			if strings.HasPrefix(tag, selector) {
				tags = append(tags, tag)
				break
			}
		}
	}

	sort.Strings(tags)
	m.tagsCache.Store(key, tags)

	return tags
}

func init() {
	common.Must(common.RegisterConfig((*proxyman.OutboundConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*proxyman.OutboundConfig))
	}))
	common.Must(common.RegisterConfig((*core.OutboundHandlerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewHandler(ctx, config.(*core.OutboundHandlerConfig))
	}))
}
