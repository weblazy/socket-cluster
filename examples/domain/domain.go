package domain

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/weblazy/socket-cluster/examples/auth"
	"github.com/weblazy/socket-cluster/examples/common"
	"github.com/weblazy/socket-cluster/examples/model"

	"github.com/jinzhu/gorm"
	"github.com/spf13/cast"
	"github.com/weblazy/core/logx"
	"gopkg.in/gomail.v2"
)

type ()

var avatarArr = []string{
	"https://img.sucai999.com/bIS1dEpwM4CqZ%7BN1Nj6vbYCqZz6kc31w%5Bnmt%5BT9zNEJxNUFyOz9zOkh4OUB5PW9yNUR1NEV2OEhxNECgNj6rdHd%3E.jpg?t=?w=300&h=551&webp=1",
	"https://img.sucai999.com/bIS1dEpwM4CqZ%7BN%7BPT6vbYCqZz6kc31w%5Bnmt%5BT9zNEJxNUFxOz9%7BNkF5NUl2O29zNkF3NkByPUFxPESgNj6rdHd%3E.jpg?t=?w=300&h=551&webp=1",
	"https://ss0.bdstatic.com/70cFvHSh_Q1YnxGkpoWK1HF6hhy/it/u=152637735,3689600067&fm=26&gp=0.jpg",
	"https://img.sucai999.com/bIS1dEpwM4CqZ%7BN%7BPT6vbYCqZz6kc31w%5Bnmt%5BT9zNEJxNUFxOz9%7BNkF5NUl2O29xNEFxOUNyPEVxPEmgNj6rdHd%3E.jpg?t=?w=300&h=551&webp=1",
	"https://img.sucai999.com/bIS1dEpwM4CqZ%7BN1Nz6vbYCqZz6kc31w%5Bnmt%5BT9zNEJxNUFyPD9zOUhxO%7BZ5Nm9yOUJ3OUlxNkVxN%7BmgNj6rdHd%3E.jpg?t=?w=300&h=551&webp=1",
	"https://img.sucai999.com/bIS1dEpwM4CqZ%7BNyOT6vbYCqZz6kc31w%5Bnmt%5BT9zNEJxNEhyND9zNUR4NUB%7BN29xPEF3N%7BN2OUJxPEOgNj6rdHd%3E.jpg?t=?w=300&h=551&webp=1",
	"https://img.sucai999.com/bIS1dEpwM4CqZ%7BF1OT6vbYCqZz6kc31w%5Bnmt%5BT9zNEF4NUFxNj9zNEV5OUJxNm9zNEB3NUhzNUF5N%7BSgNj6rdHd%3E.jpg?t=?w=300&h=551&webp=1",
	"https://img.sucai999.com/bIS1dEpwM%7BJyPD5yOj5yNkVvOERweYCtc3GlM%7BJxNUhxOEJyM%7BZ6Z%7BKz%5BHV1fXN2eEOlNkixez6rdHd%3E.jpg?t=?w=300&h=551&webp=1",
	"https://img.sucai999.com/bIS1dEpwM4CqZ%7BN1OD6vbYCqZz6kc31wdHmkM%7BJxNkByNUJzM%7BNyOkJyOEJzY%7BBxNUdyO%7BV1OUBxNG91Mnqx%5Bx%3E%3E.jpg?w=300&h=",
	"https://img.sucai999.com/bIS1dEpwM4CqZ%7BN1OD6vbYCqZz6kc31wdHmkM%7BJxNkByNUJzM%7BNyOkJyOEJzY%7BBxNUNzOUF3O%7BBxNG91Mnqx%5Bx%3E%3E.jpg?w=300&h=",
	"https://img.sucai999.com/bIS1dEpwM4CqZ%7BN1OD6vbYCqZz6kc31wdHmkM%7BJxNkByNUJzM%7BNyOkJyOEJzY%7BBxNEh2PEV6OEBxNG91Mnqx%5Bx%3E%3E.jpg?w=300&h=",
	"https://img.sucai999.com/bIS1dEpwM4CqZ%7BN1ND6vbYCqZz6kc31wdHmkM%7BJxNkByNUFyM%7BJ3PEd2NEh6Y%7BF4OEVzO%7BBxNUBxNG91Mnqx%5Bx%3E%3E.jpg?w=300&h=",
	"https://img.sucai999.com/bIS1dEpwM%7BJyPD5yOj5yNkVvOEBw%5Bnmt%5BYNwNkBzNEF3M%7BSpcU%5BpOnKzb%7BejeH%5Bsd%7Be4PD6rdHd%3E.jpg?w=300&h=",
	"https://img.sucai999.com/bIS1dEpwM%7BJyPD5yOj5yNkVvOEBw%5Bnmt%5BYNwNkBzNEF3M36scXOi%5Bol4%5Bkh%7BdEO1N311bT6rdHd%3E.jpg?w=300&h=",
	"https://img.sucai999.com/bIS1dEpwM%7BJyPD5yOj5yNkVvOEBw%5Bnmt%5BYNwNkBzNEF3M4isdH12fIq%7BdEirfU%5B6fUK7cj6rdHd%3E.jpg?w=300&h=",
	"https://img.sucai999.com/bIS1dEpwM%7BJyPD5yOj5yNkVvOEBw%5Bnmt%5BYNwNkBzNEF3M3ixcXKifkOrfoikfYB5d4evZz6rdHd%3E.jpg?w=300&h=",
	"https://img.sucai999.com/bIS1dEpwM%7BJyPD5yOj5yNkVvOEBw%5Bnmt%5BYNwNkBzNEF3M4N%7BOH54bXJ4eIR1N4ek%5BUW7fD6rdHd%3E.jpg?w=300&h=",
	"https://img.sucai999.com/bIS1dEpwM%7BJyPD5yOj5yNkVvOEBw%5Bnmt%5BYNwNkBzNEF3M3J2bISuO4pzOop1Z4epeIq4bT6rdHd%3E.jpg?w=300&h=",
	"https://img.sucai999.com/bIS1dEpwM%7BJyPD5yOj5yNkVvOEBw%5Bnmt%5BYNwNkBzNEF3M4SjeHV2fnG4bol5NnGr%5BHK7fj6rdHd%3E.jpg?w=300&h=",
	"https://img.sucai999.com/bIS1dEpwM%7BJyPD5yOj5yNkVvOEBw%5Bnmt%5BYNwNkBzNEF3M4qz%5BnK%7BfH67OoixbESjO4m7fD6rdHd%3E.jpg?w=300&h=",
	"https://img.sucai999.com/bIS1dEpwM%7BJyPD5yOj5yNkVvOEBw%5Bnmt%5BYNwNkBzNEF3M%7B%5BrbYd%7BOYq5OUisbkR%7BN3mxNz6rdHd%3E.jpg?w=300&h=",
	"https://img.sucai999.com/bIS1dEpwM%7BJyPD5yOj5yNkVvOEBw%5Bnmt%5BYNwNkByPUFzNklw%5BE%5Bme3KnZn12Nkiscn2xO4OpMnqx%5Bx%3E%3E.jpg?w=300&h=",
	"https://img.sucai999.com/bIS1dEpwM%7BJyPD5yOj5yNkVvOEBw%5Bnmt%5BYNwNkByPUFzNkRwfYN2bYOlb3mmNnu6dn2sbXt5Mnqx%5Bx%3E%3E.jpg?w=300&h=",
	"https://img.sucai999.com/bIS1dEpwM%7BJyPD5yOj5yNkVvOEBw%5Bnmt%5BYNwNkBzNEF3M3%5BvOYe4N4B%7BbIin%5BIOiZUK%7BNz6rdHd%3E.jpg?w=300&h=",
}

