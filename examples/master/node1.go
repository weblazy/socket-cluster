package main

import (
	"encoding/json"
	"flag"
	"strconv"
	websocket_cluster "websocket-cluster"
	"websocket-cluster/examples/api"
	"websocket-cluster/examples/auth"
	"websocket-cluster/examples/model"

	"github.com/labstack/echo/v4"
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
	}}, onMessage).WithPort(*port1).WithRouter(Router))
}

func onMessage(nodeInfo *websocket_cluster.NodeInfo, context *websocket_cluster.Context) {
	logx.Info("message1", string(context.Message))
	messageMap := make(map[string]interface{})
	err := json.Unmarshal(context.Message, &messageMap)
	if err != nil {
		logx.Info(err)
	}
	v1, ok := messageMap["message_type"]
	if !ok {
		logx.Info("message_type is nil")
	}
	switch v1 {
	case "login":
		data := messageMap["data"].(map[string]interface{})
		obj, err := model.AuthHandler.GetOne("username = ? and password = ?", data["username"].(string), data["password"].(string))
		if err != nil {
			logx.Info(err)
			return
		}
		uid := strconv.FormatInt(obj.Id, 10)
		token, err := auth.AuthManager.Add(uid)
		logx.Info(token)
		if err != nil {
			logx.Info(err)
			return
		} else {
			logx.Info(token)
			id, err := auth.AuthManager.Validate(token)
			logx.Info(id)
			logx.Info(err)
		}
		nodeInfo.AuthClient(context.Conn, uid)
		_, err = model.UserGroupHandler.GetList("uid = ?", uid)
		if err != nil {
			logx.Info(err)
			return
		}

		list, err := model.MessageHandler.GetList("receive_uid = ? and status = 0", uid)
		if err != nil {
			logx.Info(err)
			return
		}
		if len(list) == 0 {
			return
		}
		err = context.Conn.WriteJSON(map[string]interface{}{
			"message_type": "chat_message_list",
			"data": map[string]interface{}{
				"list":        list,
				"receive_uid": uid,
			},
		})
		if err != nil {
			logx.Info(err)
		}
	case "chat_message_list":
		data := messageMap["data"].(map[string]interface{})
		receiveUid := data["receive_uid"].(string)
		list := data["list"].([]interface{})
		for k1 := range list {
			v1 := list[k1].(map[string]interface{})
			message := model.Message{
				ReceiveUid:  receiveUid,
				MessageType: "text",
				SendUid:     context.Uid,
				Content:     v1["content"].(string),
				Status:      0,
			}
			err := model.MessageHandler.Insert(nil, &message)
			if err != nil {
				logx.Info(err)
				return
			}
			v1["id"] = message.Id
			v1["send_uid"] = context.Uid
			list[k1] = v1
		}
		messageMap["list"] = list
		messageMap["receive_uid"] = receiveUid
		nodeInfo.SendToUid(receiveUid, messageMap)
	case "ack_receive":
		data := messageMap["data"].(map[string]interface{})
		messageIdList := data["message_id_list"].([]interface{})
		uid := context.Uid
		model.MessageHandler.Update(nil, map[string]interface{}{
			"status": 1,
		}, "receive_uid = ? and id in(?) and status = 0", uid, messageIdList)
	case "send_to_group":
		data := messageMap["data"].(map[string]interface{})
		groupId := data["group_id"].(string)
		list := data["list"].([]interface{})
		for k1 := range list {
			v1 := list[k1].(map[string]interface{})
			message := model.GroupMessage{
				GroupId:     groupId,
				MessageType: "text",
				SendUid:     context.Uid,
				Content:     v1["content"].(string),
			}
			err := model.GroupMessageHandler.Insert(nil, &message)
			if err != nil {
				logx.Info(err)
				return
			}
			v1["id"] = message.Id
			v1["send_uid"] = context.Uid
			list[k1] = v1
		}
		messageMap["list"] = list
		messageMap["group_id"] = groupId
		groupIdInt, _ := strconv.ParseInt(groupId, 10, 64)
		userGroupList, err := model.UserGroupHandler.GetList("group_id = ?", groupIdInt)
		if err != nil {
			logx.Info(err)
			return
		}
		uids := make([]string, 0)
		for k1 := range userGroupList {
			uidStr := strconv.FormatInt(userGroupList[k1].Uid, 10)
			uids = append(uids, uidStr)
		}
		nodeInfo.SendToUids(uids, messageMap)
	default:
		logx.Info(string(context.Message))
	}
}

// @desc
// @auth liuguoqiang 2020-11-20
// @param
// @return
func Router(g *echo.Group) {
	g.POST("/login", api.Login)
	g.POST("/chatInit", api.ChatInit)
	g.POST("/getGroupMembers", api.GetGroupMembers)
}
