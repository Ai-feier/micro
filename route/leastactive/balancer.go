package leastactive

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"math"
	"micro/route"
	"sync/atomic"
)

type Balancer struct {
	connections []*activeConn
	filter route.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	res := &activeConn{
		cnt: math.MaxUint32,
	}
	for _, c := range b.connections {
		if b.filter != nil && !b.filter(info, c.addr) {
			continue
		}
		if atomic.LoadUint32(&c.cnt) <= res.cnt {
			res = c
		}
	}
	if res.cnt == math.MaxUint32 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
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
	Filter route.Filter
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*activeConn, 0, len(info.ReadySCs))
	for c, subInfo := range info.ReadySCs {
		conns = append(conns, &activeConn{
			c: c,
			addr: subInfo.Address,
		})
	}
	return &Balancer{
		connections: conns,
		filter: b.Filter,
	}
}

type activeConn struct {
	// 正在处理的请求数
	cnt uint32
	c balancer.SubConn
	addr resolver.Address
}


