package grpc

import (
	"google.golang.org/grpc/resolver"
)

type Builder struct {
	
}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, 
	opts resolver.BuildOptions) (resolver.Resolver, error) {
	// cc resolver.ClientConn 是一个服务的抽象, 不是一个连接
	// cc 要向下传递给 resolver, 之前要 ResolverNow()
	r := &Resolver{
		cc: cc,
	}
	r.ResolveNow(resolver.ResolveNowOptions{}) // 传一个空 Option
	return r, nil
}

func (b *Builder) Scheme() string {
	// "registry:///localhost:8081" 
	// scheme 即为 registry
	return "registry"
}

type Resolver struct {
	cc resolver.ClientConn
}

func (r *Resolver) ResolveNow(options resolver.ResolveNowOptions) {
	err := r.cc.UpdateState(resolver.State{
		Addresses: []resolver.Address{
			{
				Addr: "localhost:8081",
			},
		},
	})

	if err != nil {
		r.cc.ReportError(err)  // cc.ReportError() 会默认调一次 ResolverNow() 
	}
}

func (r *Resolver) Close() {
	
}

