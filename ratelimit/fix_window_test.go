package ratelimit

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"micro/proto/gen"
	"testing"
	"time"
)

func TestFixWindowLimiter_BuildServerInterceptor(t *testing.T) {
	// 测试时序, 窗口更新
	interceptor := NewFixWindowLimiter(3*time.Second, 1).BuildServerInterceptor()
	cnt := 0
	handler := func(ctx context.Context, req any) (any, error) {
		cnt++
		return &gen.GetByIdResp{}, nil
	}
	resp, err := interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, &gen.GetByIdResp{}, resp)

	resp, err = interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.Equal(t,  errors.New("触发瓶颈了"), err)
	assert.Nil(t, resp)

	// 睡一个三秒，确保窗口新建了
	time.Sleep(time.Second * 3)
	resp, err = interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, &gen.GetByIdResp{}, resp)
}
