package ratelimit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync/atomic"
	"time"
)

type FixWindowLimiter struct {
	// 窗口的起始时间
	timestamp int64
	// 窗口大小
	interval int64
	// 在这个窗口内，允许通过的最大请求数量
	rate int64
	cnt int64
	//mutex sync.Mutex
	
	//onReject rejectStrategy
}

func NewFixWindowLimiter(interval time.Duration, rate int64) *FixWindowLimiter {
	return &FixWindowLimiter{
		timestamp: time.Now().UnixNano(),
		interval: interval.Nanoseconds(),
		rate: rate,
	}
}

func (f *FixWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		cur := time.Now().UnixNano()
		timestamp := atomic.LoadInt64(&f.timestamp)
		cnt := atomic.LoadInt64(&f.cnt)
		if timestamp + f.interval < cur {
			// 开新窗口
			// 用原子操作
			if atomic.CompareAndSwapInt64(&f.timestamp, timestamp, cur) {
				atomic.CompareAndSwapInt64(&f.cnt, cnt, 0)
			}
		}
		cnt = atomic.AddInt64(&f.cnt, 1)
		if cnt > f.rate {
			err = errors.New("触发瓶颈了")
			return
		}
		resp, err = handler(ctx, req)
		return 
	}
}

//func (f *FixWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
//	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
//		f.mutex.Lock()
//		cur := time.Now().UnixNano()
//		if f.timestamp + f.interval < cur {
//			f.timestamp = cur	
//			f.cnt = 0
//		}
//		
//		if f.cnt >= f.rate {
//			err = errors.New("触发瓶颈了")	
//			f.mutex.Unlock()
//			return 
//		}
//		f.cnt++
//		resp, err = handler(ctx, req)
//		f.mutex.Unlock()
//		return 
//	}
//}
