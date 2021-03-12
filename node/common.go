package node

import (
	"strings"

	uuid "github.com/satori/go.uuid"
)

// GetUUID
func GetUUID() string {
	uuId := uuid.NewV4().String()
	uuIdStr := strings.Replace(uuId, "-", "", -1)
	return uuIdStr
}
