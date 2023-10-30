package rpc

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"micro/proto/gen"
	"micro/rpc/serialize/proto"
	"testing"
	"time"
)

func TestInitServiceProto(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterServer(service)
	server.RegisterSerializer(&proto.Serializer{})
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)

	usClient := &UserService{}
	client, err := NewClient(":8081", ClientWithSerializer(&proto.Serializer{}))
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct{
		name string
		mock func()

		wantErr error
		wantResp *GetByIdResp
	} {
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello, world"
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
		},
		{
			name: "error",
			mock: func() {
				service.Msg = ""
				service.Err = errors.New("mock error")
			},
			wantResp: &GetByIdResp{},
			wantErr: errors.New("mock error"),
		},

		{
			name: "both",
			mock: func() {
				service.Msg = "hello, world"
				service.Err = errors.New("mock error")
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
			wantErr: errors.New("mock error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, er := usClient.GetByIdProto(context.Background(), &gen.GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			if resp != nil && resp.User != nil {
				assert.Equal(t, tc.wantResp.Msg, resp.User.Name)
			}
		})
	}
}

func TestInitClientProxy(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterServer(service)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second)
	
	usClient := &UserService{}
	client, err := NewClient(":8081")
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)

	testCases := []struct{
		name string
		mock func()

		wantErr error
		wantResp *GetByIdResp
	} {
		{
			name: "no error",
			mock: func() {
				service.Err = nil
				service.Msg = "hello, world"
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
		},
		{
			name: "error",
			mock: func() {
				service.Msg = ""
				service.Err = errors.New("mock error")
			},
			wantResp: &GetByIdResp{},
			wantErr: errors.New("mock error"),
		},

		{
			name: "both",
			mock: func() {
				service.Msg = "hello, world"
				service.Err = errors.New("mock error")
			},
			wantResp: &GetByIdResp{
				Msg: "hello, world",
			},
			wantErr: errors.New("mock error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			resp, er := usClient.GetById(context.Background(), &GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestOneway(t *testing.T) {
	server := NewServer()
	service := &UserServiceServer{}
	server.RegisterServer(service)
	go func() {
		err := server.Start("tcp", ":8081")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)

	usClient := &UserService{}
	client, err := NewClient(":8081")
	require.NoError(t, err)
	err = client.InitService(usClient)
	require.NoError(t, err)
	testCases := []struct{
		name string
		mock func()

		wantErr error
		wantResp *GetByIdResp
	} {
		{
			name: "oneway",
			mock: func() {
				service.Err = errors.New("mock error")
				service.Msg = "hello, world"
			},
			wantResp: &GetByIdResp{},
			wantErr: errors.New("micro: 微服务服务端 oneway 请求"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mock()
			ctx := CtxWithOneway(context.Background())
			resp, er := usClient.GetById(ctx, &GetByIdReq{Id: 123})
			assert.Equal(t, tc.wantErr, er)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}