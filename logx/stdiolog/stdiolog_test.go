package stdiolog_test

import (
	"testing"
	"github.com/weblazy/socket-cluster/logx/stdiolog"
)


func TestStdioLog(t *testing.T) {
	logger := stdiolog.NewLogger().SetFlow("testFlow")
	logger.Debug("this is a debug test", 1, 'd')
}