package log

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

// 日志集
var loggers map[string]*logrus.Logger = map[string]*logrus.Logger{}

// 日志集互斥量
var loggersMutex *sync.RWMutex = &sync.RWMutex{}

// 获取日志
func GetLogger(name string) *logrus.Logger {
	loggersMutex.RLock()
	defer loggersMutex.RUnlock()
	lgr, ok := loggers[name]

	// 若没有指定名称的日志则创建一个输出到标准错误的日志
	if !ok {
		lgr = NewLogger()
		loggers[name] = lgr
	}
	return lgr
}

// 添加日志
// 日志命名规范：[模块名].[日志名]
func AddLogger(name string, lgr *logrus.Logger) {
	loggersMutex.Lock()
	defer loggersMutex.Unlock()
	loggers[name] = lgr
}
