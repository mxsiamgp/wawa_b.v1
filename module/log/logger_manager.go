package log

import (
	"wawa_b.v1/module/log/logrus/mgorus"

	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

// 创建一个输出到标准错误的日志
func NewLogger() *logrus.Logger {
	return logrus.New()
}

// 创建一个基于MongoDB的日志
func NewMongoDBLogger(logDB *mgo.Database) *logrus.Logger {
	lgr := logrus.New()
	lgr.Hooks.Add(mgorus.NewHooker(logDB))
	return lgr
}
