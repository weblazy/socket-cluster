package master

import (
	"encoding/json"
	"flag"
	"hash/fnv"
	"os"
	socket_cluster "socket-cluster"
	"strconv"

	"github.com/weblazy/socket-cluster/examples/auth"
	"github.com/weblazy/socket-cluster/examples/common"
	"github.com/weblazy/socket-cluster/examples/model"
	"github.com/weblazy/socket-cluster/examples/router"

	"github.com/spf13/cast"
	"github.com/weblazy/core/database/redis"
	"github.com/weblazy/easy/utils/logx"
)

var (
	port1 = flag.Int64("port1", 9528, "the  port")
	host1 = flag.String("host1", "ws://localhost:9528", "the  host")
	path1 = flag.String("path1", "/p1", "the  path")
)

func Node1() {
	flag.Parse()
	var err error
	redisHost := os.Getenv("REDIS_HOST")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	err = auth.InitAuth(auth.NewAuthConf([]*auth.RedisNode{
		&auth.RedisNode{
			RedisConf: redis.RedisConf{Host: redisHost, Pass: redisPassword, Type: "node"},
			Position:  1,
		}}))
	if err != nil {
		panic(err)
	}
	common.NodeINfo1, err = socket_cluster.StartNode(socket_cluster.NewNodeConf(*host1, *path1, *path1, socket_cluster.RedisConf{Addr: redisHost, Password: redisPassword, DB: 0}, []*socket_cluster.RedisNode{&socket_cluster.RedisNode{
		RedisConf: socket_cluster.RedisConf{Addr: redisHost, Password: redisPassword, DB: 0},
		Position:  1,
	}}, onMsg).WithPort(*port1).WithRouter(router.Router))
	if err != nil {
		logx.Info(err)
	}
}

func getIndex(key string) int64 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int64(h.Sum32()) % model.TableNum
}

