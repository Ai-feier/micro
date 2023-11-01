package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"sync"
)

type WeightBalancer struct {
	connections []*weightConn
}

func (w *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(w.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var totalWeight uint32
	var res *weightConn
	//w.mutex.Lock()
	//defer w.mutex.Unlock()
	for _, c := range w.connections {
		// 减小锁力度的优化
		c.mutex.Lock()
		totalWeight += c.efficientWeight
		c.currentWeight += c.efficientWeight
		if res == nil || res.currentWeight < c.currentWeight {
			res = c
		}
		c.mutex.Unlock()
	}
	res.mutex.Lock()
	// 根据加权轮询算法更新权重
	res.currentWeight -= totalWeight
	res.mutex.Unlock()
	return balancer.PickResult{
		SubConn: res.c,
		Done: func(info balancer.DoneInfo) {
			res.mutex.Lock()
			if info.Err != nil && res.efficientWeight == 0 {
				return
			}
			if info.Err == nil && res.efficientWeight == math.MaxUint32 {
				return
			}
			if info.Err != nil {
				// 处理结果有错误, 减小其有效权重
				res.efficientWeight--
			} else {
				res.efficientWeight++
			}
			res.mutex.Unlock()
		},
	}, nil
}

type WeightBalancerBulider struct {
	
}

func (w *WeightBalancerBulider) Build(info base.PickerBuildInfo) balancer.Picker {
	cs := make([]*weightConn, 0, len(info.ReadySCs))
	for sub, subInfo := range info.ReadySCs {
		// 要拿到权重信息只有从 subInfo 中拿
		// 确保与 grpc 中 resolver 类型保持一致
		weight := subInfo.Address.Attributes.Value("weight").(uint32)
		
		cs = append(cs, &weightConn{
			c:               sub,
			weight:          weight,
			currentWeight:   weight,
			efficientWeight: weight,
		})
	}
	return &WeightBalancer{
		connections: cs,
	}
}

type weightConn struct {
	mutex           sync.Mutex
	c               balancer.SubConn
	weight          uint32  
	currentWeight   uint32 // 当前权重
	efficientWeight uint32 // 有效权重
}