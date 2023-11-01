package round_robin

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"micro/proto/gen"
	"net"
	"testing"
	"time"
)

func TestBalancer_e2e_Pick(t *testing.T) {
	// 启动服务端
	go func() {
		us := &Server{}
		server := grpc.NewServer()
		gen.RegisterUserServiceServer(server, us)
		lis, err := net.Listen("tcp", ":8081")
		require.NoError(t, err)
		err = server.Serve(lis)
		t.Log(err)
	}()
	
	time.Sleep(time.Second * 3)
	
	// 启动客服端
	// 注册负载均衡
	balancer.Register(base.NewBalancerBuilder("ROUND_ROBIN",
		&Builder{},
		base.Config{HealthCheck: true}, // 健康检测
		))
	
	// 通过 grpc 选项注册负载均衡的客户端
	cc, err := grpc.Dial("localhost:8081", grpc.WithInsecure(), 
		grpc.WithDefaultServiceConfig(`{"LoadBalancingPolicy": "DEMO_ROUND_ROBIN"}`)) // 负载均衡名是一个 json 串
	require.NoError(t, err)

	client := gen.NewUserServiceClient(cc)
	
	// 可以调用 rpc 服务
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, err := client.GetById(ctx, &gen.GetByIdReq{Id: 123})
	require.NoError(t, err)
	t.Log(resp)
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
