package logx_test

import (
	"sync"
	"testing"

	"github.com/weblazy/socket-cluster/logx"
)

func TestStdoutLog(t *testing.T) {
	logger := logx.NewStdLogger()
	logger.Debug("this is a debug log")
	logger.Info("this is a info log")
	logger.Warning("this is a warning log")
	logger.Error("this is a error log")

	logger.Debugf("this is a debug log %d", 1)
	logger.Infof("this is a info log %d", 2)
	logger.Warningf("this is a warning log %d", 3)
	logger.Errorf("this is a error log %d", 4)
}

func TestFileLog(t *testing.T) {
	logger := logx.NewFileLogger().SetFlow("fileLog").ShowLevel(logx.INFO | logx.WARNING)
	logger.Debug("this is a debug log")
	logger.Info("this is a info log")
	logger.Warning("this is a warning log")
	logger.Error("this is a error log")

	logger.Debugf("this is a debug log %d", 1)
	logger.Infof("this is a info log %d", 2)
	logger.Warningf("this is a warning log %d", 3)
	logger.Errorf("this is a error log %d", 4)
}

func TestMultithreadingLog(t *testing.T) {
	logger := logx.NewFileLogger().SetFlow("fileLog").ShowLevel(logx.INFO | logx.WARNING)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func ()  {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				logger.Infof("this is a %d log %d", i, j)
			}
			
		}()
	}
	wg.Wait()
}