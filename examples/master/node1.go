package main

import (
	"encoding/json"
	"flag"
	"hash/fnv"
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
	}}, onMsg).WithPort(*port1).WithRouter(Router))
}

func getIndex(key string) int64 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int64(h.Sum32()) % model.TableNum
}

func onMsg(nodeInfo *websocket_cluster.NodeInfo, context *websocket_cluster.Context) {
	logx.Info("msg1", string(context.Msg))
	msgMap := make(map[string]interface{})
	err := json.Unmarshal(context.Msg, &msgMap)
	if err != nil {
		logx.Info(err)
	}
	v1, ok := msgMap["msg_type"]
	if !ok {
		logx.Info("msg_type is nil")
	}
	switch v1 {
	case "init":
		data := msgMap["data"].(map[string]interface{})
		token := data["token"].(string)
		uid, err := auth.AuthManager.Validate(token)
		if err != nil {
			logx.Info(err)
			return
		}
		userIndex := getIndex(uid)
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

		list, err := model.UserMsgModel(userIndex).GetList("receive_uid = ? and status = 0", obj.Id)
		if err != nil {
			logx.Info(err)
			return
		}
		if len(list) == 0 {
			return
		}
		chatMsgList := make([]map[string]interface{}, 0)
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
			chatMsgList = append(chatMsgList, obj)
		}

		//查询是否存在未读群消息
		userGroupList, err := model.UserGroupHandler.GetList("uid  = ? and last_msg_id > read_last_msg_id", uid)
		if err != nil {
			logx.Info(err)
			return
		}
		groupList := make([]map[string]interface{}, 0)
		for k1 := range userGroupList {
			v1 := userGroupList[k1]
			groupList = append(groupList, map[string]interface{}{
				"group_id":    v1.GroupId,
				"last_msg_id": v1.LastMsgId,
			})
		}
		err = context.Conn.WriteJSON(map[string]interface{}{
			"msg_type":    "chat_msg_list",
			"receive_uid": uid,
			"data": map[string]interface{}{
				"list":       chatMsgList,
				"group_list": groupList,
			},
		})
		if err != nil {
			logx.Info(err)
		}
	case "pull_group_msg":
		uid, _ := strconv.ParseInt(context.Uid, 10, 64)
		data := msgMap["data"].(map[string]interface{})
		groupId := int64(data["group_id"].(float64))
		groupIndex := getIndex(strconv.FormatInt(groupId, 10))

		lastGroupMsgId := int64(data["last_group_msg_id"].(float64))
		sort := data["sort"].(string)
		if lastGroupMsgId == 0 {
			group, err := model.UserGroupHandler.GetOne("uid = ? and group_id = ?", uid, groupId)
			if err != nil {
				logx.Info(err)
				return
			}
			lastGroupMsgId = group.ReadLastMsgId
		}
		var groupMsgList []*model.GroupMsg
		if sort == "desc" {
			groupMsgList, err = model.GroupMsgModel(groupIndex).GetListPage(50, "id desc", "group_id = ? and id < ?", groupId, lastGroupMsgId)
		} else {
			groupMsgList, err = model.GroupMsgModel(groupIndex).GetListPage(50, "id asc", "group_id = ? and id > ?", groupId, lastGroupMsgId)

		}
		if err != nil {
			logx.Info(err)
			return
		}
		if len(groupMsgList) == 0 {
			return
		}
		chatGroupMsgList := make([]map[string]interface{}, 0)
		for k1 := range groupMsgList {
			v1 := groupMsgList[k1]
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
			chatGroupMsgList = append(chatGroupMsgList, obj)
		}

		err = context.Conn.WriteJSON(map[string]interface{}{
			"msg_type":    "gorup_msg_list",
			"receive_uid": uid,
			"data": map[string]interface{}{
				"user_group_msg_list": chatGroupMsgList,
			},
		})
		if err != nil {
			logx.Info(err)
		}

	case "pull_user_msg":
		uid, _ := strconv.ParseInt(context.Uid, 10, 64)
		userIndex := getIndex(context.Uid)
		data := msgMap["data"].(map[string]interface{})
		lastUserMsgId := int64(data["last_user_msg_id"].(float64))
		if lastUserMsgId == 0 {
			user, err := model.AuthHandler.GetOne("uid = ?", uid)
			if err != nil {
				logx.Info(err)
				return
			}
			lastUserMsgId = user.MaxReadMsgId
		}
		userMsgList, err := model.UserMsgModel(userIndex).GetListPage(50, "receive_uid = ? or send_uid = ? and id > ?", uid, uid, lastUserMsgId)
		if err != nil {
			logx.Info(err)
			return
		}
		if len(userMsgList) == 0 {
			return
		}
		chatUserMsgList := make([]map[string]interface{}, 0)
		for k1 := range userMsgList {
			v1 := userMsgList[k1]
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
			chatUserMsgList = append(chatUserMsgList, obj)
		}

		err = context.Conn.WriteJSON(map[string]interface{}{
			"msg_type":    "user_msg_list",
			"receive_uid": uid,
			"data": map[string]interface{}{
				"user_msg_list": chatUserMsgList,
			},
		})
		if err != nil {
			logx.Info(err)
		}
	case "pull_msg":
		data := msgMap["data"].(map[string]interface{})
		lastUserMsgId := int64(data["last_user_msg_id"].(float64))
		groupList := data["group_list"].(map[string]interface{})
		uid, _ := strconv.ParseInt(context.Uid, 10, 64)
		userIndex := getIndex(context.Uid)
		_, err = model.UserGroupHandler.GetList("uid = ?", uid)
		if err != nil {
			logx.Info(err)
			return
		}

		userMsgList, err := model.UserMsgModel(userIndex).GetList("receive_uid = ? and id > ?", uid, lastUserMsgId)
		if err != nil {
			logx.Info(err)
			return
		}

		if len(userMsgList) == 0 {
			return
		}
		chatUserMsgList := make([]map[string]interface{}, 0)
		for k1 := range userMsgList {
			v1 := userMsgList[k1]
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
			chatUserMsgList = append(chatUserMsgList, obj)
		}

		userGroupList, err := model.UserGroupHandler.GetList("uid = ?", uid)
		if err != nil {
			logx.Info(err)
			return
		}
		if len(userGroupList) == 0 {
			return
		}

		GroupMsgList := make([]*model.GroupMsg, 0)
		for k1 := range userGroupList {
			v1 := userGroupList[k1]
			groupId := strconv.FormatInt(v1.GroupId, 10)
			v2, ok := groupList[groupId].(map[string]interface{})
			if ok {
				lastGroupMsgId := int64(v2["last_group_msg_id"].(float64))
				if lastGroupMsgId < v1.LastMsgId {
					groupIndex := getIndex(groupId)
					TempGroupMsgList, err := model.GroupMsgModel(groupIndex).GetListPage(50, "id asc", "id > ?", lastGroupMsgId)
					if err != nil {
						logx.Info(err)
						return
					}
					if len(TempGroupMsgList) == 0 {
						return
					}
					GroupMsgList = append(GroupMsgList, TempGroupMsgList...)
				}
			}
		}
		if len(GroupMsgList) == 0 {
			return
		}
		chatGroupMsgList := make([]map[string]interface{}, 0)
		for k1 := range GroupMsgList {
			v1 := GroupMsgList[k1]
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
			chatGroupMsgList = append(chatGroupMsgList, obj)
		}

		err = context.Conn.WriteJSON(map[string]interface{}{
			"msg_type":    "chat_msg_list",
			"receive_uid": uid,
			"data": map[string]interface{}{
				"user_msg_list":       chatUserMsgList,
				"user_group_msg_list": chatGroupMsgList,
			},
		})
		if err != nil {
			logx.Info(err)
		}
	case "send_to_user":
		data := msgMap["data"].(map[string]interface{})
		receiveUidFloat := data["receive_uid"].(float64)
		receiveUid := strconv.FormatFloat(receiveUidFloat, 'f', -1, 64)
		userIndex := getIndex(context.Uid)
		msg := model.UserMsg{
			Username:   data["username"].(string),
			Avatar:     data["avatar"].(string),
			ReceiveUid: receiveUid,
			MsgType:    "text",
			SendUid:    context.Uid,
			Content:    data["content"].(string),
			Status:     0,
		}
		err := model.UserMsgModel(userIndex).Insert(nil, &msg)
		if err != nil {
			logx.Info(err)
			return
		}
		err = context.Conn.WriteJSON(map[string]interface{}{
			"msg_type":    "have_new_msg",
			"receive_uid": receiveUid,
			"data": map[string]interface{}{
				"last_user_msg_id": msg.Id,
			},
		})
		if err != nil {
			logx.Info(err)
		}
	case "ack_receive":
		data := msgMap["data"].(map[string]interface{})
		msgIdList := data["msg_id_list"].([]interface{})
		uid := context.Uid
		userIndex := getIndex(context.Uid)
		model.UserMsgModel(userIndex).Update(nil, map[string]interface{}{
			"status": 0,
		}, "receive_uid = ? and id in(?) and status = 0", uid, msgIdList)
	case "ack_user_msg":
		data := msgMap["data"].(map[string]interface{})
		lastMsgId := data["last_msg_id"].([]interface{})
		uid := context.Uid
		model.AuthHandler.Update(nil, map[string]interface{}{
			"read_last_msg_id": lastMsgId,
		}, "id = ?", uid)
	case "ack_group_msg":
		data := msgMap["data"].(map[string]interface{})
		groupId := data["group_id"].([]interface{})
		lastMsgId := data["last_msg_id"].([]interface{})
		uid := context.Uid
		model.UserGroupHandler.Update(nil, map[string]interface{}{
			"read_last_msg_id": lastMsgId,
		}, "uid = ? and group_id = ?", uid, groupId)
	case "send_to_group":
		data := msgMap["data"].(map[string]interface{})
		groupIdFloat := data["group_id"].(float64)
		groupId := strconv.FormatFloat(groupIdFloat, 'f', -1, 64)
		groupIndex := getIndex(groupId)
		msg := model.GroupMsg{
			Username: data["username"].(string),
			Avatar:   data["avatar"].(string),
			GroupId:  groupId,
			MsgType:  "text",
			SendUid:  context.Uid,
			Content:  data["content"].(string),
		}
		err := model.GroupMsgModel(groupIndex).Insert(nil, &msg)
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
			"msg_type": "have_new_msg",
			"data": map[string]interface{}{
				"last_group_msg_id": msg.Id,
			},
		})
	default:
		logx.Info(string(context.Msg))
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
