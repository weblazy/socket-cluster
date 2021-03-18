package api

import (
	"github.com/weblazy/socket-cluster/examples/domain"

	"github.com/labstack/echo/v4"
)

// Login
func Login(c echo.Context) error {
	// Argument parsing
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

// Register
func Register(c echo.Context) error {
	// Argument parsing
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

// SendSmsCode
func SendSmsCode(c echo.Context) error {
	// Argument parsing
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

// ChatInit
func ChatInit(c echo.Context) error {
	// Argument parsing
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

// GetGroupMembers
func GetGroupMembers(c echo.Context) error {
	// Argument parsing
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

// Search
func Search(c echo.Context) error {
	// Argument parsing
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

// AddFriend
func AddFriend(c echo.Context) error {
	// Argument parsing
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

// ManageSystemMsg
func ManageSystemMsg(c echo.Context) error {
	// Argument parsing
	uid, req, response, err := ParseParams(c)
	if err != nil {
		return response.RetError(err, -1)
	}
	id := req.Get("id").Int()
	status := req.Get("status").Int()
	resp, err := domain.ManageSystemMsg(uid, id, status)
	if err != nil {
		return response.RetError(err, -1)
	}
	return response.RetCustomize(0, resp, "")
}

// ManageAddFriend
func ManageAddFriend(c echo.Context) error {
	// Argument parsing
	uid, req, response, err := ParseParams(c)
	if err != nil {
		return response.RetError(err, -1)
	}
	id := req.Get("id").Int()
	status := req.Get("status").Int()
	resp, err := domain.ManageAddFriend(uid, id, status)
	if err != nil {
		return response.RetError(err, -1)
	}
	return response.RetCustomize(0, resp, "")
}

// JoinGroup
func JoinGroup(c echo.Context) error {
	// Argument parsing
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

// ManageJoinGroup
func ManageJoinGroup(c echo.Context) error {
	// Argument parsing
	uid, req, response, err := ParseParams(c)
	if err != nil {
		return response.RetError(err, -1)
	}
	id := req.Get("id").Int()
	status := req.Get("status").Int()
	resp, err := domain.ManageJoinGroup(uid, id, status)
	if err != nil {
		return response.RetError(err, -1)
	}
	return response.RetCustomize(0, resp, "")
}

// CreateGroup
func CreateGroup(c echo.Context) error {
	// Argument parsing
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

// GetSystemMsg
func GetSystemMsg(c echo.Context) error {
	// Argument parsing
	uid, req, response, err := ParseParams(c)
	if err != nil {
		return response.RetError(err, -1)
	}
	lastSystemMsgId := req.Get("last_system_msg_id").Int()
	sort := req.Get("sort").String()
	resp, err := domain.GetSystemMsg(uid, lastSystemMsgId, sort)
	if err != nil {
		return response.RetError(err, -1)
	}
	return response.RetCustomize(0, resp, "")
}
