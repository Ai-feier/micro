package net

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

type Pool struct {
	// 空闲连接队列
	idlesConns chan *idleConn
	// 请求队列
	reqQueue []connReq

	// 最大连接数
	maxCnt int

	// 当前连接数——你已经建好了的
	cnt int

	// 最大空闲时间
	maxIdleTime time.Duration

	factory func() (net.Conn, error)

	lock sync.Mutex
}

func NewPool(initCnt int, maxIdleCnt int, maxCnt int,
	maxIdleTime time.Duration,
	factory func() (net.Conn, error)) (*Pool,error) {

	if initCnt > maxIdleCnt {
		return nil, errors.New("micro: 初始连接数量不能大于最大空闲连接数量")
	}
	idlesConn := make(chan *idleConn, maxIdleCnt)
	for i := 0; i < initCnt; i++ {
		conn, err := factory()
		if err != nil {
			return nil, err
		}
		idlesConn <- &idleConn{
			c:              conn,
			lastActiveTime: time.Now(),
		}
	}
	return &Pool{
		idlesConns:  idlesConn,
		maxCnt:      maxCnt,
		cnt:         0,
		maxIdleTime: maxIdleTime,
		factory:     factory,
	}, nil
}

func (p *Pool) Get(ctx context.Context) (net.Conn, error) {
	select {
	case <-ctx.Done():
	default:
	}
	for {
		select {
		case ic := <-p.idlesConns:
			if ic.lastActiveTime.Before(time.Now()) {
				_ = ic.c.Close()
				continue
			}
			return ic.c, nil
		default:
			// 没有空闲连接
			p.lock.Lock()
			if p.cnt >= p.maxCnt {
				// 超出上限, 把自己加入等待队列
				req := connReq{connChan: make(chan net.Conn, 1)}
				p.reqQueue = append(p.reqQueue, req)
				p.lock.Unlock()
				select {
				case <-ctx.Done():
					// 超时, 异步开启创建连接, 在放回连接池
					go func() {
						c := <-req.connChan
						_ = p.Put(context.Background(), c)
					}()
					return nil, ctx.Err()
					case c := <- req.connChan:
						return c, nil
				}
			}
			// 没有超出上限自己创建一个
			c, err := p.factory()
			if err != nil {
				return nil, err
			}
			p.cnt++
			return c, nil
		}
	}
}

func (p *Pool) Put(ctx context.Context, conn net.Conn) error {
	p.lock.Unlock()
	if len(p.reqQueue) > 0 {
		req := p.reqQueue[0]
		p.reqQueue = p.reqQueue[1:]
		p.lock.Unlock()
		req.connChan <- conn
		return nil
	}
	defer p.lock.Unlock()
	ic := &idleConn{
		c:              conn,
		lastActiveTime: time.Now(),
	}
	select {
	case p.idlesConns <- ic:
	default:
		// 空闲连接队列已满, 销毁队列
		_ = conn.Close()
		p.cnt--
	}
	return nil
}

type idleConn struct{
	c net.Conn
	lastActiveTime time.Time
}

type connReq struct{
	connChan chan net.Conn
}
