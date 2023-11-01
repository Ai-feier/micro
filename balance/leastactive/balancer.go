package leastactive

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"sync/atomic"
)

type Balancer struct {
	connections []*activeConn
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	res := &activeConn{
		cnt: math.MaxUint32,
	}
	for _, c := range b.connections {
		if atomic.LoadUint32(&c.cnt) <= res.cnt {
			res = c
		}
	}
	atomic.AddUint32(&res.cnt, 1)
	return balancer.PickResult{
		SubConn: res.c,
		Done: func(info balancer.DoneInfo) {
			atomic.AddUint32(&res.cnt, -1)
		},
	}, nil
}

type Builder struct {
	
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*activeConn, 0, len(info.ReadySCs))
	for c := range info.ReadySCs {
		conns = append(conns, &activeConn{
			c: c,
		})
	}
	return &Balancer{
		connections: conns,
	}
}

type activeConn struct {
	// 正在处理的请求数
	cnt uint32
	c balancer.SubConn
}


