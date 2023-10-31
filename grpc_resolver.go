package micro

import (
	"context"
	"google.golang.org/grpc/resolver"
	"micro/registry"
	"time"
)

type grpcResolverBuilder struct {
	r registry.Registry
	timeout time.Duration
}

func NewRegistryBuilder(r registry.Registry, timeout time.Duration) (*grpcResolverBuilder, error) {
	return &grpcResolverBuilder{r: r, timeout: timeout}, nil
}

func (g *grpcResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn,
	opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &grpcResolver{
		cc: cc,
		target: target,
		timeout: g.timeout,
		r: g.r,
	}
	r.resolve()
	// 开启注册中心的事件监听
	go r.watch()
	return r, nil
}

func (g *grpcResolverBuilder) Scheme() string {
	return "registry"
}

type grpcResolver struct {
	// - "dns://some_authority/foo.bar"
	//   Target{Scheme: "dns", Authority: "some_authority", Endpoint: "foo.bar"}
	// registry:///localhost:8081
	target resolver.Target
	// 注册中心
	r registry.Registry
	// grpc 的服务连接抽象
	cc resolver.ClientConn
	// ResolverNow() 服务发现的过期时间
	timeout time.Duration
	close chan struct{}
}

func (g *grpcResolver) ResolveNow(options resolver.ResolveNowOptions) {
	g.resolve()
}

// 监听注册中心事件
func (g *grpcResolver) watch() {
	events, err := g.r.Subscribe(g.target.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
	}
	for {
		select {
		case <-events:
			// 直接全量从注册中心更新
			g.resolve()

			//case event := <- events:
			//switch event.Type {
			//case "DELETE":
			//	// 删除已有的节点
			//
			//} 
		case <-g.close:
			return
		}
	}
}

func (g *grpcResolver) Close() {
	close(g.close)
}

func (g *grpcResolver) resolve() {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	// 根据 endpoint 获取节点列表
	instanses, err := g.r.ListServices(ctx, g.target.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
		return
	}
	address := make([]resolver.Address, 0, len(instanses))
	for _, si := range instanses {
		address = append(address, resolver.Address{
			Addr: si.Address,
		})
	}

	// 从注册中心拿到全部节点列表后, 更新 grpc 连接层面上的 State
	err = g.cc.UpdateState(resolver.State{
		Addresses: address,
	})
	if err != nil {
		// grpc 服务连接抽象出错, 就报告
		g.cc.ReportError(err)
	}
}