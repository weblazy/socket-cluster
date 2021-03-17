package stdiolog

import (
	"encoding/json"
	"fmt"
	"path"
	"runtime"
	"time"

	"github.com/weblazy/socket-cluster/logx"
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

func (s *stdLogger) SetDateFormat(temp string) *stdLogger {
	s.dateFormat = temp
	return s
}

func (s *stdLogger) SetFlow(flow string) *stdLogger {
	s.flow = flow
	return s
}

func (s *stdLogger) Debug(v ...interface{}) {
	s.printMsg(logx.DEBUG, v...)
}

func (s *stdLogger) Debugf(format string, v ...interface{}) {
	panic("implement me")
}

func (s *stdLogger) Info(v ...interface{}) {
	s.printMsg(logx.INFO, v...)
}

func (s *stdLogger) Infof(format string, v ...interface{}) {
	panic("implement me")
}

func (s *stdLogger) Warning(v ...interface{}) {
	s.printMsg(logx.WARNING, v...)
}

func (s *stdLogger) Warningf(format string, v ...interface{}) {
	panic("implement me")
}

func (s *stdLogger) Error(v ...interface{}) {
	s.printMsg(logx.ERROR, v...)
}

func (s *stdLogger) Errorf(format string, v ...interface{}) {
	panic("implement me")
}

func (s *stdLogger) printMsg(level logx.LogLevel, v ...interface{})  {
	var levelStr string
	switch level {
	case logx.DEBUG:
		levelStr = "DEBUG"
	case logx.INFO:
		levelStr = "INFO"
	case logx.WARNING:
		levelStr = "WARNING"
	case logx.ERROR:
		levelStr = "ERROR"
	}
	funcName, fileName, lineNo := getCaseLineInfo(3)
	record := logx.Record{
		TimeStamp: time.Now().Format(s.dateFormat),
		Flow: s.flow,
		Loc: fmt.Sprintf("%s:%d:%s", fileName, lineNo, funcName),
		Level: levelStr,
		Content: v,
	}
	bytes, _ := json.Marshal(record)
	fmt.Println(string(bytes))
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