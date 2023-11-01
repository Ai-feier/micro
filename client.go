package micro

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"micro/registry"
	"time"
)

type ClientOption func(c *Client)

type Client struct {
	insecure bool
	r registry.Registry
	timeout time.Duration
	// 负载均衡的 pirckerbuilder
	balancer balancer.Builder
}

// NewClient 可以不使用注册中心
func NewClient(opts ...ClientOption) *Client {
	res := &Client{}
	
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func ClientWithPickBuilder(name string, b base.PickerBuilder) ClientOption {
	return func(c *Client) {
		builder := base.NewBalancerBuilder(name, b, base.Config{HealthCheck: true})
		balancer.Register(builder)
		c.balancer = builder
	}
}

func ClientInsecure() ClientOption {
	return func(c *Client) {
		c.insecure = true
	}
}

func ClientWithRegistry(r registry.Registry, timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.r = r
		c.timeout = timeout
	}
}

func (c *Client) Dial(ctx context.Context, service string, 
	dialOptions...grpc.DialOption) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	// 如果有注册中心, 构造 grpc 服务发现的 option
	if c.r != nil {
		// 拿到自定义的 resolverBuiler 
		rb, err := NewRegistryBuilder(c.r, c.timeout)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithResolvers(rb))
	}
	if c.insecure {
		opts = append(opts, grpc.WithInsecure())
	}
	// 增加负载均衡的 grpc option
	if c.balancer != nil {
		opts = append(opts, grpc.WithDefaultServiceConfig(
			fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, c.balancer.Name())))
	}
	if len(dialOptions) > 0 {
		opts = append(opts, dialOptions...)
	}
	cc, err := grpc.DialContext(ctx, fmt.Sprintf("registry:///%s", service), opts...)
	return cc, err
}