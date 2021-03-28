package logx_test

import (
	"github.com/weblazy/socket-cluster/logx"
	"testing"
)

func TestStdLogger(t *testing.T) {
	logger := logx.NewStdLogger().SetFlow("std").SetShowLevel(logx.INFO | logx.ERROR)
	logger.Info("info")
	logger.Debug("debug")
	logger.Warning("warning")
	logger.Error("error")

	logger.Infof("infof")
	logger.Debugf("debugf")
	logger.Warningf("warningf")
	logger.Errorf("errorf")
}
