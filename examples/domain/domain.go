package domain

import (
	"strconv"
	"websocket-cluster/examples/auth"
	"websocket-cluster/examples/model"
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
		"email":    user.Username + "@qq.com",
		"avatar":   user.Avatar,
		"token":    token,
	}, nil
}

// @desc 注册
// @auth liuguoqiang 2020-11-20
// @param
// @return
func Register(username, password, confirmPassword, email, code string) (map[string]interface{}, error) {
	user, err := model.AuthHandler.GetOne("email = ?", email)
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
		"email":    user.Username + "@qq.com",
		"avatar":   user.Avatar,
		"token":    token,
	}, nil
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
