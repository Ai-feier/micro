package broadcast

import (
	"context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"micro/registry"
	"reflect"
)

// grpc 不能通过 picker 实现广播与组播
// 可通过 interceptor(aop) 实现
// 实现 grpc 中的 UnaryClientInterceptor 方法签名
// 使用通过注册中获取所有服务端节点, 曲线救国

type ClusterBuilder struct {
	registry registry.Registry
	service string
	dialOptions []grpc.DialOption
}

func NewClusterBuilder(r registry.Registry, service string, dialOptions...grpc.DialOption) *ClusterBuilder {
	return &ClusterBuilder{
		registry: r,
		service: service,
		dialOptions: dialOptions,
	}
}

func (b ClusterBuilder) BuildUnaryInterceptor() grpc.UnaryClientInterceptor {
	// method: users.UserService/GetByID
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, 
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !isBroadcast(ctx) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		// 从注册中心去捞服务端节点
		instanses, err := b.registry.ListServices(ctx, b.service)
		if err != nil {
			return err
		}
		var eg errgroup.Group
		typ := reflect.TypeOf(reply).Elem()
		for _, ins := range instanses {
			// 再这里可以加入 filter, 过滤节点, 达成组播效果
			addr := ins.Address
			eg.Go(func() error {
				// 这里每次调用都会新建 tcp 连接
				insCC, er := grpc.Dial(addr, b.dialOptions...)
				if er != nil {
					return er
				}
				// // 这里的实现存在 reply 复用, 不能处理每个响应, 而是最慢的那个
				ry := reflect.New(typ)
				_ = invoker(ctx, method, req, ry, insCC, opts...)
				return nil
			})
		}
		return eg.Wait()
	}
}

func UseBroadcast(ctx context.Context) context.Context {
	return context.WithValue(ctx, broadcastKey{}, true)
}

type broadcastKey struct{}

func isBroadcast(ctx context.Context) bool {
	val, ok := ctx.Value(broadcastKey{}).(bool)
	return ok && val  // 存在且正确
}