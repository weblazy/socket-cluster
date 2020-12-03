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
	friends, err := model.FriendHandler.GetList("uid = ? or friend_uid = ?", uid, uid)
	uids := make([]int64, 0)
	for k1 := range friends {
		v1 := friends[k1]
		uids = append(uids, v1.Uid)
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
				"groupname": v1.GroupName,
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
func AddFriends(uid, friendUid, msgContent string) (map[string]interface{}, error) {
	return nil, nil
}

// @desc 接受加好友申请
// @auth liuguoqiang 2020-11-20
// @param
// @return
func AcceptAddFriends(keyword, searchType string) (map[string]interface{}, error) {
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

// @desc 加入群聊
// @auth liuguoqiang 2020-11-20
// @param
// @return
func JoinGroup(uid, groupId, msgContent string) (map[string]interface{}, error) {
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
	err = common.NodeINfo1.SendToUid(cast.ToString(uid), map[string]interface{}{
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
