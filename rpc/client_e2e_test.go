package rpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestInitClientProxy(t *testing.T) {
	server := NewServer()
	server.RegisterServer(&UserServiceServer{})
	go func() {
		err := server.Start("tcp", ":8082")
		t.Log(err)
	}()
	time.Sleep(time.Second)
	usClient := &UserService{}
	err := InitClientProxy("localhost:8082", usClient)
	require.NoError(t, err)
	resp, err := usClient.GetById(context.Background(), &GetByIdReq{Id: 123})
	require.NoError(t, err)
	assert.Equal(t, &GetByIdResp{Msg: "hello-world"}, resp)
}

