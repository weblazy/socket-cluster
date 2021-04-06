# socket-cluster
即时通讯框架,如果大家觉得好用,右上角帮忙点个star吧。(^_^)
> 欢迎感兴趣的小伙伴一同开发,解决日常消息推送,长连接需求。
- 后端(golang): https://github.com/weblazy/socket-cluster
- 使用案例:	
    - 在线体验: http://web.xiaoyuantongbbs.cn:3333/login.html#/
 	- 后端源码: https://github.com/weblazy/socket-cluster-examples
	- 前端源码: https://github.com/weblazy/socket-cluster-web
- 使用文档: http://doc.xiaoyuantongbbs.cn:4000/
- 后端好用的工具库(golang): https://github.com/weblazy/easy
# 框架库依赖
- redis

# 框架使用方式
```
	// 客户端协议组件
	protocolHandler := &ws_protocol.WsProtocol{}
	// 在线状态存储组件
	sessionStorageHandler := redis_storage.NewRedisStorage([]*redis_storage.RedisNode{&redis_storage.RedisNode{
		RedisConf: &redis.Options{Addr: redisHost, Password: redisPassword, DB: 0},
		Position:  10000,
	}})
    // 服务发现组件
	discoveryHandler := redis_discovery.NewRedisDiscovery(&redis.Options{Addr: redisHost, Password: redisPassword, DB: 0})
	// 启动服务
	common.NodeInfo, err = node.NewNode(node.NewNodeConf(*host, protocolHandler, sessionStorageHandler, discoveryHandler, onMsg).WithPort(*port))
	if err != nil {
		logx.Info(err)
	}
```
```
func onMsg(context *node.Context) {
	logx.Info("msg:", string(context.Msg))
}
```

# 架构框图
![scheme 1](pic/websocket.png)

# 联系我们
- 技术支持/合作/咨询请联系作者QQ: 2276282419
- 作者邮箱: 2276282419@qq.com
- 即时通讯技术交流QQ群: 33280853
