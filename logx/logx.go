package logx

import (
	"encoding/json"
	"fmt"
	"path"
	"runtime"
	"time"
)

const (
	DEBUG logLevel = 1 << iota
	INFO
	WARNING
	ERROR

	defaultTimeTemp = "2006-01-02 15:04:05.000"
	defaultDateTemp = "2006-01-02"
)

type (
	logLevel uint16
	Logx     interface {
		Debug(v ...interface{})
		Debugf(format string, v ...interface{})
		Info(v ...interface{})
		Infof(format string, v ...interface{})
		Warning(v ...interface{})
		Warningf(format string, v ...interface{})
		Error(v ...interface{})
		Errorf(format string, v ...interface{})
	}

	record struct {
		TimeStamp string        `json:"timestamp"`         // 日志生成时间
		Level     string        `json:"level"`             // 日志等级
		Flow      string        `json:"flow,omitempty"`    // 操作流
		Loc       string        `json:"loc"`               // 发生位置
		Content   []interface{} `json:"content,omitempty"` // 日志内容
		Msg       string        `json:"msg,omitempty"`     // 日志内容
	}
)

func makeContent(dateFormat, flow string, level logLevel, v ...interface{}) string {
	levelStr := levelToStr(level)
	funcName, fileName, lineNo := getCaseLineInfo(3)
	record := record{
		TimeStamp: time.Now().Format(dateFormat),
		Flow:      flow,
		Loc:       fmt.Sprintf("%s:%d:%s", fileName, lineNo, funcName),
		Level:     levelStr,
		Content:   v,
	}
	bytes, _ := json.Marshal(record)
	return string(bytes)
}

func makeMsg(dateFormat, flow string, level logLevel, format string, v ...interface{}) string {
	levelStr := levelToStr(level)
	funcName, fileName, lineNo := getCaseLineInfo(3)
	record := record{
		TimeStamp: time.Now().Format(dateFormat),
		Flow:      flow,
		Loc:       fmt.Sprintf("%s:%d:%s", fileName, lineNo, funcName),
		Level:     levelStr,
		Msg:       fmt.Sprintf(format, v...),
	}
	bytes, _ := json.Marshal(record)
	return string(bytes)
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


