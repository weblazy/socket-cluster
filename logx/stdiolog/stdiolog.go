package stdiolog

import (
	"encoding/json"
	"fmt"
	"path"
	"runtime"
	"time"

	"github.com/weblazy/socket-cluster/logx"
)

type logLevel uint16

const (
	DEBUG logLevel = 1 << iota
	INFO
	WARNING
	ERROR
)
const (
	black = iota + 30
	red
	green
	yellow
	blue
)

type stdLogger struct {
	dateFormat 	string			// 时间格式 默认：2006-01-05 15:04:05.000
	flow		string			// 操作流 用于标记日志信息 不设置无输出
	color		bool			// 颜色输出	不设置无颜色
	level		logLevel		// 输出level 不设置默认全部输出
}

func NewLogger() *stdLogger {
	return &stdLogger{
		dateFormat: "2006-01-05 15:04:05.000",
		color: false,
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

func (s *stdLogger) SetColor(color bool) *stdLogger {
	s.color = color
	return s
}

func (s *stdLogger) ShowLevel(level logLevel) *stdLogger {
	s.level = level
	return s
}

func (s *stdLogger) Debug(v ...interface{}) {
	if s.level == 0 || (s.level & DEBUG != 0) {
		s.printContent(DEBUG, v...)
	}
}

func (s *stdLogger) Debugf(format string, v ...interface{}) {
	if s.level == 0 || (s.level & DEBUG != 0) {
		s.printMsg(DEBUG, format, v...)
	}
}

func (s *stdLogger) Info(v ...interface{}) {
	if s.level == 0 || (s.level & INFO != 0) {
		s.printContent(INFO, v...)
	}
}

func (s *stdLogger) Infof(format string, v ...interface{}) {
	if s.level == 0 || (s.level & INFO != 0) {
		s.printMsg(INFO, format, v...)
	}
}

func (s *stdLogger) Warning(v ...interface{}) {
	if s.level == 0 || (s.level & WARNING != 0) {
		s.printContent(WARNING, v...)
	}
}

func (s *stdLogger) Warningf(format string, v ...interface{}) {
	if s.level == 0 || (s.level & WARNING != 0) {
		s.printMsg(WARNING, format, v...)
	}
}

func (s *stdLogger) Error(v ...interface{}) {
	if s.level == 0 || (s.level & ERROR != 0) {
		s.printContent(ERROR, v...)
	}
}

func (s *stdLogger) Errorf(format string, v ...interface{}) {
	if s.level == 0 || (s.level & ERROR != 0) {
		s.printMsg(ERROR, format, v...)
	}
}

func (s *stdLogger) printContent(level logLevel, v ...interface{})  {
	levelStr := levelToStr(level)
	funcName, fileName, lineNo := getCaseLineInfo(3)
	record := logx.Record{
		TimeStamp: time.Now().Format(s.dateFormat),
		Flow: s.flow,
		Loc: fmt.Sprintf("%s:%d:%s", fileName, lineNo, funcName),
		Level: levelStr,
		Content: v,
	}
	bytes, _ := json.Marshal(record)
	consoleLog(level, string(bytes), s.color)
}

func (s *stdLogger) printMsg(level logLevel, format string, v ...interface{})  {
	levelStr := levelToStr(level)
	funcName, fileName, lineNo := getCaseLineInfo(3)
	record := logx.Record{
		TimeStamp: time.Now().Format(s.dateFormat),
		Flow: s.flow,
		Loc: fmt.Sprintf("%s:%d:%s", fileName, lineNo, funcName),
		Level: levelStr,
		Msg: fmt.Sprintf(format, v...),
	}
	bytes, _ := json.Marshal(record)
	consoleLog(level, string(bytes), s.color)
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

func levelToStr(level logLevel) (levelStr string) {
	switch level {
	case DEBUG:
		levelStr = "DEBUG"
	case INFO:
		levelStr = "INFO"
	case WARNING:
		levelStr = "WARNING"
	case ERROR:
		levelStr = "ERROR"
	}
	return levelStr
}

func consoleLog(level logLevel, log string, color bool) {
	if !color {
		fmt.Println(log)
		return
	}
	colorNum := black
	switch level {
	case DEBUG:
		colorNum = blue
	case INFO:
		colorNum = green
	case WARNING:
		colorNum = yellow
	case ERROR:
		colorNum = red
	}
	log = fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", colorNum, log)
	fmt.Println(log)
}
