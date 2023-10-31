package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"micro/registry"
	"sync"
)

type Registry struct {
	c *clientv3.Client
	sess *concurrency.Session
	cancels []func()
	mutex sync.Mutex
}

// 从配置中区加载
//func NewRegistryV1(cfg []byte) *Registry {
//	client := clientv3.New(cfg)
//}

// 让用户把租约的 session 一并创建好
//func NewRegistry(sess *concurrency.Session) *Registry {
//
//}

func NewRegistry(c *clientv3.Client) (*Registry, error) {
	// 根据客户端创建租约 session
	sess, err := concurrency.NewSession(c)
	if err != nil {
		return nil, err
	}
	return &Registry{
		c: c,
		sess: sess,
	}, nil
}

func (r *Registry) Register(ctx context.Context, si registry.ServiceInstance) error {
	// 把节点信息以 json 格式写入
	val, err := json.Marshal(si)
	if err != nil {
		return err
	}
	
	// 把节点信息写入 etcd key:节点实例标识, val:json后的节点实例, 以及新建一个租约
	// 第一个返回不好解析, 用处不大
	_, err = r.c.Put(ctx, r.instanceKey(si), string(val), clientv3.WithLease(r.sess.Lease()))
	return err
}

func (r *Registry) UnRegister(ctx context.Context, si registry.ServiceInstance) error {
	_, err := r.c.Delete(ctx, r.instanceKey(si))
	return err
}

func (r *Registry) ListServices(ctx context.Context, serviceName string) ([]registry.ServiceInstance, error) {
	// 根据服务名拿到全部节点
	getResp, err := r.c.Get(ctx, r.serviceKey(serviceName), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	res := make([]registry.ServiceInstance, 0, len(getResp.Kvs))
	for _, kv := range getResp.Kvs {
		var si registry.ServiceInstance
		err = json.Unmarshal(kv.Value, &si)
		if err != nil {
			return nil, err
		}
		res = append(res, si)
	}
	return res, nil
}

func (r *Registry) Subscribe(serviceName string) (<-chan registry.Event, error) {
	ctx, cancel := context.WithCancel(context.Background())
	r.mutex.Lock()
	r.cancels = append(r.cancels, cancel)
	r.mutex.Unlock()
	
	// 只要有 Leader 时的事件, 避免主从切换的误差
	ctx = clientv3.WithRequireLeader(ctx)
	watchResp := r.c.Watch(ctx, r.serviceKey(serviceName), clientv3.WithPrefix())
	
	res := make(chan registry.Event)
	go func() {
		for {
			select {
			case resp := <-watchResp:
				if resp.Err() != nil {
					//return
					continue
				}
				if resp.Canceled {
					return
				}
				// 返回为一批事件
				for range resp.Events {
					// 只需要通知一下, 注册中心就会全量从 etcd... 更新节点信息
					res <- registry.Event{}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return res, nil
}

func (r *Registry) Close() error {
	r.mutex.Lock()
	cancels := r.cancels
	r.cancels = nil
	r.mutex.Unlock()
	for _, c := range cancels {
		c()
	}
	// 关闭 etcd 的 session 就会关闭其租约
	return r.sess.Close()
}

func (r *Registry) instanceKey(si registry.ServiceInstance) string {
	// 可以考虑直接用 Address
	// 也可以说在 si 里面引入一个 InstanceName 的字段
	return fmt.Sprintf("/micro/%s/%s", si.Name, si.Address)
}

func (r *Registry) serviceKey(sn string) string {
	return fmt.Sprintf("/micro/%s", sn)
}

