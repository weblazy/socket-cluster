package stdiolog_test

import (
	"github.com/weblazy/socket-cluster/logx/stdiolog"
	"testing"
)


func TestStdioLog(t *testing.T) {

	logStr := "20210318"

	logger := stdiolog.NewLogger().SetFlow("withColor").SetColor(true).ShowLevel(stdiolog.WARNING | stdiolog.ERROR)
	logger.Debug("this is a debug test", 1, 'd')
	logger.Info("this is a info test", 1, 'd')
	logger.Warning("this is a warning test", 1, 'd')
	logger.Error("this is a error test", 1, 'd')

	loggerWithoutColor := stdiolog.NewLogger()
	loggerWithoutColor.Infof("----------------")
	loggerWithoutColor.Infof("++++++++++++++++")

	logger.Debugf("today %s", logStr)
	logger.Infof("today %s", logStr)
	logger.Warningf("today %s", logStr)
	logger.Errorf("today %s", logStr)
}