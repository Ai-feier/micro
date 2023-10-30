package rpc

import (
	"context"
	"errors"
	"log"
	"micro/rpc/message"
	"micro/rpc/serialize"
	"micro/rpc/serialize/json"
	"net"
	"reflect"
	"strconv"
	"time"
)

type Server struct {
	services map[string]reflectionStub
	serializers map[uint8]serialize.Serializer
}

func NewServer() *Server {
	 res := &Server{
		services: make(map[string]reflectionStub, 8),
		serializers: make(map[uint8]serialize.Serializer, 4),
	}
	res.RegisterSerializer(&json.Serializer{})
	return res
}

func (s *Server) RegisterServer(service Service) {
	s.services[service.Name()] = reflectionStub{
		s: service,
		value: reflect.ValueOf(service),
		serializers: s.serializers,
	}
}

func (s *Server) RegisterSerializer(sl serialize.Serializer) {
	s.serializers[sl.Code()] = sl
}

func (s *Server) Start(network, addr string) error {
	lis, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}
		go func() {
			if er := s.handleConn(conn); er != nil {
				_ = conn.Close()
			}
		}()
	}
}

func (s *Server) handleConn(conn net.Conn) error {
	for {
		reqBs, err := ReadMsg(conn)
		if err != nil {
			return err
		}
		// 还原请求信息
		req := message.DecodeReq(reqBs)
		
		ctx := context.Background()
		cancel := func() {}
		log.Println(req.Meta)
		if deadlineStr, ok := req.Meta["deadline"]; ok {
			log.Println(deadlineStr)
			deadline, er := strconv.ParseInt(deadlineStr, 10, 64)
			if er == nil {
				ctx, cancel = context.WithDeadline(ctx, time.UnixMilli(deadline))
			}
		}
		oneway, ok := req.Meta["oneway"]
		if ok && oneway == "true" {
			ctx = CtxWithOneway(ctx)
		}

		resp, err := s.Invoke(ctx, req)
		// 服务链路调用结束, 结束 ctx
		cancel()
		if err != nil {
			// 处理业务 error
			resp.Error = []byte(err.Error())
		}
		
		// 设置好 response 
		resp.CalculateHeaderLength()
		resp.CalculateBodyLength()
		
		_, err = conn.Write(message.EncodeResp(resp))
		if err != nil {
			return err
		}
	}
}

func (s *Server) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	service, ok := s.services[req.ServiceName]
	// 服务端复用请求中的一定数据
	resp := &message.Response{
		RequestID: req.RequestID,
		Version: req.Version,
		Compresser: req.Compresser,
		Serializer: req.Serializer,
	}
	if !ok {
		return resp, errors.New("你要调用的服务不存在")
	}
	if isOneway(ctx) {
		go func() {
			_, _ = service.invoke(ctx, req)
		}()
		return resp, errors.New("micro: 微服务服务端 oneway 请求")
	}
	respData, err := service.invoke(ctx, req)
	resp.Data = respData
	if err != nil {
		return resp, err
	}
	return resp, nil
}

type reflectionStub struct {
	s Service
	value reflect.Value
	serializers map[uint8]serialize.Serializer
}

func (r *reflectionStub) invoke(ctx context.Context, req *message.Request) ([]byte, error) {
	// 反射找到方法，并且执行调用
	method := r.value.MethodByName(req.MethodName)
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(context.Background())
	inReq := reflect.New(method.Type().In(1).Elem())  // new 出来为指针
	serializer, ok := r.serializers[req.Serializer]
	if !ok {
		return nil, errors.New("micro: 不支持的序列化协议")
	}
	err := serializer.Decode(req.Data, inReq.Interface())
	if err != nil {
		return nil, err
	}
	in[1] = inReq
	// 方法输入构造完成, 方法调用
	result := method.Call(in)
	// res[0] 返回值
	// res[1] error
	if result[1].Interface() != nil {
		err = result[1].Interface().(error)
	}
	
	var res []byte
	if result[0].IsNil() {
		return nil, err
	} else {
		var er error
		res, er = serializer.Encode(result[0].Interface())
		if er != nil {
			return nil, er
		}
	}
	return res, err
}