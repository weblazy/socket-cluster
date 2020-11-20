package api

import (
	"io/ioutil"
	"websocket-cluster/examples/auth"

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

func ParseParams(c echo.Context) (string, gjson.Result, *api.Response, error) {
	request := c.Request()
	response := api.NewResponse(c)
	var jsonParams gjson.Result

	token := request.Header.Get("token")

	id, err := auth.AuthManager.Validate(token)
	if err != nil {
		return id, jsonParams, response, err
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
