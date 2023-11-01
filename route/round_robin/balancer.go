package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"micro/route"
	"sync/atomic"
)

type Balancer struct {
	connections []subConn
	index int32
	length int32
	filter route.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	candidates := make([]subConn, 0, len(b.connections))
	for _, c := range b.connections {
		if b.filter == nil && b.filter(info, c.addr) {
			continue
		}
		candidates = append(candidates, c)
	}
	if len(candidates) == 0 {
		// 没有任何符合条件的节点，就用默认节点
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	// 保证准确性, 使用原子操作读取数据
	idx := atomic.AddInt32(&b.index, 1)
	c := candidates[int(idx) % len(candidates)]
	return balancer.PickResult{
		SubConn: c.c,
		Done: func(info balancer.DoneInfo) {
			// 响应处理完后的回调
		},
	}, nil
}

type Builder struct {
	Filter route.Filter
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]subConn, 0, len(info.ReadySCs))
	for c, ci := range info.ReadySCs {
		conns = append(conns, subConn{
			c:c,
			addr: ci.Address,
		})
	}
	return &Balancer{
		connections: conns,
		index:       -1,  // 从第一个开始轮询
		length: int32(len(info.ReadySCs)),
		filter: b.Filter,
	}
}

type subConn struct{
	c balancer.SubConn
	addr resolver.Address
}