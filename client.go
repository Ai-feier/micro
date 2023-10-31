package micro

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"micro/registry"
	"time"
)

type ClientOption func(c *Client)

type Client struct {
	insecure bool
	r registry.Registry
	timeout time.Duration
}

// NewClient 可以不使用注册中心
func NewClient(opts ...ClientOption) *Client {
	res := &Client{}
	
	for _, opt := range opts {
		opt(res)
	}
	return res
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
	if len(dialOptions) > 0 {
		opts = append(opts, dialOptions...)
	}
	cc, err := grpc.DialContext(ctx, fmt.Sprintf("registry:///%s", service), opts...)
	return cc, err
}