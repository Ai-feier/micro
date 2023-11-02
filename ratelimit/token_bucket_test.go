package ratelimit

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"micro/proto/gen"
	"testing"
)

func TestTokenBucketLimiter_BuildServerInterceptor(t *testing.T) {
	testCases := []struct{
		name string
		
		b func() *TokenBucketLimiter
		ctx context.Context
		handler func(ctx context.Context, req interface{}) (interface{}, error)
		
		wantErr error
		wantResp interface{}
	}{
		{
			name: "closed",
			b: func() *TokenBucketLimiter {
				closed := make(chan struct{})
				close(closed)
				return &TokenBucketLimiter{
					tokens: make(chan struct{}),
					close: closed,
				}
			},
			ctx: context.Background(),
			wantErr: errors.New("缺乏保护，拒绝请求"),
		},
		{
			name: "context canceled",
			b: func() *TokenBucketLimiter {
				return &TokenBucketLimiter{
					tokens: make(chan struct{}),
					close: make(chan struct{}),
				}
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
		{
			name: "get tokens",
			b: func() *TokenBucketLimiter {
				ch := make(chan struct{}, 1)
				ch <- struct{}{}
				return &TokenBucketLimiter{
					tokens: ch,
					close: make(chan struct{}),
				}
			},
			ctx: context.Background(),
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return &gen.GetByIdResp{}, nil 
			},
			wantResp: &gen.GetByIdResp{},
		},
		
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interceptor := tc.b().BuildServerInterceptor()
			resp, err := interceptor(tc.ctx, &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, tc.handler)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}
