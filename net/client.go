package net

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func Connect(network, addr string) error {
	conn, err := net.DialTimeout(network, addr, time.Second*3)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	for i := 0; i < 10; i++ {
		_, err := conn.Write([]byte("micro"))
		if err != nil {
			return err
		}
		res := make([]byte, 128)
		_, err = conn.Read(res)
		if err != nil {
			return err
		}
		fmt.Println(string(res))
	}
	return nil
}

type Client struct {
	network string
	addr string
}

func (c *Client) Send(data string) (string, error) {
	conn, err := net.DialTimeout(c.network, c.addr, time.Second*3)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = conn.Close()
	}()
	
	reqLen := len(data)
	req := make([]byte, headerLength+reqLen)
	binary.BigEndian.PutUint64(req[:headerLength], uint64(reqLen))
	copy(req[headerLength:], data)

	_, err = conn.Write(req)
	if err != nil {
		return "", err
	}
	
	respHeadBs := make([]byte, headerLength)
	_, err = conn.Read(respHeadBs)
	if err != nil {
		return "", err
	}

	// 从连接中读
	respLen := binary.BigEndian.Uint64(respHeadBs)
	
	respBs := make([]byte, respLen)
	_, err = conn.Read(respBs)
	if err != nil {
		return "", err
	}
	return string(respBs), nil
}