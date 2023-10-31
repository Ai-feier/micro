package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"micro/proto/gen"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	cc, err := grpc.Dial("registry:///localhost:8081", grpc.WithInsecure(), grpc.WithResolvers(&Builder{}))
	require.NoError(t, err)
	client := gen.NewUserServiceClient(cc)
	// 创建链路控制 ctx
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 3)
	defer cancel()
	resp, err := client.GetById(ctx, &gen.GetByIdReq{Id: 123})
	require.NoError(t, err)
	t.Log(resp)
}