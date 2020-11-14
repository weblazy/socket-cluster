package main

import (
	"encoding/json"
	"flag"
	"fmt"
	websocket_cluster "websocket-cluster"
	"websocket-cluster/examples/auth"
	"websocket-cluster/examples/model"

	"github.com/weblazy/core/database/redis"
	"github.com/weblazy/easy/utils/logx"
)

var (
	port1         = flag.Int64("port1", 9528, "the  port")
	host1         = flag.String("host1", "ws://localhost:9528", "the  host")
	path1         = flag.String("path1", "/p1", "the  path")
	masterAddress = flag.String("masterAddress", "ws://127.0.0.1:9527/master", "the  masterAddress")
)

func node1() {
	flag.Parse()
	auth.InitAuth(auth.NewAuthConf([]*auth.RedisNode{
		&auth.RedisNode{
			RedisConf: redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"},
			Position:  1,
		}}))
	websocket_cluster.StartNode(websocket_cluster.NewNodeConf(*host1, *path1, *masterAddress, redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"}, []*websocket_cluster.RedisNode{&websocket_cluster.RedisNode{
		RedisConf: redis.RedisConf{Host: "127.0.0.1:6379", Type: "node"},
		Position:  1,
	}}, onMessage).WithPort(*port1))
}

func onMessage(context *websocket_cluster.Context) {
	logx.Info("message1", string(context.Message))
	data := make(map[string]interface{})
	err := json.Unmarshal(context.Message, &data)
	if err != nil {
		logx.Info(err)
	}
	v1, ok := data["message_type"]
	if !ok {
		logx.Info("message_type is nil")
	}
	switch v1 {
	case "login":
		content := data["data"].(map[string]interface{})
		// nodeInfo.AuthClient(conn, data["data"].(map[string]interface{}))
		obj, err := model.AuthHandler.GetOne("username = ? and password = ?", content["username"].(string), content["password"].(string))
		if err != nil {
			logx.Info(err)
		}
		token, err := auth.AuthManager.Add(fmt.Sprintf("%d", obj.Id))
		logx.Info(token)
		if err != nil {
			logx.Info(err)
		} else {
			logx.Info(token)
			id, err := auth.AuthManager.Validate(token)
			logx.Info(id)
			logx.Info(err)
		}
	case "chatMessage":
		// nodeInfo.SendToUid(data["data"].(map[string]interface{})["uid"].(string), "", data)
	default:
		logx.Info(string(context.Message))
	}
}
