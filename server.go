package micro

import (
	"context"
	"google.golang.org/grpc"
	"micro/registry"
	"net"
	"time"
)

type ServerOption func(server *Server)

// Server 标识信息服务名
// rpc 连接的服务端
// 兼容自定注册中心
type Server struct {
	name string
	*grpc.Server
	registry registry.Registry
	registerTimeout time.Duration
	listener net.Listener
	weight uint32
	group string
}

func NewServer(name string, opts...ServerOption) (*Server, error) {
	res := &Server{
		name: name,
		Server: grpc.NewServer(),
		registerTimeout: 10 * time.Second, // 初始固定注册超时时间
	}
	
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}


func ServerWithRegistry(r registry.Registry) ServerOption {
	return func(server *Server) {
		server.registry = r
	}
}

func (s *Server) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = lis
	
	// 当前服务是否有注册中心
	if s.registry != nil {
		// 在这里注册
		ctx, cancel := context.WithTimeout(context.Background(), s.registerTimeout)
		defer cancel()
		err = s.registry.Register(ctx, registry.ServiceInstance{
			Name:    s.name,
			// 节点的唯一定位信息
			Address: s.listener.Addr().String(),
		})
		if err != nil {
			return err
		}
		// 这里已经注册成功了
		// 最佳实践: 实现 Close(), 同一关闭资源
		//defer func() {
		// 忽略或者 log 一下错误
		//_ = s.registry.Close()
		//_ = s.registry.UnRegister(registry.ServiceInstance{})
		//}()
	}
	
	// 启动 rpc 监听
	err = s.Serve(s.listener)
	return err
}

func (s *Server) Close() error {
	if s.registry != nil {
		err := s.registry.Close()
		if err != nil {
			return err
		}
	}
	// 直接 grpc 的优雅退出方法, 关闭链接
	s.GracefulStop()
	return nil
}