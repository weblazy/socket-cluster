package logx

type LogLevel uint8

const (
	DEBUG LogLevel = 1 << iota
	INFO
	WARNING
	ERROR
)

type Logx interface {
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Warning(v ...interface{})
	Warningf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
}

