package main

import (
	"context"
	"net"
	"testing"

	pb "github.com/CascadiaLabs/node/proto/node"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const testToken = "s3cret-token"

// newTestClient поднимает gRPC-сервер на bufconn (без реальной сети) с auth-токеном testToken.
func newTestClient(t *testing.T) pb.NodeServiceClient {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	srv := NewGRPCServer(NewManager(""), testToken)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return pb.NewNodeServiceClient(conn)
}

func withToken(token string) context.Context {
	return metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+token)
}

func TestAuthMissingToken(t *testing.T) {
	c := newTestClient(t)
	_, err := c.GetStatus(context.Background(), &pb.GetStatusRequest{})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("want Unauthenticated, got %v", err)
	}
}

func TestAuthWrongToken(t *testing.T) {
	c := newTestClient(t)
	_, err := c.GetStatus(withToken("wrong"), &pb.GetStatusRequest{})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("want Unauthenticated, got %v", err)
	}
}

func TestAuthValidToken(t *testing.T) {
	c := newTestClient(t)
	resp, err := c.GetStatus(withToken(testToken), &pb.GetStatusRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetRunning() {
		t.Fatal("fresh manager: want running=false")
	}
}
