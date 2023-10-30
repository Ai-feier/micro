package rpc

import (
	"encoding/binary"
	"net"
)

const (
	headerLength = 8
)

func ReadMsg(conn net.Conn) ([]byte, error) {
	lenBs := make([]byte, headerLength)
	_, err := conn.Read(lenBs)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint64(lenBs)
	data := make([]byte, length)
	_, err = conn.Read(data)
	return data, err
}

func EncodeMsg(data []byte) []byte {
	reqLen := len(data)
	res := make([]byte, reqLen+headerLength)
	binary.BigEndian.PutUint64(res[:headerLength], uint64(reqLen))
	copy(res[headerLength:], data)
	return res
}