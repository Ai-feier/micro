package ratelimit

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

type LeakyBucketLimiter struct {
	producer *time.Ticker
}

func NewLeakyBucketLimiter(interval time.Duration) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{
		producer: time.NewTicker(interval),
	}
}

func (l *LeakyBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return 
		case <-l.producer.C:
			resp, err = handler(ctx, req)
		}
		return 
	}
}

func (l *LeakyBucketLimiter) Close() error {
	l.producer.Stop()
	return nil
}