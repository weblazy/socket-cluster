package domain

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
	"websocket-cluster/examples/auth"
	"websocket-cluster/examples/common"
	"websocket-cluster/examples/model"

	"github.com/jinzhu/gorm"
	"github.com/spf13/cast"
	"github.com/weblazy/core/logx"
	"gopkg.in/gomail.v2"
)

type ()

// @desc 登录
// @auth liuguoqiang 2020-11-20
// @param
// @return
func Login(username, password string) (map[string]interface{}, error) {
	user, err := model.AuthHandler.GetOne("username = ? and password = ?", username, password)
	if err != nil {
		return nil, err
	}
	uid := strconv.FormatInt(user.Id, 10)
	token, err := auth.AuthManager.Add(uid)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"uid":      user.Id,
		"username": user.Username,
		"email":    user.Email,
		"avatar":   user.Avatar,
		"token":    token,
	}, nil
}

// @desc 注册
// @auth liuguoqiang 2020-11-20
// @param
// @return
func Register(username, password, confirmPassword, email, code string) (map[string]interface{}, error) {
	if password != confirmPassword {
		return nil, fmt.Errorf("密码不一致")
	}
	_, err := model.AuthHandler.GetOne("email = ?", email)
	if err == nil {
		return nil, fmt.Errorf("该邮箱已经注册")
	}
	smsCode, err := model.SmsCodeModel().GetOne("email = ? and code  = ? and msg_type = 1", email, code)
	if err != nil {
		return nil, fmt.Errorf("该验证码不存在")
	}
	if smsCode.Status != 1 || smsCode.CreatedAt.Unix()+3600 < time.Now().Unix() {
		return nil, fmt.Errorf("该验证码已失效")
	}
	user := model.Auth{
		Username: username,
		Email:    email,
		Avatar:   "http://images.cookmami.com/data_upload_activity_bg_5a30a152a163a.png",
		Password: password,
	}

	err = model.AuthHandler.Insert(nil, &user)
	if err != nil {
		return nil, err
	}
	_, err = model.SmsCodeModel().Update(nil, map[string]interface{}{
		"status": 2,
	}, "id = ?", smsCode.Id)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"uid": user.Id,
	}, nil
}

// @desc 发送验证码
// @auth liuguoqiang 2020-11-20
// @param
// @return
func SendSmsCode(email string) (map[string]interface{}, error) {
	code := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
	now := time.Now()
	smsCode, err := model.SmsCodeModel().GetOne("email = ? and created_at > ? and msg_type = 1 and status in(0,1)", email, now.Add(-1*time.Hour))
	if err == nil {
		code = smsCode.Code
	} else {
		smsCode = &model.SmsCode{
			Email:   email,
			Code:    code,
			MsgType: 1,
			Status:  0,
		}
		err = model.SmsCodeModel().Insert(nil, smsCode)
		if err != nil {
			return nil, fmt.Errorf("数据库插入失败")
		}
	}
	tx := model.Orm().Begin()
	defer tx.RollbackUnlessCommitted()
	num, err := model.SmsCodeModel().Update(tx, map[string]interface{}{
		"status":     1,
		"send_times": gorm.Expr("send_times + 1"),
	}, "id = ? and send_times + 1 < 6", smsCode.Id)
	if err != nil {
		return nil, err
	}
	if num == 0 {
		return nil, fmt.Errorf("一小时之内最多发送五次验证码")
	}
	err = SendEmail(email, "注册验证码", fmt.Sprintf("【即时通讯】验证码：%s，1小时内有效，为了保障您的账户安全，请勿向他人泄漏验证码信息。", code))
	if err != nil {
		return nil, err
	}
	tx.Commit()
	return map[string]interface{}{}, nil
}

