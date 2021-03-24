package logx

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type fileLogger struct {
	dateFormat 		string   		// 时间格式 默认：2006-01-05 15:04:05.000
	flow       		string   		// 操作流 用于标记日志信息 不设置无输出
	level      		logLevel 		// 输出level 不设置默认全部输出
	output			*os.File		
	expiredTime 	int				// 清理时间 <0 不清理 单位：天 默认：7
}
func NewFileLogger() *fileLogger {
	file, err := makeLogFile()
	if err != nil {
		file = os.Stdout
		fmt.Println("file logger create failed, redirect to stdlogger.")
	}
	return &fileLogger{
		dateFormat: defaultTimeTemp,
		level: DEBUG | INFO | WARNING | ERROR,
		expiredTime: 7,
		output: file,
	}
}
func (l *fileLogger) SetDateFormat(temp string) *fileLogger {
	l.dateFormat = temp
	return l
}

func (l *fileLogger) SetFlow(flow string) *fileLogger {
	l.flow = flow
	return l
}

func (l *fileLogger) SetShowLevel(level logLevel) *fileLogger {
	l.level = level
	return l
}

func (l *fileLogger) SetExpiredTimeTime(expiredTime int) *fileLogger {
	l.expiredTime = expiredTime
	return l
}



func (l *fileLogger) Debug(v ...interface{}) {
	if l.level & DEBUG != 0 {
		content := makeContent(l.dateFormat, l.flow, DEBUG, v...)
		l.print(content)
	}
}

func (l *fileLogger) Debugf(format string, v ...interface{}) {
	if l.level & DEBUG != 0 {
		msg := makeMsg(l.dateFormat, l.flow, DEBUG, format, v...)
		l.print(msg)
	}
}

func (l *fileLogger) Info(v ...interface{}) {
	if l.level & INFO != 0 {
		content := makeContent(l.dateFormat, l.flow, INFO, v...)
		l.print(content)
	}
}

func (l *fileLogger) Infof(format string, v ...interface{}) {
	if l.level & INFO != 0 {
		msg := makeMsg(l.dateFormat, l.flow, INFO, format, v...)
		l.print(msg)
	}
}

func (l *fileLogger) Warning(v ...interface{}) {
	if l.level & WARNING != 0 {
		content := makeContent(l.dateFormat, l.flow, WARNING, v...)
		l.print(content)
	}
}

func (l *fileLogger) Warningf(format string, v ...interface{}) {
	if l.level & WARNING != 0 {
		msg := makeMsg(l.dateFormat, l.flow, WARNING, format, v...)
		l.print(msg)
	}
}

func (l *fileLogger) Error(v ...interface{}) {
	if l.level & ERROR != 0 {
		content := makeContent(l.dateFormat, l.flow, ERROR, v...)
		l.print(content)
	}
}

func (l *fileLogger) Errorf(format string, v ...interface{}) {
	if l.level & ERROR != 0 {
		msg := makeMsg(l.dateFormat, l.flow, ERROR, format, v...)
		l.print(msg)
	}
}

func makeLogFile() (*os.File, error) {
	now := time.Now().Format(defaultDateTemp)
	fileName := now + ".log"
	fmt.Println(fileName)
	file, err := os.OpenFile(fileName,os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("log file open failed, err = %v", err)
	}
	return file, nil
}

func (s *fileLogger) print(log string) {
	if s.output == os.Stdout {
		fmt.Fprintln(s.output, log)
	} else {
		s.checkout()
		fmt.Fprintln(s.output, log)
	}
}

func (s *fileLogger) checkout() {
	now := time.Now().Format(defaultDateTemp)
	filename := s.output.Name()
	if filename[:10] != now {
		s.changeOutput()
		s.removeExpiredLogs()
	}
}

func (s *fileLogger) changeOutput() {
	s.output.Close()
	file, err := makeLogFile()
	if err == nil {
		s.output = file
	} else {
		fmt.Println("file logger create failed, redirect to stdlogger.")
		s.output = os.Stdout
	}
}

func (s *fileLogger) removeExpiredLogs() {
	files, err := filepath.Glob("*.log")
	if err != nil {
		fmt.Printf("failed to remove expired log files, error: %s", err)
		return
	}
	delTime := time.Now().AddDate(0, 0, -s.expiredTime).Format(defaultDateTemp)
	for _, file := range files {
		if file[:10] > delTime {
			continue
		}
		if err := os.Remove(file); err != nil {
			fmt.Printf("failed to remove expired log file: %s", file)
		}
	}
}
