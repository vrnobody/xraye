package command

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/errors"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/core"

	"github.com/xtls/xray-core/features/inbound"
	"github.com/xtls/xray-core/features/outbound"
	"github.com/xtls/xray-core/proxy"
	grpc "google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// InboundOperation is the interface for operations that applies to inbound handlers.
type InboundOperation interface {
	// ApplyInbound applies this operation to the given inbound handler.
	ApplyInbound(context.Context, inbound.Handler) error
}

// OutboundOperation is the interface for operations that applies to outbound handlers.
type OutboundOperation interface {
	// ApplyOutbound applies this operation to the given outbound handler.
	ApplyOutbound(context.Context, outbound.Handler) error
}

func getInbound(handler inbound.Handler) (proxy.Inbound, error) {
	gi, ok := handler.(proxy.GetInbound)
	if !ok {
		return nil, errors.New("can't get inbound proxy from handler.")
	}
	return gi.GetInbound(), nil
}

// ApplyInbound implements InboundOperation.
func (op *AddUserOperation) ApplyInbound(ctx context.Context, handler inbound.Handler) error {
	p, err := getInbound(handler)
	if err != nil {
		return err
	}
	um, ok := p.(proxy.UserManager)
	if !ok {
		return errors.New("proxy is not a UserManager")
	}
	mUser, err := op.User.ToMemoryUser()
	if err != nil {
		return errors.New("failed to parse user").Base(err)
	}
	return um.AddUser(ctx, mUser)
}

// ApplyInbound implements InboundOperation.
func (op *RemoveUserOperation) ApplyInbound(ctx context.Context, handler inbound.Handler) error {
	p, err := getInbound(handler)
	if err != nil {
		return err
	}
	um, ok := p.(proxy.UserManager)
	if !ok {
		return errors.New("proxy is not a UserManager")
	}
	return um.RemoveUser(ctx, op.Email)
}

type handlerServer struct {
	s   *core.Instance
	ihm inbound.Manager
	ohm outbound.Manager
}

func (s *handlerServer) GetAllInbounds(ctx context.Context, request *GetAllInboundsRequest) (*GetAllInboundsResponse, error) {
	configs := make([]*core.InboundHandlerConfig, 0)
	inboundConfigCache.Range(func(_ interface{}, config interface{}) bool {
		if c, ok := config.(*core.InboundHandlerConfig); ok {
			configs = append(configs, c)
		}
		return true
	})
	return &GetAllInboundsResponse{
		Configs: configs,
	}, nil
}

func (s *handlerServer) AddInbound(ctx context.Context, request *AddInboundRequest) (*AddInboundResponse, error) {
	inboundMutex.Lock()
	defer inboundMutex.Unlock()

	if err := core.AddInboundHandler(s.s, request.Inbound); err != nil {
		if hs, err := s.ihm.GetAllHandlers(ctx); err == nil {
			cleanupInboundConfigCache(hs)
		}
		return nil, err
	}
	return &AddInboundResponse{}, nil
}

func (s *handlerServer) RemoveInbound(ctx context.Context, request *RemoveInboundRequest) (*RemoveInboundResponse, error) {
	inboundMutex.Lock()
	defer inboundMutex.Unlock()

	if h, err := s.ihm.GetHandler(ctx, request.Tag); err == nil {
		if _, ok := inboundConfigCache.LoadAndDelete(h); ok {
			ht := reflect.TypeOf(h)
			errors.LogDebug(ctx, "remove ", ht, " from cache")
		}
	}
	return &RemoveInboundResponse{}, s.ihm.RemoveHandler(ctx, request.Tag)
}

