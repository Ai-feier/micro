package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"micro/proto/gen"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	us := &Server{}
	server := grpc.NewServer()
	gen.RegisterUserServiceServer(server, us)
	lis, err := net.Listen("tcp", ":8081")
	require.NoError(t, err)
	err = server.Serve(lis)
	t.Log()
}

type Server struct {
	gen.UnimplementedUserServiceServer
}

func (s *Server) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: "hello, world",
		},
	}, nil
}
