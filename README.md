# go 微服务素材
## 1. 简单 rpc 框架
#### 实现功能
- 采用代理方式实现`rpc`调用, 通过注册服务类, 通过反射为服务类构造具体实现方法 (参考`Dubbo-go`实现)
- 设计自定义`rpc`协议
- 支持多种序列化协议, 提供抽象接口
- 引入`oneway`单向调用实现
- 设计根据时间戳的`rpc`链路控制
#### 优化
- 对于`rpc`客户端采用连接, 最大化利用连接