func (s *handlerServer) AlterInbound(ctx context.Context, request *AlterInboundRequest) (*AlterInboundResponse, error) {
	rawOperation, err := request.Operation.GetInstance()
	if err != nil {
		return nil, errors.New("unknown operation").Base(err)
	}
	operation, ok := rawOperation.(InboundOperation)
	if !ok {
		return nil, errors.New("not an inbound operation")
	}

	handler, err := s.ihm.GetHandler(ctx, request.Tag)
	if err != nil {
		return nil, errors.New("failed to get handler: ", request.Tag).Base(err)
	}

	return &AlterInboundResponse{}, operation.ApplyInbound(ctx, handler)
}

func (s *handlerServer) GetInboundUsers(ctx context.Context, request *GetInboundUserRequest) (*GetInboundUserResponse, error) {
	handler, err := s.ihm.GetHandler(ctx, request.Tag)
	if err != nil {
		return nil, errors.New("failed to get handler: ", request.Tag).Base(err)
	}
	p, err := getInbound(handler)
	if err != nil {
		return nil, err
	}
	um, ok := p.(proxy.UserManager)
	if !ok {
		return nil, errors.New("proxy is not a UserManager")
	}
	if len(request.Email) > 0 {
		return &GetInboundUserResponse{Users: []*protocol.User{protocol.ToProtoUser(um.GetUser(ctx, request.Email))}}, nil
	}
	var result = make([]*protocol.User, 0, 100)
	users := um.GetUsers(ctx)
	for _, u := range users {
		result = append(result, protocol.ToProtoUser(u))
	}
	return &GetInboundUserResponse{Users: result}, nil
}

func (s *handlerServer) GetInboundUsersCount(ctx context.Context, request *GetInboundUserRequest) (*GetInboundUsersCountResponse, error) {
	handler, err := s.ihm.GetHandler(ctx, request.Tag)
	if err != nil {
		return nil, errors.New("failed to get handler: ", request.Tag).Base(err)
	}
	p, err := getInbound(handler)
	if err != nil {
		return nil, err
	}
	um, ok := p.(proxy.UserManager)
	if !ok {
		return nil, errors.New("proxy is not a UserManager")
	}
	return &GetInboundUsersCountResponse{Count: um.GetUsersCount(ctx)}, nil
}

func (s *handlerServer) GetAllOutbounds(ctx context.Context, request *GetAllOutboundsRequest) (*GetAllOutboundsResponse, error) {
	configs := make([]*core.OutboundHandlerConfig, 0)
	outboundConfigCache.Range(func(_ interface{}, config interface{}) bool {
		if c, ok := config.(*core.OutboundHandlerConfig); ok {
			configs = append(configs, c)
		}
		return true
	})
	return &GetAllOutboundsResponse{
		Configs: configs,
	}, nil
}

func (s *handlerServer) AddOutbound(ctx context.Context, request *AddOutboundRequest) (*AddOutboundResponse, error) {
	outboundMutex.Lock()
	defer outboundMutex.Unlock()

	if err := core.AddOutboundHandler(s.s, request.Outbound); err != nil {
		if hs, err := s.ohm.GetAllHandlers(ctx); err == nil {
			cleanupOutboundConfigCache(hs)
		}
		return nil, err
	}
	return &AddOutboundResponse{}, nil
}

func (s *handlerServer) RemoveOutbound(ctx context.Context, request *RemoveOutboundRequest) (*RemoveOutboundResponse, error) {
	outboundMutex.Lock()
	defer outboundMutex.Unlock()

	tag := request.Tag
	resp := &RemoveOutboundResponse{}

	if tag != "*" {
		h := s.ohm.GetHandler(tag)
		if _, ok := outboundConfigCache.LoadAndDelete(h); ok {
			ht := reflect.TypeOf(h)
			errors.LogDebug(ctx, "remove ", ht, " from config cache")
		}
		return resp, s.ohm.RemoveHandler(ctx, tag)
	}

	// remove untagged handlers
	for i := 0; true; {
		t := fmt.Sprintf("#%d", i)
		h := s.ohm.GetHandler(t)
		if h == nil {
			break
		}
		if _, ok := outboundConfigCache.LoadAndDelete(h); ok {
			if err := s.ohm.RemoveHandler(ctx, t); err != nil {
				return nil, err
			}
		} else {
			i += 1
		}
	}

	hs, err := s.ohm.GetAllHandlers(ctx)
	if err != nil {
		return nil, err
	}

	// make sure do not delete the api outbound
	for _, h := range hs {
		if t := h.Tag(); t != "" {
			if _, ok := outboundConfigCache.LoadAndDelete(h); ok {
				if err := s.ohm.RemoveHandler(ctx, t); err != nil {
					return nil, err
				}
			}
		}
	}
	errors.LogDebug(ctx, "all outbounds are removed from config cache")
	return resp, nil
}

