package registry

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
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

	var eg errgroup.Group
	for i := 0; i < 3; i++ {
		var group = "A"
		if i % 2 == 1 {
			group = "B"
			// 可通过设置 group 字段, 将请求发到压力测试服务器
		}
		server, err := micro.NewServer("user-service", 
			micro.ServerWithRegistry(r), micro.ServerWithGroup(group))
		require.NoError(t, err)
		
		us := &UserServiceServer{group: group}
		gen.RegisterUserServiceServer(server, us)
		
		port := fmt.Sprintf(":808%d", i+1)
		eg.Go(func() error {
			return server.Start(port)
		})
	}
	err = eg.Wait()
	t.Log(err)
}

type UserServiceServer struct {
	group string
	gen.UnimplementedUserServiceServer
}

func (s *UserServiceServer) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(s.group)
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: "hello, world",
		},
	}, nil
}
