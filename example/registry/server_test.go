package registry

import (
	"context"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"micro"
	"micro/proto/gen"
	"micro/registry/etcd"
	"testing"
)

func TestServer(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	// 用 etcd 连接创建注册中心
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)
	// 根据注册中心创建一个 rpc 服务端
	server, err := micro.NewServer("user-service", micro.ServerWithRegistry(r))
	require.NoError(t, err)
	us := &UserServiceServer{}
	// 把服务注册到 rpc 服务端
	gen.RegisterUserServiceServer(server, us)

	err = server.Start(":8081")
	t.Log(err)
}

type UserServiceServer struct {
	gen.UnimplementedUserServiceServer
}

func (s *UserServiceServer) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: "hello, world",
		},
	}, nil
}
