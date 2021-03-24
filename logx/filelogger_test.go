package logx_test

import (
	"testing"
	"time"

	"github.com/weblazy/socket-cluster/logx"
)

func TestFileLog(t *testing.T) {
	logger := logx.NewFileLogger()
	for {
		logger.Debug("debug")
		time.Sleep(2 * time.Second)
	}
}