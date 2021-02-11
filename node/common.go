package node

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

type (
	Auth struct {
		TransAddress string `json:"trans_address"` //Node address ip:port
		Password     string `json:"password"`      //Password for auth when node connect on
	}

	SocketConfig struct {
		Ip   string
		Port int64
	}
)

var (
	authTime = 10 * time.Second
	upgrader = websocket.Upgrader{
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	defaultMasterPort int64 = 9527
)

// @desc 生成唯一UUD 用去除-后的系统UUID 32位
// @auth liuguoqiang 2020-11-23
// @param
// @return
func GetUUID() string {
	uuId := uuid.NewV4().String()
	uuIdStr := strings.Replace(uuId, "-", "", -1)
	return uuIdStr
}