// Login
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

// Register
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

	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(avatarArr))

	user := model.Auth{
		Username: username,
		Email:    email,
		Avatar:   avatarArr[index],
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

// SendSmsCode
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

// SendEmail
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

// ChatInit
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

// GetGroupMembers
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

// Search
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

// AddFriend
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
	} else {
		systemMsg := model.SystemMsg{
			NotifyUid:  cast.ToString(uid),
			Username:   friend.Username,
			Avatar:     friend.Avatar,
			ReceiveUid: cast.ToString(friendUid),
			MsgType:    "1",
			SendUid:    cast.ToString(uid),
			Content:    remark,
			Status:     0,
		}
		err = model.SystemMsgModel().Insert(nil, &systemMsg)
		if err != nil {
			return nil, err
		}
		err = model.SystemMsgModel().Insert(nil, &model.SystemMsg{
			NotifyUid:  cast.ToString(friendUid),
			SendMsgId:  systemMsg.Id,
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
	}

	count, err := model.SystemMsgModel().Count("notify_uid = ? and id > ? and send_uid <> ?", friend.Id, friend.MaxReadSystemMsgId, friend.Id)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}
	err = common.NodeINfo1.SendToClientId(cast.ToString(friendUid), map[string]interface{}{
		"msg_type": "add_friend",
		"data": map[string]interface{}{
			"count": count,
		},
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ManageSystemMsg
func ManageSystemMsg(uid, id, status int64) (map[string]interface{}, error) {
	systemMsg, err := model.SystemMsgModel().GetOne("id = ? and notify_uid = ?", id, uid)
	if err != nil {
		return nil, fmt.Errorf("请求不存在")
	}
	if cast.ToInt64(systemMsg.GroupId) > 0 {
		return ManageJoinGroup(uid, id, status)
	}
	return ManageAddFriend(uid, id, status)
}

// ManageAddFriend
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
	}, "send_msg_id = ?", systemMsg.SendMsgId)
	if err != nil {
		return nil, err
	}
	err = model.SystemMsgModel().Update(nil, map[string]interface{}{
		"status": status,
	}, "id = ?", systemMsg.SendMsgId)
	if err != nil {
		return nil, err
	}
	if status == 2 {
		return nil, nil
	}
	sendUser, err := model.AuthModel().GetOne("id = ?", systemMsg.SendUid)
	if err != nil {
		return nil, err
	}
	err = common.NodeINfo1.SendToClientId(cast.ToString(uid), map[string]interface{}{
		"msg_type": "manage_add_friend",
		"data": map[string]interface{}{
			"id":       sendUser.Id,
			"group_id": 1,
			"avatar":   sendUser.Avatar,
			"username": sendUser.Username,
			"sign":     sendUser.Username + "的个性签名",
		},
	})
	if err != nil {
		logx.Info(err)
	}
	receiveUser, err := model.AuthModel().GetOne("id = ?", uid)
	if err != nil {
		return nil, err
	}
	err = common.NodeINfo1.SendToClientId(cast.ToString(systemMsg.SendUid), map[string]interface{}{
		"msg_type": "manage_add_friend",
		"data": map[string]interface{}{
			"id":       receiveUser.Id,
			"group_id": 1,
			"avatar":   receiveUser.Avatar,
			"username": receiveUser.Username,
			"sign":     receiveUser.Username + "的个性签名",
		},
	})
	if err != nil {
		logx.Info(err)
	}

	return nil, nil
}

