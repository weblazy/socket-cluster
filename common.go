package websocket_cluster

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
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
