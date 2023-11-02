package ratelimit

import (
	"container/list"
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type SlideWindowLimiter struct {
	queue *list.List
	interval int64
	rate int
	mutex sync.Mutex
}

func NewSlideWindowLimiter(interval time.Duration, rate int) *SlideWindowLimiter {
	return &SlideWindowLimiter{
		queue: list.New(),
		interval: interval.Nanoseconds(),
		rate: rate,
	}
}

func (s *SlideWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		now := time.Now().UnixNano()
		boundary := now - s.interval
		
		// 快路径: 窗口计数没有超过时不进行删除
		s.mutex.Lock()  
		length := s.queue.Len()
		if length < s.rate {
			resp, err = handler(ctx, req)
			s.queue.PushBack(now)
			s.mutex.Unlock()
			return 
		}
		
		// 慢路径
		timestamp := s.queue.Front()
		for timestamp != nil && timestamp.Value.(int64) < boundary {
			s.queue.Remove(timestamp)
			timestamp = s.queue.Front()
		}
		length = s.queue.Len()
		if length >= s.rate {
			err = errors.New("触发瓶颈了")
			s.mutex.Unlock()
			return
		}
		resp, err = handler(ctx, req)
		// 记住了请求的时间戳
		s.queue.PushBack(now)
		s.mutex.Unlock()
		return
	}
}
