package api

import (
	"websocket-cluster/examples/domain"

	"github.com/labstack/echo/v4"
)

// @desc 登录
// @auth liuguoqiang 2020-11-20
// @param
// @return
func Login(c echo.Context) error {
	//参数验证绑定
	req, response, err := ParseJson(c)
	if err != nil {
		return response.RetError(err, -1)
	}
	username := req.Get("username").String()
	password := req.Get("password").String()
	resp, err := domain.Login(username, password)
	if err != nil {
		return response.RetError(err, -1)
	}
	return response.RetCustomize(0, resp, "")
}

// @desc 注册
// @auth liuguoqiang 2020-11-20
// @param
// @return
func Register(c echo.Context) error {
	//参数验证绑定
	req, response, err := ParseJson(c)
	if err != nil {
		return response.RetError(err, -1)
	}
	username := req.Get("username").String()
	password := req.Get("password").String()
	confirmPassword := req.Get("confirm_password").String()
	email := req.Get("email").String()
	code := req.Get("code").String()
	resp, err := domain.Register(username, password, confirmPassword, email, code)
	if err != nil {
		return response.RetError(err, -1)
	}
	return response.RetCustomize(0, resp, "")
}

// @desc 发送验证码
// @auth liuguoqiang 2020-11-20
// @param
// @return
func SendSmsCode(c echo.Context) error {
	//参数验证绑定
	req, response, err := ParseJson(c)
	if err != nil {
		return response.RetError(err, -1)
	}
	email := req.Get("email").String()
	resp, err := domain.SendSmsCode(email)
	if err != nil {
		return response.RetError(err, -1)
	}
	return response.RetCustomize(0, resp, "")
}

// @desc 聊天初始化
// @auth liuguoqiang 2020-11-20
// @param
// @return
func ChatInit(c echo.Context) error {
	uid, _, response, err := ParseParams(c)
	if err != nil {
		return response.RetError(err, -1)
	}
	resp, err := domain.ChatInit(uid)
	if err != nil {
		return response.RetError(err, -1)
	}
	return response.RetCustomize(0, resp, "")
}

// @desc 获取群成员
// @auth liuguoqiang 2020-11-20
// @param
// @return
func GetGroupMembers(c echo.Context) error {
	uid, _, response, err := ParseParams(c)
	if err != nil {
		return response.RetError(err, -1)
	}
	resp, err := domain.GetGroupMembers(uid)
	if err != nil {
		return response.RetError(err, -1)
	}
	return response.RetCustomize(0, resp, "")
}

// @desc 搜索
// @auth liuguoqiang 2020-11-20
// @param
// @return
func Search(c echo.Context) error {
	_, req, response, err := ParseParams(c)
	if err != nil {
		return response.RetError(err, -1)
	}
	keyword := req.Get("keyword").String()
	searchType := req.Get("search_type").String()
	resp, err := domain.Search(keyword, searchType)
	if err != nil {
		return response.RetError(err, -1)
	}
	return response.RetCustomize(0, resp, "")
}

// @desc 添加好友
// @auth liuguoqiang 2020-11-20
// @param
// @return
func AddFriend(c echo.Context) error {
	uid, req, response, err := ParseParams(c)
	if err != nil {
		return response.RetError(err, -1)
	}
	friendUid := req.Get("friend_uid").Int()
	remark := req.Get("remark").String()
	resp, err := domain.AddFriend(uid, friendUid, remark)
	if err != nil {
		return response.RetError(err, -1)
	}
	return response.RetCustomize(0, resp, "")
}

// @desc 回复添加好友
// @auth liuguoqiang 2020-11-20
// @param
// @return
func AcceptAddFriend(c echo.Context) error {
	uid, req, response, err := ParseParams(c)
	if err != nil {
		return response.RetError(err, -1)
	}
	friendUid := req.Get("friend_uid").Int()
	resp, err := domain.AcceptAddFriend(uid, friendUid)
	if err != nil {
		return response.RetError(err, -1)
	}
	return response.RetCustomize(0, resp, "")
}

// @desc 添加好友
// @auth liuguoqiang 2020-11-20
// @param
// @return
func JoinGroup(c echo.Context) error {
	uid, req, response, err := ParseParams(c)
	if err != nil {
		return response.RetError(err, -1)
	}
	groupId := req.Get("group_id").Int()
	remark := req.Get("remark").String()
	resp, err := domain.JoinGroup(uid, groupId, remark)
	if err != nil {
		return response.RetError(err, -1)
	}
	return response.RetCustomize(0, resp, "")
}

// @desc 同意加群
// @auth liuguoqiang 2020-11-20
// @param
// @return
func AcceptJoinGroup(c echo.Context) error {
	uid, req, response, err := ParseParams(c)
	if err != nil {
		return response.RetError(err, -1)
	}
	friendUid := req.Get("friend_uid").Int()
	resp, err := domain.AcceptJoinGroup(uid, friendUid)
	if err != nil {
		return response.RetError(err, -1)
	}
	return response.RetCustomize(0, resp, "")
}

// @desc 建群
// @auth liuguoqiang 2020-11-20
// @param
// @return
func CreateGroup(c echo.Context) error {
	uid, req, response, err := ParseParams(c)
	if err != nil {
		return response.RetError(err, -1)
	}
	groupName := req.Get("group_name").String()
	avatar := req.Get("avatar").String()
	resp, err := domain.CreateGroup(uid, groupName, avatar)
	if err != nil {
		return response.RetError(err, -1)
	}
	return response.RetCustomize(0, resp, "")
}
