package logx

import (
	"fmt"
	"os"
)

type stdLogger struct {
	dateFormat 	string   // 时间格式 默认：2006-01-05 15:04:05.000
	flow       	string   // 操作流 用于标记日志信息 不设置无输出
	level      	logLevel // 输出level 不设置默认全部输出
}

func NewStdLogger() *stdLogger {
	return &stdLogger{
		dateFormat: defaultTimeTemp,
		level: DEBUG | INFO | WARNING | ERROR,
	}
}

func (l *stdLogger) SetDateFormat(temp string) *stdLogger {
	l.dateFormat = temp
	return l
}

func (l *stdLogger) SetFlow(flow string) *stdLogger {
	l.flow = flow
	return l
}

func (l *stdLogger) SetShowLevel(level logLevel) *stdLogger {
	l.level = level
	return l
}


func (l *stdLogger) Debug(v ...interface{}) {
	if l.level & DEBUG != 0 {
		content := makeContent(l.dateFormat, l.flow, DEBUG, v...)
		fmt.Fprintln(os.Stdout, content)
	}
}

func (l *stdLogger) Debugf(format string, v ...interface{}) {
	if l.level & DEBUG != 0 {
		msg := makeMsg(l.dateFormat, l.flow, DEBUG, format, v...)
		fmt.Fprintln(os.Stdout, msg)
	}
}

func (l *stdLogger) Info(v ...interface{}) {
	if l.level & INFO != 0 {
		content := makeContent(l.dateFormat, l.flow, INFO, v...)
		fmt.Fprintln(os.Stdout, content)
	}
}

func (l *stdLogger) Infof(format string, v ...interface{}) {
	if l.level & INFO != 0 {
		msg := makeMsg(l.dateFormat, l.flow, INFO, format, v...)
		fmt.Fprintln(os.Stdout, msg)

	}
}

func (l *stdLogger) Warning(v ...interface{}) {
	if l.level & WARNING != 0 {
		content := makeContent(l.dateFormat, l.flow, WARNING, v...)
		fmt.Fprintln(os.Stdout, content)
	}
}

func (l *stdLogger) Warningf(format string, v ...interface{}) {
	if l.level & WARNING != 0 {
		msg := makeMsg(l.dateFormat, l.flow, WARNING, format, v...)
		fmt.Fprintln(os.Stdout, msg)
	}
}

func (l *stdLogger) Error(v ...interface{}) {
	if l.level & ERROR != 0 {
		content := makeContent(l.dateFormat, l.flow, ERROR, v...)
		fmt.Fprintln(os.Stdout, content)
	}
}

func (l *stdLogger) Errorf(format string, v ...interface{}) {
	if l.level & ERROR != 0 {
		msg := makeMsg(l.dateFormat, l.flow, ERROR, format, v...)
		fmt.Fprintln(os.Stdout, msg)
	}
}