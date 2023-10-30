package rpc

import (
	"context"
	"errors"
	"github.com/silenceper/pool"
	"micro/rpc/message"
	"micro/rpc/serialize"
	"micro/rpc/serialize/json"
	"net"
	"reflect"
	"time"
)

// InitService 要为 GetById 之类的函数类型的字段赋值
func (c *Client) InitService(service Service) error {
	return setFuncField(service, c, c.serializer)
}

func setFuncField(service Service, p Proxy, s serialize.Serializer) error {
	if service == nil {
		return errors.New("rpc: 不支持 nil")
	}
	val := reflect.ValueOf(service)
	typ := val.Type()
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return errors.New("rpc: 只支持指向结构体的一级指针")
	}
	
	val = val.Elem()
	typ = typ.Elem()
	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		fieldTyp := typ.Field(i)
		fieldVal := val.Field(i)
		
		if fieldVal.CanSet() {
			fn := func(args []reflect.Value) (results []reflect.Value) {
				// retVal 是个一级指针(reflect.New)
				retVal := reflect.New(fieldTyp.Type.Out(0).Elem())
				
				ctx := args[0].Interface().(context.Context)
				reqData, err := s.Encode(args[1].Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}
				meta := make(map[string]string)
				if isOneway(ctx) {
					meta["oneway"] = "true"
				}
				req := &message.Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Data:        reqData,
					Serializer: s.Code(),
					Meta: meta,
				}

				req.CalculateHeaderLength()
				req.CalculateBodyLength()
				
				// 发起调用
				resp, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}
				
				var retErr error
				if len(resp.Error) > 0 {
					retErr = errors.New(string(resp.Error))
				}
				
				if len(resp.Data) > 0 {
					err = s.Decode(resp.Data, retVal.Interface())
					if err != nil {
						return []reflect.Value{retVal, reflect.ValueOf(err)}
					}
				}
				
				var retErrVal reflect.Value
				if retErr == nil {
					retErrVal = reflect.Zero(reflect.TypeOf(new(error)).Elem())
				} else {
					retErrVal = reflect.ValueOf(retErr)
				}

				// reflect zero error 踩坑
				return []reflect.Value{retVal, retErrVal}
			}
			fnVal := reflect.MakeFunc(fieldTyp.Type, fn)
			fieldVal.Set(fnVal)
		}
	}
	return nil
}

type ClientOption func(client *Client)

type Client struct {
	pool pool.Pool
	serializer serialize.Serializer
}

func ClientWithSerializer(sl serialize.Serializer) ClientOption {
	return func(client *Client) {
		client.serializer = sl
	}
}

func NewClient(addr string, opts ...ClientOption) (*Client,error) {
	p, err := pool.NewChannelPool(&pool.Config{
		InitialCap:  1,
		MaxCap:      30,
		MaxIdle:     10,
		Factory: func() (interface{}, error) {
			return net.DialTimeout("tcp", addr, time.Second*3)
		},
		Close: func(i interface{}) error {
			return i.(net.Conn).Close()
		},
		IdleTimeout: time.Minute,
	})
	if err != nil {
		return nil, err
	}
	
	res := &Client{
		pool: p,
		serializer: &json.Serializer{},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func (c *Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	data := message.EncodeReq(req)
	// 发送给服务端
	resp, err := c.Send(data)
	if err != nil {
		return nil, err
	}
	return message.DecodeResp(resp), nil
}

func (c *Client) Send(data []byte) ([]byte, error) {
	val, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	conn := val.(net.Conn)
	defer func() {
		_ = c.pool.Put(conn)
	}()
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}
	return ReadMsg(conn)
}