// @desc SendEmail body支持html格式字符串
// @auth liuguoqiang 2020-11-30
// @param
// @return
func SendEmail(email, subject, body string) error {
	m := gomail.NewMessage()

	m.SetHeader("From", "2276282419@qq.com")
	m.SetHeader("To", email)
	//m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	//m.Attach("/home/Alex/lolcat.jpg")
	password := os.Getenv("EMAIL_PASSWORD")
	client := gomail.NewDialer("smtp.qq.com", 587, "2276282419@qq.com", password)
	if err := client.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

// @desc 获取群成员
// @auth liuguoqiang 2020-11-20
// @param
// @return
func ChatInit(uid int64) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	user, err := model.AuthHandler.GetOne("id = ?", uid)
	if err != nil {
		return nil, err
	}
	resp["mine"] = map[string]interface{}{
		"username": user.Username,
		"id":       user.Id,
		"status":   "online",
		"sign":     "我的个性签名",
		"avatar":   user.Avatar,
	}
	friend := make(map[string]interface{})
	friend["groupname"] = "默认分组"
	friend["id"] = 1
	friendList := make([]map[string]interface{}, 0)
	friends, err := model.FriendHandler.GetList("uid = ?", uid)
	uids := make([]int64, 0)
	for k1 := range friends {
		v1 := friends[k1]
		uids = append(uids, v1.FriendUid)
	}
	if len(uids) > 0 {
		userList, err := model.AuthHandler.GetList("id in(?)", uids)
		if err != nil {
			return nil, err
		}
		for k1 := range userList {
			v1 := userList[k1]
			obj := make(map[string]interface{})
			obj["username"] = v1.Username
			obj["id"] = v1.Id
			obj["avatar"] = v1.Avatar
			obj["status"] = "online"
			obj["sign"] = v1.Username + "的个性签名"
			friendList = append(friendList, obj)
		}
	}
	friend["list"] = friendList
	resp["friend"] = []map[string]interface{}{friend}
	userGroupList, err := model.UserGroupHandler.GetList("uid = ?", uid)
	if err != nil {
		return nil, err
	}
	groupids := make([]int64, 0)
	for k1 := range userGroupList {
		groupids = append(groupids, userGroupList[k1].GroupId)
	}
	groupArr := make([]map[string]interface{}, 0)
	if len(groupids) > 0 {
		groupList, err := model.GroupHandler.GetList("id in(?)", groupids)
		if err != nil {
			return nil, err
		}
		for k1 := range groupList {
			v1 := groupList[k1]
			groupArr = append(groupArr, map[string]interface{}{
				"groupname": v1.GroupName + fmt.Sprintf("(ID:%d)", v1.Id+9527),
				"id":        v1.Id,
				"avatar":    v1.Avatar,
			})
		}
	}
	resp["group"] = groupArr
	return resp, nil
}

// @desc 获取群成员
// @auth liuguoqiang 2020-11-20
// @param
// @return
func GetGroupMembers(groupId int64) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	userGroupList, err := model.UserGroupHandler.GetList("group_id = ?", groupId)
	if err != nil {
		return nil, err
	}
	uids := make([]int64, 0)
	for k1 := range userGroupList {
		uids = append(uids, userGroupList[k1].Uid)
	}
	list := make([]map[string]interface{}, 0)
	if len(uids) > 0 {
		userList, err := model.AuthHandler.GetList("id in(?)", uids)
		if err != nil {
			return nil, err
		}
		for k1 := range userList {
			v1 := userList[k1]
			obj := make(map[string]interface{})
			obj["username"] = v1.Username
			obj["id"] = strconv.FormatInt(v1.Id, 10)
			obj["avatar"] = v1.Avatar
			obj["sign"] = v1.Username + "的个性签名"
			list = append(list, obj)
		}
	}
	resp["list"] = list
	return resp, nil
}

