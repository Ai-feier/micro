package rpc

import (
	"encoding/binary"
	"net"
)

const (
	headLength = 8
)

func ReadMsg(conn net.Conn) ([]byte, error) {
	// 协议头和协议体
	lenBs := make([]byte, headLength)
	_, err := conn.Read(lenBs)
	if err != nil {
		return nil, err
	}
	headerLength := binary.BigEndian.Uint32(lenBs[:4])
	bodyLength := binary.BigEndian.Uint32(lenBs[4:])
	length := headerLength + bodyLength
	data := make([]byte, length)
	_, err = conn.Read(data[8:])
	copy(data[:8], lenBs)
	return data, err
}