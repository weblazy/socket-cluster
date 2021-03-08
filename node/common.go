package node

import (
	"strings"
	"time"

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
	authTime                = 10 * time.Second
	defaultMasterPort int64 = 9527
)

// GetUUID
func GetUUID() string {
	uuId := uuid.NewV4().String()
	uuIdStr := strings.Replace(uuId, "-", "", -1)
	return uuIdStr
}
