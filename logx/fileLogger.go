package logx

import (
	"os"
	"time"
	"fmt"
)

type fileLogger struct {
	dateFormat 		string   		// 时间格式 默认：2006-01-05 15:04:05.000
	flow       		string   		// 操作流 用于标记日志信息 不设置无输出
	level      		logLevel 		// 输出level 不设置默认全部输出
	expiredTime 	int				// 清理时间 <0 不清理 单位：天 默认：7
}
func NewFileLogger() *fileLogger {
	return &fileLogger{
		dateFormat: defaultTimeTemp,
		level: DEBUG | INFO | WARNING | ERROR,
		expiredTime: 7,
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
