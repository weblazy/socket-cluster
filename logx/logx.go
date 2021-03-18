package logx

import (
	"encoding/json"
	"fmt"
	"os"
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

	logger struct {
		dateFormat string   // 时间格式 默认：2006-01-05 15:04:05.000
		flow       string   // 操作流 用于标记日志信息 不设置无输出
		level      logLevel // 输出level 不设置默认全部输出
		output     *os.File
	}
	Record struct {
		TimeStamp string        `json:"timestamp"`         // 日志生成时间
		Level     string        `json:"level"`             // 日志等级
		Flow      string        `json:"flow,omitempty"`    // 操作流
		Loc       string        `json:"loc"`               // 发生位置
		Content   []interface{} `json:"content,omitempty"` // 日志内容
		Msg       string        `json:"msg,omitempty"`     // 日志内容
	}
)

func NewStdLogger() *logger {
	return &logger{
		dateFormat: defaultTimeTemp,
		output:     os.Stdout,
	}
}

func NewFileLogger() *logger {
	file := makeLogFile()
	if file == nil {
		file = os.Stdout
	}
	return &logger{
		dateFormat: defaultTimeTemp,
		output:     file,
	}
}

func (s *logger) SetDateFormat(temp string) *logger {
	s.dateFormat = temp
	return s
}

func (s *logger) SetFlow(flow string) *logger {
	s.flow = flow
	return s
}

func (s *logger) ShowLevel(level logLevel) *logger {
	s.level = level
	return s
}

func (s *logger) Debug(v ...interface{}) {
	if s.level == 0 || (s.level&DEBUG != 0) {
		s.printContent(DEBUG, v...)
	}
}

func (s *logger) Debugf(format string, v ...interface{}) {
	if s.level == 0 || (s.level&DEBUG != 0) {
		s.printMsg(DEBUG, format, v...)
	}
}

func (s *logger) Info(v ...interface{}) {
	if s.level == 0 || (s.level&INFO != 0) {
		s.printContent(INFO, v...)
	}
}

func (s *logger) Infof(format string, v ...interface{}) {
	if s.level == 0 || (s.level&INFO != 0) {
		s.printMsg(INFO, format, v...)
	}
}

func (s *logger) Warning(v ...interface{}) {
	if s.level == 0 || (s.level&WARNING != 0) {
		s.printContent(WARNING, v...)
	}
}

func (s *logger) Warningf(format string, v ...interface{}) {
	if s.level == 0 || (s.level&WARNING != 0) {
		s.printMsg(WARNING, format, v...)
	}
}

func (s *logger) Error(v ...interface{}) {
	if s.level == 0 || (s.level&ERROR != 0) {
		s.printContent(ERROR, v...)
	}
}

func (s *logger) Errorf(format string, v ...interface{}) {
	if s.level == 0 || (s.level&ERROR != 0) {
		s.printMsg(ERROR, format, v...)
	}
}

func (s *logger) printContent(level logLevel, v ...interface{}) {
	levelStr := levelToStr(level)
	funcName, fileName, lineNo := getCaseLineInfo(3)
	record := Record{
		TimeStamp: time.Now().Format(s.dateFormat),
		Flow:      s.flow,
		Loc:       fmt.Sprintf("%s:%d:%s", fileName, lineNo, funcName),
		Level:     levelStr,
		Content:   v,
	}
	bytes, _ := json.Marshal(record)
	s.print(string(bytes))
}

func (s *logger) printMsg(level logLevel, format string, v ...interface{}) {
	levelStr := levelToStr(level)
	funcName, fileName, lineNo := getCaseLineInfo(3)
	record := Record{
		TimeStamp: time.Now().Format(s.dateFormat),
		Flow:      s.flow,
		Loc:       fmt.Sprintf("%s:%d:%s", fileName, lineNo, funcName),
		Level:     levelStr,
		Msg:       fmt.Sprintf(format, v...),
	}
	bytes, _ := json.Marshal(record)
	s.print(string(bytes))
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

func (s *logger) print(log string) {
	if s.output == os.Stdout {
		fmt.Fprintln(s.output, log)
	} else {
		now := time.Now().Format(defaultDateTemp)
		filename := s.output.Name()
		if filename[:10] == now {
			fmt.Fprintln(s.output, log)
		} else {
			s.output.Close()
			file := makeLogFile()
			if file != nil {
				s.output = file
			} else {
				s.output = os.Stdout
			}
		}
	}
}

func makeLogFile() *os.File {
	now := time.Now().Format(defaultDateTemp)
	fileName := now + ".log"
	fmt.Println(fileName)
	file, err := os.OpenFile(fileName,os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("log file open failed, err =", err)
		return nil
	}
	return file
}