// @desc 搜索
// @auth liuguoqiang 2020-11-20
// @param
// @return
func Search(keyword, searchType string) (map[string]interface{}, error) {
	if searchType == "email" {
		user, err := model.AuthHandler.GetOne("email = ?", keyword)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"uid":    user.Id,
			"name":   user.Username,
			"email":  user.Email,
			"avatar": user.Avatar,
		}, nil
	} else {
		groupId := cast.ToInt64(keyword) - 9527
		group, err := model.GroupHandler.GetOne("id = ?", groupId)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"group_id": group.Id,
			"name":     group.GroupName,
			"avatar":   group.Avatar,
		}, nil
	}
}

// @desc 加好友
// @auth liuguoqiang 2020-11-20
// @param
// @return
func AddFriend(uid, friendUid int64, remark string) (map[string]interface{}, error) {
	_, err := model.FriendHandler.GetOne("uid = ? and friend_uid = ?", uid, friendUid)
	if err == nil {
		return nil, fmt.Errorf("您和对方已经是好友")
	}
	user, err := model.AuthHandler.GetOne("id = ?", uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	friend, err := model.AuthHandler.GetOne("id = ?", friendUid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	_, err = model.SystemMsgModel().GetOne("receive_uid = ? and send_uid = ? and status = 0 and msg_type = 1", friendUid, uid)
	if err == nil {
		err = model.SystemMsgModel().Update(nil, map[string]interface{}{
			"content": remark,
		}, "receive_uid = ? and send_uid = ? and status = 0  and msg_type = 1", friendUid, uid)
		if err != nil {
			return nil, fmt.Errorf("操作失败")
		}
		return nil, nil
	}
	sendMsg := model.SystemMsg{
		NotifyUid:  cast.ToString(uid),
		Username:   friend.Username,
		Avatar:     friend.Avatar,
		ReceiveUid: cast.ToString(friendUid),
		MsgType:    "1",
		SendUid:    cast.ToString(uid),
		Content:    remark,
		Status:     0,
	}
	err = model.SystemMsgModel().Insert(nil, &sendMsg)
	if err != nil {
		return nil, err
	}
	err = model.SystemMsgModel().Insert(nil, &model.SystemMsg{
		NotifyUid:  cast.ToString(friendUid),
		SendMsgId:  sendMsg.Id,
		Username:   user.Username,
		Avatar:     user.Avatar,
		ReceiveUid: cast.ToString(friendUid),
		MsgType:    "1",
		SendUid:    cast.ToString(uid),
		Content:    remark,
		Status:     0,
	})
	if err != nil {
		return nil, err
	}

	err = common.NodeINfo1.SendToClientId(cast.ToString(friendUid), map[string]interface{}{
		"msg_type": "add_friend",
		"data": map[string]interface{}{
			"friend_uid": uid,
		},
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// @desc 管理加好友申请
// @auth liuguoqiang 2020-11-20
// @param
// @return
func ManageAddFriend(uid, id, status int64) (map[string]interface{}, error) {
	systemMsg, err := model.SystemMsgModel().GetOne("id = ? and notify_uid = ? and receive_uid = ? and msg_type = 1", id, uid, uid)
	if err != nil {
		return nil, fmt.Errorf("请求不存在")
	}
	_, err = model.FriendModel().GetOne("uid = ? and friend_uid = ?", systemMsg.ReceiveUid, systemMsg.SendUid)
	if err == nil {
		return nil, fmt.Errorf("您和对方已经是好友")
	}
	if status == 1 {
		err = model.FriendHandler.Insert(nil, &model.Friend{
			Uid:       cast.ToInt64(systemMsg.ReceiveUid),
			FriendUid: cast.ToInt64(systemMsg.SendUid),
		})
		if err != nil {
			return nil, err
		}
		err = model.FriendHandler.Insert(nil, &model.Friend{
			Uid:       cast.ToInt64(systemMsg.SendUid),
			FriendUid: cast.ToInt64(systemMsg.ReceiveUid),
		})
		if err != nil {
			return nil, err
		}
	}
	err = model.SystemMsgModel().Update(nil, map[string]interface{}{
		"status": status,
	}, "id = ? and notify_uid = ?", id, uid)
	if err != nil {
		return nil, err
	}
	err = model.SystemMsgModel().Update(nil, map[string]interface{}{
		"status": status,
	}, "id = ?", systemMsg.SendMsgId)
	if err != nil {
		return nil, err
	}
	err = common.NodeINfo1.SendToClientId(cast.ToString(uid), map[string]interface{}{
		"msg_type": "manage_add_friend",
		"data": map[string]interface{}{
			"uid":        uid,
			"friend_uid": systemMsg.SendUid,
		},
	})
	if err != nil {
		logx.Info(err)
	}
	err = common.NodeINfo1.SendToClientId(cast.ToString(systemMsg.SendUid), map[string]interface{}{
		"msg_type": "manage_add_friend",
		"data": map[string]interface{}{
			"uid":        uid,
			"friend_uid": uid,
		},
	})
	if err != nil {
		logx.Info(err)
	}

	return nil, nil
}

// @desc 加入群聊
// @auth liuguoqiang 2020-11-20
// @param
// @return
func JoinGroup(uid, groupId int64, remark string) (map[string]interface{}, error) {
	_, err := model.UserGroupHandler.GetOne("uid = ? and group_id = ?", uid, groupId)
	if err == nil {
		return nil, fmt.Errorf("已加入群里")
	}
	user, err := model.AuthHandler.GetOne("id = ?", uid)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	group, err := model.GroupHandler.GetOne("id = ?", groupId)
	if err != nil {
		return nil, fmt.Errorf("群不存在")
	}
	_, err = model.SystemMsgModel().GetOne("group_id = ? and send_uid = ? and status = 0 and msg_type = 2", groupId, uid)
	if err == nil {
		err = model.SystemMsgModel().Update(nil, map[string]interface{}{
			"content": remark,
		}, "group_id = ? and send_uid = ? and status = 0  and msg_type = 2", groupId, uid)
		if err != nil {
			return nil, fmt.Errorf("操作失败")
		}
		return nil, nil
	}
	systemMsg := model.SystemMsg{
		NotifyUid: cast.ToString(uid),
		Username:  group.GroupName,
		Avatar:    group.Avatar,
		GroupId:   cast.ToString(groupId),
		MsgType:   "2",
		SendUid:   cast.ToString(uid),
		Content:   remark,
		Status:    0,
	}
	err = model.SystemMsgModel().Insert(nil, &systemMsg)
	if err != nil {
		return nil, err
	}

	err = model.SystemMsgModel().Insert(nil, &model.SystemMsg{
		NotifyUid: cast.ToString(group.Uid),
		SendMsgId: systemMsg.Id,
		Username:  user.Username,
		Avatar:    user.Avatar,
		GroupId:   cast.ToString(groupId),
		MsgType:   "2",
		SendUid:   cast.ToString(uid),
		Content:   remark,
		Status:    0,
	})
	if err != nil {
		return nil, err
	}

	err = common.NodeINfo1.SendToClientId(cast.ToString(group.Uid), map[string]interface{}{
		"msg_type": "join_group",
		"data": map[string]interface{}{
			"group_id": groupId,
		},
	})
	if err != nil {
		logx.Info(err)
	}
	return nil, nil
}

// @desc 管理加群
// @auth liuguoqiang 2020-11-20
// @param
// @return
func ManageJoinGroup(uid, id, status int64) (map[string]interface{}, error) {
	systemMsg, err := model.SystemMsgModel().GetOne("id = ? and notify_uid = ? and msg_type = 2", id, uid)
	if err != nil {
		return nil, fmt.Errorf("请求不存在")
	}
	_, err = model.UserGroupHandler.GetOne("uid = ? and group_id = ?", cast.ToInt64(systemMsg.SendUid), cast.ToInt64(systemMsg.GroupId))
	if err == nil {
		return nil, fmt.Errorf("您已经加入该群")
	}
	if status == 1 {
		err = model.UserGroupHandler.Insert(nil, &model.UserGroup{
			Uid:     cast.ToInt64(systemMsg.SendUid),
			GroupId: cast.ToInt64(systemMsg.GroupId),
		})
		if err != nil {
			return nil, err
		}
	}
	err = model.SystemMsgModel().Update(nil, map[string]interface{}{
		"status": status,
	}, "id = ? and notify_uid = ?", id, uid)
	if err != nil {
		return nil, err
	}
	err = model.SystemMsgModel().Update(nil, map[string]interface{}{
		"status": status,
	}, "id = ?", systemMsg.SendMsgId)
	if err != nil {
		return nil, err
	}
	err = common.NodeINfo1.SendToClientId(cast.ToString(uid), map[string]interface{}{
		"msg_type": "manage_add_group",
		"data": map[string]interface{}{
			"uid":        uid,
			"friend_uid": systemMsg.SendUid,
		},
	})
	if err != nil {
		logx.Info(err)
	}
	err = common.NodeINfo1.SendToClientId(cast.ToString(systemMsg.SendUid), map[string]interface{}{
		"msg_type": "manage_add_group",
		"data": map[string]interface{}{
			"uid":        uid,
			"friend_uid": uid,
		},
	})
	if err != nil {
		logx.Info(err)
	}

	return nil, nil
}

// @desc 建群
// @auth liuguoqiang 2020-11-20
// @param
// @return
func CreateGroup(uid int64, groupName, avatar string) (map[string]interface{}, error) {
	group := model.Group{
		Uid:       uid,
		GroupName: groupName,
		Avatar:    avatar,
	}
	err := model.GroupHandler.Insert(nil, &group)
	if err != nil {
		return nil, err
	}
	err = model.UserGroupHandler.Insert(nil, &model.UserGroup{
		Uid:     uid,
		GroupId: group.Id,
	})
	if err != nil {
		return nil, err
	}

	err = common.NodeINfo1.SendToClientId(cast.ToString(uid), map[string]interface{}{
		"msg_type": "create_group",
		"data": map[string]interface{}{
			"type":      "group",
			"avatar":    group.Avatar,
			"groupname": group.GroupName,
			"id":        group.Id,
		},
	})
	if err != nil {
		logx.Info(err)
	}
	return map[string]interface{}{}, nil
}

// @desc 获取系统消息
// @auth liuguoqiang 2020-11-20
// @param
// @return
func GetSystemMsg(uid int64, lastSystemMsgId int64, sort string) (map[string]interface{}, error) {
	list := make([]map[string]interface{}, 0)
	resp := map[string]interface{}{
		"list":  list,
		"pages": 1,
	}

	var msgList []*model.SystemMsg
	var err error
	where := "notify_uid = ?"
	args := []interface{}{uid}
	if sort == "desc" {
		if lastSystemMsgId > 0 {
			where += " and id < ?"
			args = append(args, lastSystemMsgId)
		}
		msgList, err = model.SystemMsgModel().GetListPage(50, "id desc", where, args...)
	} else {
		//只拉取最新50条
		msgList, err = model.SystemMsgModel().GetListPage(50, "id asc", "notify_uid = ? and id <= (select max(id) from "+model.SystemMsgModel().TableName()+" where notify_uid = ?) and id > ?", uid, uid, lastSystemMsgId)
	}

	if err != nil {
		logx.Info(err)
		return nil, err
	}
	if len(msgList) == 0 {
		return resp, nil
	}

	for k1 := range msgList {
		v1 := msgList[k1]
		obj := map[string]interface{}{
			"id":       v1.Id,
			"username": v1.Username,
			"avatar":   v1.Avatar,
			"send_uid": v1.SendUid,
			// "group_id":   groupId,
			"msg_type":   v1.MsgType,
			"content":    v1.Content,
			"created_at": v1.CreatedAt.Unix(),
			"status":     v1.Status,
		}
		list = append(list, obj)
	}
	resp["list"] = list
	return resp, nil
}
