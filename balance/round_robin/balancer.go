package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync/atomic"
)

type Balancer struct {
	connections []balancer.SubConn
	index int32
	length int32
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	// 保证准确性, 使用原子操作读取数据
	idx := atomic.AddInt32(&b.index, 1)
	c := b.connections[idx]
	return balancer.PickResult{
		SubConn: c,
		Done: func(info balancer.DoneInfo) {
			// 响应处理完后的回调
		},
	}, nil
}

type Builder struct {
	
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for c := range info.ReadySCs {
		conns = append(conns, c)
	}
	return &Balancer{
		connections: conns,
		index:       -1,  // 从第一个开始轮询
		length: int32(len(info.ReadySCs)),
	}
}

