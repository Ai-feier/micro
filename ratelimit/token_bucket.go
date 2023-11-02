package ratelimit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"time"
)

type TokenBucketLimiter struct {
	tokens chan struct{}
	close  chan struct{}
}

// NewTokenBucketLimiter interval 隔多久产生一个令牌
func NewTokenBucketLimiter(capacity int, interval time.Duration) *TokenBucketLimiter {
	ch := make(chan struct{}, capacity)
	closeCh := make(chan struct{})
	producer := time.NewTicker(interval)
	go func() {
		defer func() {
			producer.Stop()
		}()
		for {
			select {
			case <-closeCh:
				return
			case <-producer.C:
				// 令牌满了避免阻塞
				select {
				case ch <- struct{}{}:
				default:
				}
			}
		}
	}()
	return &TokenBucketLimiter{
		tokens: ch,
		close: closeCh,
	}
}

func (t *TokenBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-t.close:
			// 你已经关掉故障检测了
			//resp, err = handler(ctx, req)
			err = errors.New("缺乏保护，拒绝请求")
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <-t.tokens:
			resp, err = handler(ctx, req)
		}
		return 
	}
}

func (t *TokenBucketLimiter) Close() error {
	close(t.close)
	return nil
}