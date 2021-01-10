# websocket-cluster
即时通讯框架
> 欢迎感兴趣的小伙伴一同开发,解决日常消息推送,长连接需求。
- 后端: websocket-cluster(golang)
- 前端: websocket-cluster-web
- 文档: websocket-cluster-doc
- 后端好用的工具库: https://github.com/weblazy/easy(golang)
# 框架库依赖
- redis

# demo快速开始
- 依赖服务
    - redis
    - mysql
- 环境变量配置
```
#mysql
export DBDEFAULT_DBHOST="127.0.0.1"
export DBDEFAULT_DBNAME="websocket_cluster"
export DBDEFAULT_DBUSER="root"
export DBDEFAULT_DBPASSWD=""

#redis
export REDIS_HOST="127.0.0.1:6379"
export REDIS_PASSWORD=""

#config
export RUN_TIME="local"
export EMAIL_PASSWORD=""
```
- 启动命令
```
cd websocket-cluster/examples/main/
./main
```
# 框架使用方式
```
	common.NodeINfo1, err = websocket_cluster.StartNode(websocket_cluster.NewNodeConf(*host1, *path1, *path1, websocket_cluster.RedisConf{Addr: redisHost, Password: redisPassword, DB: 0}, []*websocket_cluster.RedisNode{&websocket_cluster.RedisNode{
		RedisConf: websocket_cluster.RedisConf{Addr: redisHost, Password: redisPassword, DB: 0},
		Position:  1,
	}}, onMsg).WithPort(*port1).WithRouter(router.Router))
	if err != nil {
		logx.Info(err)
	}
```
```
func onMsg(context *websocket_cluster.Context) {
	logx.Info("msg:", string(context.Msg))
}
```

# 架构框图
![scheme 1](pic/websocket.png)

# 联系我们
- 技术支持/合作/咨询请联系作者QQ: 2276282419
- 作者邮箱: 2276282419@qq.com
- 即时通讯技术交流QQ群: 33280853
