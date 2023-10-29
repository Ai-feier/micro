package net

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

// 长度字段使用的字节数量
const headerLength = 8

func Serve(network, addr string) error {
	lis, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	
	for {
		conn, err := lis.Accept()
		if err != nil {
			// 连接出现问题
			return err
		}
		go func() {
			if er := handleConn(conn); er != nil {
				_ = conn.Close()
			}
		}()
	}
}

func handleConn(conn net.Conn) error {
	// 循环处理连接
	for {
		bs := make([]byte, 8)
		_, err := conn.Read(bs)
		// 连接不可用
		if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) || 
			errors.Is(err, io.ErrUnexpectedEOF) {
			return err
		}
		if err != nil {
			continue
		}
		// 处理数据
		res := handleMsg(bs)
		
		// 回写响应
		n, err := conn.Write(res)
		if err != nil {
			return err
		}
		if n != len(res) {
			return errors.New("micro: 没写完数据")
		}
	}
}

func handleConnV1(conn net.Conn) error {
	for {
		bs := make([]byte, 8)
		n, err := conn.Read(bs)
		if err != nil {
			return err
		}
		if n != 8 {
			return errors.New("micro: 没读够数据")
		}
		res := handleMsg(bs)
		n, err = conn.Write(res)
		if err != nil {
			return err
		}
		if n != len(res) {
			return errors.New("micro: 没写完数据")
		}
	}
}

func handleMsg(req []byte) []byte {
	res := make([]byte, len(req) * 2)
	copy(res[:len(req)], req)
	copy(res[len(req):], req)
	return res
}

type Server struct {
	//network string
	//addr string
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

// 我们可以认为，一个请求包含两部分
// 1. 长度字段：用八个字节表示
// 2. 请求数据：
// 响应也是这个规范
func (s *Server) handleConn(conn net.Conn) error {
	for {
		// 先从连接中读取长度字段
		lenBs := make([]byte, headerLength)
		_, err := conn.Read(lenBs)
		if err != nil {
			return err
		}
		
		// 消息长度
		length := binary.BigEndian.Uint64(lenBs)
		
		reqBs := make([]byte, length)
		_, err = conn.Read(reqBs)
		if err != nil {
			return err
		}
		
		respData := handleMsg(reqBs)
		respLen := len(respData)
		
		// 构造响应头+体
		resp := make([]byte, respLen+headerLength)
		binary.BigEndian.PutUint64(resp[:headerLength], uint64(respLen))
		copy(resp[headerLength:], respData)

		_, err = conn.Write(resp)
		if err != nil {
			return err
		}
	}
}
