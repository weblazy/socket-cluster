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

		//查询是否存在未读群消息
		userGroupList, err := model.UserGroupHandler.GetList("uid  = ? and last_message_id > read_last_message_id", uid)
		if err != nil {
			logx.Info(err)
			return
		}
		groupList := make([]map[string]interface{}, 0)
		for k1 := range userGroupList {
			v1 := userGroupList[k1]
			groupList = append(groupList, map[string]interface{}{
				"group_id":        v1.GroupId,
				"last_message_id": v1.LastMessageId,
			})
		}
		err = context.Conn.WriteJSON(map[string]interface{}{
			"message_type": "chat_message_list",
			"receive_uid":  uid,
			"data": map[string]interface{}{
				"list":       chatMessageList,
				"group_list": groupList,
			},
		})
		if err != nil {
			logx.Info(err)
		}
	case "pull_group_message":
		uid, _ := strconv.ParseInt(context.Uid, 10, 64)
		data := messageMap["data"].(map[string]interface{})
		groupId := int64(data["group_id"].(float64))
		lastGroupMessageId := int64(data["last_group_message_id"].(float64))
		sort := data["sort"].(string)
		if lastGroupMessageId == 0 {
			group, err := model.UserGroupHandler.GetOne("uid = ? and group_id = ?", uid, groupId)
			if err != nil {
				logx.Info(err)
				return
			}
			lastGroupMessageId = group.ReadLastMessageId
		}
		var groupMessageList []*model.GroupMessage
		if sort == "desc" {
			groupMessageList, err = model.GroupMessageHandler.GetListPage(50, "id desc", "group_id = ? and id < ?", groupId, lastGroupMessageId)
		} else {
			groupMessageList, err = model.GroupMessageHandler.GetListPage(50, "id asc", "group_id = ? and id > ?", groupId, lastGroupMessageId)

		}
		if err != nil {
			logx.Info(err)
			return
		}
		if len(groupMessageList) == 0 {
			return
		}
		chatGroupMessageList := make([]map[string]interface{}, 0)
		for k1 := range groupMessageList {
			v1 := groupMessageList[k1]
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
			"message_type": "gorup_message_list",
			"receive_uid":  uid,
			"data": map[string]interface{}{
				"user_group_message_list": chatGroupMessageList,
			},
		})
		if err != nil {
			logx.Info(err)
		}

	case "pull_user_message":
		uid, _ := strconv.ParseInt(context.Uid, 10, 64)
		data := messageMap["data"].(map[string]interface{})
		lastUserMessageId := int64(data["last_user_message_id"].(float64))
		if lastUserMessageId == 0 {
			user, err := model.AuthHandler.GetOne("uid = ?", uid)
			if err != nil {
				logx.Info(err)
				return
			}
			lastUserMessageId = user.ReadLastMessageId
		}
		userMessageList, err := model.UserMessageHandler.GetListPage(50, "receive_uid = ? or send_uid = ? and id > ?", uid, uid, lastUserMessageId)
		if err != nil {
			logx.Info(err)
			return
		}
		if len(userMessageList) == 0 {
			return
		}
		chatUserMessageList := make([]map[string]interface{}, 0)
		for k1 := range userMessageList {
			v1 := userMessageList[k1]
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
			chatUserMessageList = append(chatUserMessageList, obj)
		}

		err = context.Conn.WriteJSON(map[string]interface{}{
			"message_type": "user_message_list",
			"receive_uid":  uid,
			"data": map[string]interface{}{
				"user_message_list": chatUserMessageList,
			},
		})
		if err != nil {
			logx.Info(err)
		}
	case "pull_message":
		data := messageMap["data"].(map[string]interface{})
		lastUserMessageId := int64(data["last_user_message_id"].(float64))
		groupList := data["group_list"].(map[string]interface{})
		uid, _ := strconv.ParseInt(context.Uid, 10, 64)

		_, err = model.UserGroupHandler.GetList("uid = ?", uid)
		if err != nil {
			logx.Info(err)
			return
		}

		userMessageList, err := model.UserMessageHandler.GetList("receive_uid = ? and id > ?", uid, lastUserMessageId)
		if err != nil {
			logx.Info(err)
			return
		}

		if len(userMessageList) == 0 {
			return
		}
		chatUserMessageList := make([]map[string]interface{}, 0)
		for k1 := range userMessageList {
			v1 := userMessageList[k1]
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

		userGroupList, err := model.UserGroupHandler.GetList("uid = ?", uid)
		if err != nil {
			logx.Info(err)
			return
		}
		if len(userGroupList) == 0 {
			return
		}

		GroupMessageList := make([]*model.GroupMessage, 0)
		for k1 := range userGroupList {
			v1 := userGroupList[k1]
			groupId := strconv.FormatInt(v1.GroupId, 10)
			v2, ok := groupList[groupId].(map[string]interface{})
			if ok {
				lastGroupMessageId := int64(v2["last_group_message_id"].(float64))
				if lastGroupMessageId < v1.LastMessageId {
					TempGroupMessageList, err := model.GroupMessageHandler.GetListPage(50, "id > ?", lastGroupMessageId)
					if err != nil {
						logx.Info(err)
						return
					}
					if len(TempGroupMessageList) == 0 {
						return
					}
					GroupMessageList = append(GroupMessageList, TempGroupMessageList...)
				}
			}
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
		err = context.Conn.WriteJSON(map[string]interface{}{
			"message_type": "have_new_message",
			"receive_uid":  receiveUid,
			"data": map[string]interface{}{
				"last_user_message_id": message.Id,
			},
		})
		if err != nil {
			logx.Info(err)
		}
	case "ack_receive":
		data := messageMap["data"].(map[string]interface{})
		messageIdList := data["message_id_list"].([]interface{})
		uid := context.Uid
		model.UserMessageHandler.Update(nil, map[string]interface{}{
			"status": 0,
		}, "receive_uid = ? and id in(?) and status = 0", uid, messageIdList)
	case "ack_user_message":
		data := messageMap["data"].(map[string]interface{})
		lastMessageId := data["last_message_id"].([]interface{})
		uid := context.Uid
		model.AuthHandler.Update(nil, map[string]interface{}{
			"read_last_message_id": lastMessageId,
		}, "id = ?", uid)
	case "ack_group_message":
		data := messageMap["data"].(map[string]interface{})
		groupId := data["group_id"].([]interface{})
		lastMessageId := data["last_message_id"].([]interface{})
		uid := context.Uid
		model.UserGroupHandler.Update(nil, map[string]interface{}{
			"read_last_message_id": lastMessageId,
		}, "uid = ? and group_id = ?", uid, groupId)
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
		nodeInfo.SendToUids(uids, map[string]interface{}{
			"message_type": "have_new_message",
			"data": map[string]interface{}{
				"last_group_message_id": message.Id,
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
