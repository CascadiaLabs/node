package main

import (
	"context"
	"crypto/subtle"
	"strings"

	pb "github.com/CascadiaLabs/node/proto/node"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// nodeServer реализует pb.NodeServiceServer поверх Manager.
type nodeServer struct {
	pb.UnimplementedNodeServiceServer
	mgr *Manager
}

func (s *nodeServer) UpdateConfig(_ context.Context, req *pb.UpdateConfigRequest) (*pb.UpdateConfigResponse, error) {
	if err := s.mgr.Apply([]byte(req.GetConfigJson())); err != nil {
		// Ошибка валидации/применения — текущий конфиг остался рабочим.
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pb.UpdateConfigResponse{Ok: true}, nil
}

func (s *nodeServer) GetConfig(_ context.Context, _ *pb.GetConfigRequest) (*pb.GetConfigResponse, error) {
	return &pb.GetConfigResponse{ConfigJson: string(s.mgr.Current())}, nil
}

func (s *nodeServer) GetStatus(_ context.Context, _ *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {
	st := s.mgr.Status()
	return &pb.GetStatusResponse{
		Running:        st.Running,
		SingboxVersion: st.SingboxVersion,
		UptimeSeconds:  st.UptimeSeconds,
		Inbounds:       int32(st.Inbounds),
		Outbounds:      int32(st.Outbounds),
		LastUpdateUnix: st.LastUpdateUnix,
	}, nil
}

// NewGRPCServer собирает gRPC-сервер с bearer-аутентификацией.
func NewGRPCServer(mgr *Manager, token string) *grpc.Server {
	srv := grpc.NewServer(grpc.UnaryInterceptor(bearerAuth(token)))
	pb.RegisterNodeServiceServer(srv, &nodeServer{mgr: mgr})
	return srv
}

// bearerAuth проверяет заголовок "authorization: Bearer <token>" на каждом unary-вызове.
func bearerAuth(token string) grpc.UnaryServerInterceptor {
	want := []byte(token)
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization")
		}
		const scheme = "Bearer "
		if !strings.HasPrefix(values[0], scheme) {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization scheme")
		}
		got := strings.TrimPrefix(values[0], scheme)
		if subtle.ConstantTimeCompare([]byte(got), want) != 1 {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		return handler(ctx, req)
	}
}