// JoinGroup
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

	} else {
		systemMsg := model.SystemMsg{
			NotifyUid: cast.ToString(uid),
			Username:  group.GroupName,
			Avatar:    group.Avatar,
			GroupId:   cast.ToString(groupId),
			GroupName: group.GroupName,
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
			GroupName: group.GroupName,
			MsgType:   "2",
			SendUid:   cast.ToString(uid),
			Content:   remark,
			Status:    0,
		})
		if err != nil {
			return nil, err
		}
	}

	notifyUser, err := model.AuthModel().GetOne("id = ?", group.Uid)
	if err != nil {
		return nil, err
	}
	count, err := model.SystemMsgModel().Count("notify_uid = ? and id > ? and send_uid <> ?", group.Uid, notifyUser.MaxReadSystemMsgId, group.Uid)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}
	err = common.NodeINfo1.SendToClientId(cast.ToString(group.Uid), map[string]interface{}{
		"msg_type": "join_group",
		"data": map[string]interface{}{
			"count": count,
		},
	})
	if err != nil {
		logx.Info(err)
	}
	return nil, nil
}

// ManageJoinGroup
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
	}, "send_msg_id = ?", systemMsg.SendMsgId)
	if err != nil {
		return nil, err
	}
	err = model.SystemMsgModel().Update(nil, map[string]interface{}{
		"status": status,
	}, "id = ?", systemMsg.SendMsgId)
	if err != nil {
		return nil, err
	}
	if status == 2 {
		return nil, nil
	}
	group, err := model.GroupModel().GetOne("id = ?", systemMsg.GroupId)
	if err != nil {
		return nil, err
	}
	err = common.NodeINfo1.SendToClientId(cast.ToString(systemMsg.SendUid), map[string]interface{}{
		"msg_type": "manage_join_group",
		"data": map[string]interface{}{
			"id":         group.Id,
			"group_name": group.GroupName + fmt.Sprintf("(ID:%d)", group.Id+9527),
			"avatar":     group.Avatar,
		},
	})
	if err != nil {
		logx.Info(err)
	}

	return nil, nil
}

// CreateGroup
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
			"groupname": group.GroupName + fmt.Sprintf("(ID:%d)", group.Id+9527),
			"id":        group.Id,
		},
	})
	if err != nil {
		logx.Info(err)
	}
	return map[string]interface{}{}, nil
}

// GetSystemMsg
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
	var maxMsgId int64
	for k1 := range msgList {
		v1 := msgList[k1]
		obj := map[string]interface{}{
			"id":         v1.Id,
			"username":   v1.Username,
			"avatar":     v1.Avatar,
			"send_uid":   v1.SendUid,
			"group_id":   v1.GroupId,
			"group_name": v1.GroupName + fmt.Sprintf("(ID:%d)", v1.Id+9527),
			"msg_type":   v1.MsgType,
			"content":    v1.Content,
			"created_at": v1.CreatedAt.Unix(),
			"status":     v1.Status,
		}
		if maxMsgId < v1.Id && cast.ToInt64(v1.SendUid) != uid {
			maxMsgId = v1.Id
		}
		list = append(list, obj)
	}
	num, err := model.AuthModel().Update(nil, map[string]interface{}{
		"max_read_system_msg_id": maxMsgId,
	}, "id = ? and max_read_system_msg_id < ?", uid, maxMsgId)
	if err != nil {
		logx.Info(err)
		return nil, err
	}
	if num > 0 {
		err = common.NodeINfo1.SendToClientId(cast.ToString(uid), map[string]interface{}{
			"msg_type": "read_system_msg",
			"data": map[string]interface{}{
				"system_msg_id": maxMsgId,
				"count":         0,
			},
		})
		if err != nil {
			logx.Info(err)
		}
	}
	resp["list"] = list
	return resp, nil
}
