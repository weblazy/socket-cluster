package stdiolog

import (
	"fmt"
	"github.com/weblazy/socket-cluster/logx"
	"path"
	"runtime"
)

type stdLogger struct {
	dateFormat 	string		// 时间格式
	flow		string		// 操作流 用于标记日志信息
}

func NewLogger() *stdLogger {
	return &stdLogger{
		dateFormat: "2006-01-05 15:04:05.000",
	}
}

func (s *stdLogger) SetDateFormat(temp string) {
	s.dateFormat = temp
}

func (s *stdLogger) SetFlow(flow string) {
	s.dateFormat = flow
}

func (s *stdLogger) Debug(v ...interface{}) {
	panic("implement me")
}

func (s *stdLogger) Debugf(format string, v ...interface{}) {
	panic("implement me")
}

func (s *stdLogger) Info(v ...interface{}) {
	panic("implement me")
}

func (s *stdLogger) Infof(format string, v ...interface{}) {
	panic("implement me")
}

func (s *stdLogger) Warning(v ...interface{}) {
	panic("implement me")
}

func (s *stdLogger) Warningf(format string, v ...interface{}) {
	panic("implement me")
}

func (s *stdLogger) Error(v ...interface{}) {
	panic("implement me")
}

func (s *stdLogger) Errorf(format string, v ...interface{}) {
	panic("implement me")
}

func (s *stdLogger) printMsg(level logx.LogLevel, v ...interface{})  {
	//now := time.Now().Format(s.dateFormat)
	//var levelStr string
	//switch level {
	//case logx.DEBUG:
	//	levelStr = "DEBUG"
	//case logx.INFO:
	//	levelStr = "INFO"
	//case logx.WARNING:
	//	levelStr = "WARNING"
	//case logx.ERROR:
	//	levelStr = "ERROR"
	//}
	//funcName, fileName, lineNo := getCaseLineInfo(3)

}

func getCaseLineInfo(skip int) (funcName, fileName string, lineNo int) {
	pc, file, lineNo, ok := runtime.Caller(skip)
	if !ok {
		fmt.Println("runtime.Caller() failed")
		return
	}
	funcName = path.Base(runtime.FuncForPC(pc).Name())
	fileName = path.Base(file)
	return
}