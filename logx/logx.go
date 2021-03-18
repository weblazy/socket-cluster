package logx

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

type Record struct {
	TimeStamp 		string			`json:"timestamp"`			// 日志生成时间
	Level			string			`json:"level"`				// 日志等级
	Flow			string			`json:"flow,omitempty"`		// 操作流
	Loc 			string			`json:"loc"`				// 发生位置
	Content			[]interface{}	`json:"content,omitempty"`	// 日志内容
	Msg 			string			`json:"msg,omitempty"`		// 日志内容
}