func onMsg(context *socket_cluster.Context) {
	logx.Info("msg:", string(context.Msg))
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
		uidStr, err := auth.AuthManager.Validate(token)
		uidInt := cast.ToInt64(uidStr)
		if err != nil {
			logx.Info(err)
			return
		}
		userIndex := getIndex(uidStr)
		obj, err := model.AuthModel().GetOne("id = ?", uidInt)
		if err != nil {
			logx.Info(err)
			return
		}
		common.NodeINfo1.AuthClient(context.Conn, uidStr)
		maxUserMsgId, err := model.UserMsgModel(userIndex).Max("notify_uid = ?", obj.Id)
		if err != nil {
			logx.Info(err)
			return
		}
		list, err := model.UserGroupModel().GetList("uid = ?", obj.Id)
		if err != nil {
			logx.Info(err)
			return
		}
		userGroupList := make([]map[string]interface{}, 0)
		for k1 := range list {
			v1 := list[k1]
			obj := map[string]interface{}{
				"group_id":         v1.GroupId,
				"max_group_msg_id": v1.MaxMsgId,
			}
			userGroupList = append(userGroupList, obj)
		}
		err = context.Conn.WriteJSON(map[string]interface{}{
			"msg_type": "init",
			"data": map[string]interface{}{
				"user_group_list": userGroupList,
				"max_user_msg_id": maxUserMsgId,
			},
		})
		if err != nil {
			logx.Info(err)
		}
	case "pull_group_msg":
		// uid, _ := strconv.ParseInt(context.ClientId, 10, 64)
		data := msgMap["data"].(map[string]interface{})
		groupId := cast.ToInt64(data["group_id"])
		groupIndex := getIndex(strconv.FormatInt(groupId, 10))

		lastGroupMsgId := cast.ToInt64(data["last_group_msg_id"])
		sort := data["sort"].(string)
		// group, err := model.UserGroupHandler.GetOne("uid = ? and group_id = ?", uid, groupId)
		// if err != nil {
		// 	logx.Info(err)
		// 	return
		// }
		// if lastGroupMsgId == 0 {
		// 	lastGroupMsgId = group.MaxPulledMsgId
		// } else {
		// 	if lastGroupMsgId > group.MaxPulledMsgId {
		// 		err = model.UserGroupHandler.Update(nil, map[string]interface{}{
		// 			"max_pulled_msg_id": lastGroupMsgId,
		// 		}, "uid = ? and group_id = ? and max_pulled_msg_id > ?", uid, groupId, lastGroupMsgId)
		// 		if err != nil {
		// 			logx.Info(err)
		// 			return
		// 		}
		// 	}
		// }
		var groupMsgList []*model.GroupMsg
		if sort == "desc" {
			groupMsgList, err = model.GroupMsgModel(groupIndex).GetListPage(50, "id desc", "group_id = ? and id < ?", groupId, lastGroupMsgId)
		} else {
			//只拉取最新50条
			groupMsgList, err = model.GroupMsgModel(groupIndex).GetListPage(50, "id asc", "group_id = ? and id <= (select max(id) from "+model.GroupMsgModel(groupIndex).TableName()+" where group_id = ?) and id > ?", groupId, groupId, lastGroupMsgId)
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
				"username":     v1.Username,
				"avatar":       v1.Avatar,
				"send_uid":     v1.SendUid,
				"group_id":     groupId,
				"group_msg_id": v1.Id,
				"content":      v1.Content,
				"created_at":   v1.CreatedAt.Unix(),
			}
			chatGroupMsgList = append(chatGroupMsgList, obj)
		}

		err = context.Conn.WriteJSON(map[string]interface{}{
			"msg_type": "pull_group_msg",
			"data": map[string]interface{}{
				"group_msg_list": chatGroupMsgList,
			},
		})
		if err != nil {
			logx.Info(err)
		}

	case "sync_user_msg":
		uidInt, _ := strconv.ParseInt(context.ClientId, 10, 64)
		userIndex := getIndex(context.ClientId)
		data := msgMap["data"].(map[string]interface{})
		lastUserMsgId := int64(data["last_user_msg_id"].(float64))
		if lastUserMsgId == 0 {
			user, err := model.AuthHandler.GetOne("id = ?", uidInt)
			if err != nil {
				logx.Info(err)
				return
			}
			lastUserMsgId = user.MaxPulledMsgId
		}
		userMsgList, err := model.UserMsgModel(userIndex).GetListPage(50, "notify_uid = ? and id > ?", uidInt, lastUserMsgId)
		if err != nil {
			logx.Info(err)
			return
		}
		if len(userMsgList) == 0 {
			return
		}
		list := make([]map[string]interface{}, 0)
		for k1 := range userMsgList {
			v1 := userMsgList[k1]
			obj := map[string]interface{}{
				"username":    v1.Username,
				"avatar":      v1.Avatar,
				"send_uid":    v1.SendUid,
				"receive_uid": v1.ReceiveUid,
				"user_msg_id": v1.Id,
				"content":     v1.Content,
				"created_at":  v1.CreatedAt.Unix(),
			}
			list = append(list, obj)
		}

		err = context.Conn.WriteJSON(map[string]interface{}{
			"msg_type": "sync_user_msg",
			"data": map[string]interface{}{
				"user_msg_list": list,
			},
		})
		if err != nil {
			logx.Info(err)
		}
	case "pull_msg":
		data := msgMap["data"].(map[string]interface{})
		lastUserMsgId := int64(data["last_user_msg_id"].(float64))
		groupList := data["group_list"].(map[string]interface{})
		uid, _ := strconv.ParseInt(context.ClientId, 10, 64)
		userIndex := getIndex(context.ClientId)
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
				if lastGroupMsgId < v1.MaxMsgId {
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
		receiveUid := cast.ToString(data["receive_uid"])
		userIndex := getIndex(context.ClientId)
		receiveIndex := getIndex(receiveUid)
		sendMsg := model.UserMsg{
			NotifyUid:  context.ClientId,
			Username:   data["username"].(string),
			Avatar:     data["avatar"].(string),
			ReceiveUid: receiveUid,
			MsgType:    "text",
			SendUid:    context.ClientId,
			Content:    data["content"].(string),
		}
		err = model.UserMsgModel(userIndex).Insert(nil, &sendMsg)
		if err != nil {
			logx.Info(err)
			return
		}
		common.NodeINfo1.SendToClientId(context.ClientId, map[string]interface{}{
			"msg_type": "have_new_msg",
			"data": map[string]interface{}{
				"max_user_msg_id": sendMsg.Id,
			},
		})
		if err != nil {
			logx.Info(err)
		}
		//发给他人
		if receiveUid != context.ClientId {
			receiveMsg := model.UserMsg{
				NotifyUid:  receiveUid,
				Username:   data["username"].(string),
				Avatar:     data["avatar"].(string),
				ReceiveUid: receiveUid,
				MsgType:    "text",
				SendUid:    context.ClientId,
				Content:    data["content"].(string),
			}
			err := model.UserMsgModel(receiveIndex).Insert(nil, &receiveMsg)
			if err != nil {
				logx.Info(err)
				return
			}
			common.NodeINfo1.SendToClientId(receiveUid, map[string]interface{}{
				"msg_type": "have_new_msg",
				"data": map[string]interface{}{
					"max_user_msg_id": receiveMsg.Id,
				},
			})
			if err != nil {
				logx.Info(err)
			}
		}

	case "ack_receive":
		data := msgMap["data"].(map[string]interface{})
		msgIdList := data["msg_id_list"].([]interface{})
		uid := context.ClientId
		userIndex := getIndex(context.ClientId)
		model.UserMsgModel(userIndex).Update(nil, map[string]interface{}{
			"status": 0,
		}, "receive_uid = ? and id in(?) and status = 0", uid, msgIdList)
	case "ack_user_msg":
		data := msgMap["data"].(map[string]interface{})
		lastMsgId := data["last_msg_id"].([]interface{})
		uid := context.ClientId
		model.AuthHandler.Update(nil, map[string]interface{}{
			"read_last_msg_id": lastMsgId,
		}, "id = ?", uid)
	case "ack_group_msg":
		data := msgMap["data"].(map[string]interface{})
		groupId := data["group_id"].([]interface{})
		lastMsgId := data["last_msg_id"].([]interface{})
		uid := context.ClientId
		model.UserGroupHandler.Update(nil, map[string]interface{}{
			"read_last_msg_id": lastMsgId,
		}, "uid = ? and group_id = ?", uid, groupId)
	case "send_to_group":
		data := msgMap["data"].(map[string]interface{})
		groupId := cast.ToString(data["group_id"])
		groupIndex := getIndex(groupId)
		msg := model.GroupMsg{
			Username: data["username"].(string),
			Avatar:   data["avatar"].(string),
			GroupId:  groupId,
			MsgType:  "text",
			SendUid:  context.ClientId,
			Content:  data["content"].(string),
		}
		err := model.GroupMsgModel(groupIndex).Insert(nil, &msg)
		if err != nil {
			logx.Info(err)
			return
		}
		model.UserGroupHandler.Update(nil, map[string]interface{}{
			"max_msg_id": msg.Id,
		}, "group_id = ? and max_msg_id < ?", groupId, msg.Id)
		groupIdInt, _ := strconv.ParseInt(groupId, 10, 64)
		userGroupList, err := model.UserGroupHandler.GetList("group_id = ?", groupIdInt)
		if err != nil {
			logx.Info(err)
			return
		}
		uids := make([]string, 0)
		for k1 := range userGroupList {
			uids = append(uids, cast.ToString(userGroupList[k1].Uid))
		}
		common.NodeINfo1.SendToClientIds(uids, map[string]interface{}{
			"msg_type": "have_new_msg",
			"data": map[string]interface{}{
				"max_group_msg_id": msg.Id,
				"group_id":         groupId,
			},
		})
	default:
		logx.Info(string(context.Msg))
	}
}
