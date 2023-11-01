package hash

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type Balancer struct {
	connections []balancer.SubConn
	length int
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if b.length == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	// grpc 没有为负载均衡提供很好的设计, 较难获取对应的数据做哈希负载均衡
	// 哈希负载均衡: 哈希相同的请求会落到同一台服务器, 可以有效利用缓存, 但都是大请求服务器负载不均匀
	// 在这个地方你拿不到请求，无法做根据请求特性做负载均衡
	//idx := info.Ctx.Value("user_id")
	//idx := info.Ctx.Value("hash_code")

	return balancer.PickResult{
		SubConn: b.connections[0],
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type Builder struct {
	
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for c := range info.ReadySCs {
		connections = append(connections, c)
	}
	return &Balancer{
		connections: connections,
		length: len(connections),
	}
}

