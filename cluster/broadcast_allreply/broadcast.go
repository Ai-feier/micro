package broadcast

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"micro/registry"
	"reflect"
	"sync"
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
		ok, ch := isBroadcast(ctx)
		if !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		// 从注册中心去捞服务端节点
		instanses, err := b.registry.ListServices(ctx, b.service)
		if err != nil {
			return err
		}
		var wg sync.WaitGroup
		typ := reflect.TypeOf(reply).Elem()
		wg.Add(len(instanses))
		for _, ins := range instanses {
			// 再这里可以加入 filter, 过滤节点, 达成组播效果
			addr := ins.Address
			go func() {
				insCC, er := grpc.Dial(addr, b.dialOptions...)
				if er != nil {
					ch <- Resp{Err: er}
					wg.Done()
					return
				}
				newReply := reflect.New(typ)
				er = invoker(ctx, method, req, reply, insCC, opts...)	
				// 必须有人接收, 否则阻塞
				select {
				case <-ctx.Done():
					err = fmt.Errorf("响应没有人接收, %w", ctx.Err())
				case ch <- Resp{Err: er, Reply: newReply}:
				}
				wg.Done()
			}()
		}
		wg.Wait()
		return err
	}
}

func UseBroadcast(ctx context.Context) (context.Context, <-chan Resp) {
	ch := make(chan Resp)
	return context.WithValue(ctx, broadcastKey{}, ch), ch
}

type broadcastKey struct{}

func isBroadcast(ctx context.Context) (bool, chan Resp){
	val, ok := ctx.Value(broadcastKey{}).(chan Resp)
	return ok, val  // 存在且正确
}

type Resp struct {
	Err error
	Reply any
}