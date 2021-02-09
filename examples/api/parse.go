package api

import (
	"fmt"
	"io/ioutil"
	"socket-cluster/examples/auth"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/sunmi-OS/gocore/api"
	"github.com/tidwall/gjson"
)

func ParseJson(c echo.Context) (gjson.Result, *api.Response, error) {
	request := c.Request()
	response := api.NewResponse(c)
	var jsonParams gjson.Result
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return jsonParams, response, err
	}
	c.Set("requestBody", string(body))
	if err != nil {
		return jsonParams, response, err
	}
	jsonParams = gjson.Parse(string(body))
	return jsonParams, response, nil
}

func ParseParams(c echo.Context) (int64, gjson.Result, *api.Response, error) {
	request := c.Request()
	response := api.NewResponse(c)

	var jsonParams gjson.Result

	token := request.Header.Get("token")
	fmt.Printf("%#v\n", token)
	idStr, err := auth.AuthManager.Validate(token)
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return id, jsonParams, response, fmt.Errorf("token验证失败")
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return id, jsonParams, response, err
	}

	c.Set("requestBody", string(body))
	if err != nil {
		return id, jsonParams, response, err
	}
	jsonParams = gjson.Parse(string(body))
	return id, jsonParams, response, nil
}
