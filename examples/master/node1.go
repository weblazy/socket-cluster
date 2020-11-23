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
	case "init":
		data := messageMap["data"].(map[string]interface{})
		token := data["token"].(string)
		uid, err := auth.AuthManager.Validate(token)
		if err != nil {
			logx.Info(err)
			return
		}
		obj, err := model.AuthHandler.GetOne("id = ?", uid)
		if err != nil {
			logx.Info(err)
			return
		}
		nodeInfo.AuthClient(context.Conn, uid)
		_, err = model.UserGroupHandler.GetList("uid = ?", obj.Id)
		if err != nil {
			logx.Info(err)
			return
		}

		list, err := model.UserMessageHandler.GetList("receive_uid = ? and status = 0", obj.Id)
		if err != nil {
			logx.Info(err)
			return
		}
		if len(list) == 0 {
			return
		}
		chatMessageList := make([]map[string]interface{}, 0)
		for k1 := range list {
			v1 := list[k1]
			obj := map[string]interface{}{
				"username":  v1.Username,
				"avatar":    v1.Avatar,
				"id":        v1.SendUid,
				"type":      "friend",
				"content":   v1.Content,
				"cid":       v1.Id,
				"mine":      false,
				"fromid":    v1.SendUid,
				"timestamp": v1.CreatedAt.Unix() * 1000,
			}
			chatMessageList = append(chatMessageList, obj)
		}
		err = context.Conn.WriteJSON(map[string]interface{}{
			"message_type": "chat_message_list",
			"receive_uid":  uid,
			"data": map[string]interface{}{
				"list": chatMessageList,
			},
		})
		if err != nil {
			logx.Info(err)
		}
	case "pull_message":
		data := messageMap["data"].(map[string]interface{})
		lastUserMessageId := int64(data["last_user_message_id"].(float64))
		lastGroupMessageId := int64(data["last_group_message_id"].(float64))
		uid, _ := strconv.ParseInt(context.Uid, 10, 64)

		_, err = model.UserGroupHandler.GetList("uid = ?", uid)
		if err != nil {
			logx.Info(err)
			return
		}

		UserMessageList, err := model.UserMessageHandler.GetList("receive_uid = ? and id > ?", uid, lastUserMessageId)
		if err != nil {
			logx.Info(err)
			return
		}

		if len(UserMessageList) == 0 {
			return
		}
		chatUserMessageList := make([]map[string]interface{}, 0)
		for k1 := range UserMessageList {
			v1 := UserMessageList[k1]
			obj := map[string]interface{}{
				"username":  v1.Username,
				"avatar":    v1.Avatar,
				"id":        v1.SendUid,
				"type":      "friend",
				"content":   v1.Content,
				"cid":       v1.Id,
				"mine":      false,
				"fromid":    v1.SendUid,
				"timestamp": v1.CreatedAt.Unix() * 1000,
			}
			chatUserMessageList = append(chatUserMessageList, obj)
		}

		UserGroupMessageList, err := model.UserGroupMessageHandler.GetList("id > ?", lastUserMessageId)
		if err != nil {
			logx.Info(err)
			return
		}
		if len(UserGroupMessageList) == 0 {
			return
		}
		groupMessageIds := make([]int64, 0)
		for k1 := range UserGroupMessageList {
			groupMessageIds = append(groupMessageIds, UserGroupMessageList[k1].GroupMessageId)
		}

		GroupMessageList, err := model.GroupMessageHandler.GetList("id in(?)", groupMessageIds)
		if err != nil {
			logx.Info(err)
			return
		}
		if len(GroupMessageList) == 0 {
			return
		}
		chatGroupMessageList := make([]map[string]interface{}, 0)
		for k1 := range GroupMessageList {
			v1 := GroupMessageList[k1]
			obj := map[string]interface{}{
				"username":  v1.Username,
				"avatar":    v1.Avatar,
				"id":        v1.SendUid,
				"type":      "group",
				"content":   v1.Content,
				"cid":       v1.Id,
				"mine":      false,
				"fromid":    v1.SendUid,
				"timestamp": v1.CreatedAt.Unix() * 1000,
			}
			chatGroupMessageList = append(chatGroupMessageList, obj)
		}

		err = context.Conn.WriteJSON(map[string]interface{}{
			"message_type": "chat_message_list",
			"receive_uid":  uid,
			"data": map[string]interface{}{
				"user_message_list":       chatUserMessageList,
				"user_group_message_list": chatGroupMessageList,
			},
		})
		if err != nil {
			logx.Info(err)
		}
	case "send_to_user":
		data := messageMap["data"].(map[string]interface{})
		receiveUidFloat := data["receive_uid"].(float64)
		receiveUid := strconv.FormatFloat(receiveUidFloat, 'f', -1, 64)
		message := model.UserMessage{
			Username:    data["username"].(string),
			Avatar:      data["avatar"].(string),
			ReceiveUid:  receiveUid,
			MessageType: "text",
			SendUid:     context.Uid,
			Content:     data["content"].(string),
			Status:      0,
		}
		err := model.UserMessageHandler.Insert(nil, &message)
		if err != nil {
			logx.Info(err)
			return
		}
		chatMessageList := make([]map[string]interface{}, 0)
		obj := map[string]interface{}{
			"username":  message.Username,
			"avatar":    message.Avatar,
			"id":        message.SendUid,
			"type":      "friend",
			"content":   message.Content,
			"cid":       message.Id,
			"mine":      false,
			"fromid":    message.SendUid,
			"timestamp": message.CreatedAt.Unix() * 1000,
		}
		chatMessageList = append(chatMessageList, obj)
		nodeInfo.SendToUid(receiveUid, map[string]interface{}{
			"message_type": "chat_message_list",
			"receive_uid":  receiveUid,
			"data": map[string]interface{}{
				"list": chatMessageList,
			},
		})

	case "ack_receive":
		data := messageMap["data"].(map[string]interface{})
		messageIdList := data["message_id_list"].([]interface{})
		uid := context.Uid
		model.UserMessageHandler.Update(nil, map[string]interface{}{
			"status": 0,
		}, "receive_uid = ? and id in(?) and status = 0", uid, messageIdList)
	case "send_to_group":
		data := messageMap["data"].(map[string]interface{})
		groupIdFloat := data["group_id"].(float64)
		groupId := strconv.FormatFloat(groupIdFloat, 'f', -1, 64)
		message := model.GroupMessage{
			Username:    data["username"].(string),
			Avatar:      data["avatar"].(string),
			GroupId:     groupId,
			MessageType: "text",
			SendUid:     context.Uid,
			Content:     data["content"].(string),
		}
		err := model.GroupMessageHandler.Insert(nil, &message)
		if err != nil {
			logx.Info(err)
			return
		}
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
		chatMessageList := make([]map[string]interface{}, 0)
		obj := map[string]interface{}{
			"username":  message.Username,
			"avatar":    message.Avatar,
			"id":        message.GroupId,
			"type":      "group",
			"content":   message.Content,
			"cid":       message.Id,
			"mine":      false,
			"fromid":    message.SendUid,
			"timestamp": message.CreatedAt.Unix() * 1000,
		}
		chatMessageList = append(chatMessageList, obj)
		nodeInfo.SendToUids(uids, map[string]interface{}{
			"message_type": "chat_message_list",
			"data": map[string]interface{}{
				"list": chatMessageList,
			},
		})
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
	g.OPTIONS("/login", websocket_cluster.OptionHandler)
	g.OPTIONS("/chatInit", websocket_cluster.OptionHandler)
	g.OPTIONS("/getGroupMembers", websocket_cluster.OptionHandler)
}
