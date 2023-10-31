package registry

import (
	"context"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"micro"
	"micro/proto/gen"
	"micro/registry/etcd"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	require.NoError(t, err)
	// 同样根据 etcd 连接拿到一个注册中心
	r, err := etcd.NewRegistry(etcdClient)
	require.NoError(t, err)
	
	// 根据注册中心新建一个自定义 rpc 客户端
	client := micro.NewClient(micro.ClientWithRegistry(r, time.Second*3), micro.ClientInsecure())
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// 拿到真正的 rpc 连接
	cc, err := client.Dial(ctx, "user-service")
	require.NoError(t, err)

	uc := gen.NewUserServiceClient(cc)
	resp, err := uc.GetById(ctx, &gen.GetByIdReq{Id: 123})
	require.NoError(t, err)
	t.Log(resp)
} 