func (s *handlerServer) AlterOutbound(ctx context.Context, request *AlterOutboundRequest) (*AlterOutboundResponse, error) {
	rawOperation, err := request.Operation.GetInstance()
	if err != nil {
		return nil, errors.New("unknown operation").Base(err)
	}
	operation, ok := rawOperation.(OutboundOperation)
	if !ok {
		return nil, errors.New("not an outbound operation")
	}

	handler := s.ohm.GetHandler(request.Tag)
	return &AlterOutboundResponse{}, operation.ApplyOutbound(ctx, handler)
}

func (s *handlerServer) mustEmbedUnimplementedHandlerServiceServer() {}

type service struct {
	v *core.Instance
}

func (s *service) Register(server *grpc.Server) {
	hs := &handlerServer{
		s: s.v,
	}
	common.Must(s.v.RequireFeatures(func(im inbound.Manager, om outbound.Manager) {
		hs.ihm = im
		hs.ohm = om
	}))
	RegisterHandlerServiceServer(server, hs)

	// For compatibility purposes
	vCoreDesc := HandlerService_ServiceDesc
	vCoreDesc.ServiceName = "v2ray.core.app.proxyman.command.HandlerService"
	server.RegisterService(&vCoreDesc, hs)
}

var (
	inboundConfigCache  sync.Map
	outboundConfigCache sync.Map
	inboundMutex        sync.Mutex
	outboundMutex       sync.Mutex
)

// cleanupInboundConfigCache remove handlers not in inbound manager
func cleanupInboundConfigCache(hs []inbound.Handler) {
	rm := make([]interface{}, 0)
	inboundConfigCache.Range(func(key interface{}, _ interface{}) bool {
		for _, h := range hs {
			if h == key {
				return true
			}
		}
		kt := reflect.TypeOf(key)
		errors.LogDebug(nil, "remove ", kt, " from cache")
		rm = append(rm, key)
		return true
	})
	for _, h := range rm {
		inboundConfigCache.Delete(h)
	}
}

// cleanupOutboundConfigCache remove handlers not in outbound manager
func cleanupOutboundConfigCache(hs []outbound.Handler) {
	rm := make([]interface{}, 0)
	outboundConfigCache.Range(func(key interface{}, _ interface{}) bool {
		for _, h := range hs {
			if h == key {
				return true
			}
		}
		kt := reflect.TypeOf(key)
		errors.LogDebug(nil, "remove ", kt, " from cache")
		rm = append(rm, key)
		return true
	})

	for _, h := range rm {
		outboundConfigCache.Delete(h)
	}
}

// interceptConfig cache in(out)bound config when handler is created
func interceptConfig(key interface{}, config interface{}) {
	if _, ok := config.(proto.Message); !ok {
		return
	}

	switch key.(type) {
	case inbound.Handler:
		inboundConfigCache.Store(key, config)
	case outbound.Handler:
		outboundConfigCache.Store(key, config)
	default:
		return
	}

	kt := reflect.TypeOf(key)
	ct := reflect.TypeOf(config)
	errors.LogDebug(nil, "add ", kt, " with config type ", ct, " to cache")
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, cfg interface{}) (interface{}, error) {
		common.ConfigIntercepterFn = interceptConfig
		s := core.MustFromContext(ctx)
		return &service{v: s}, nil
	}))
}
