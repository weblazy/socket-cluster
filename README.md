# socket-cluster
即时通讯框架,如果大家觉得好用,右上角帮忙点个star吧。(^_^)
> 欢迎感兴趣的小伙伴一同开发,解决日常消息推送,长连接需求。
- 后端(golang): https://github.com/weblazy/socket-cluster
- 前端: https://github.com/weblazy/socket-cluster-web
- 文档: socket-cluster-doc
- 后端好用的工具库(golang): https://github.com/weblazy/easy
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
export DBDEFAULT_DBNAME="socket_cluster"
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
cd socket-cluster/examples/main/
./main
```
# 框架使用方式
```
	common.NodeINfo1, err = socket_cluster.StartNode(socket_cluster.NewNodeConf(*host1, *path1, *path1, socket_cluster.RedisConf{Addr: redisHost, Password: redisPassword, DB: 0}, []*socket_cluster.RedisNode{&socket_cluster.RedisNode{
		RedisConf: socket_cluster.RedisConf{Addr: redisHost, Password: redisPassword, DB: 0},
		Position:  1,
	}}, onMsg).WithPort(*port1).WithRouter(router.Router))
	if err != nil {
		logx.Info(err)
	}
```
```
func onMsg(context *socket_cluster.Context) {
	logx.Info("msg:", string(context.Msg))
}
```

# 架构框图
![scheme 1](pic/websocket.png)

# 联系我们
- 技术支持/合作/咨询请联系作者QQ: 2276282419
- 作者邮箱: 2276282419@qq.com
- 即时通讯技术交流QQ群: 33280853
