package mgorus

import (
	"github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoDBHooker struct {
	c *mgo.Collection
}

func NewHooker(logDB *mgo.Database) *MongoDBHooker {
	return &MongoDBHooker{
		c: logDB.C("Logs"),
	}
}

func (h *MongoDBHooker) Fire(entry *logrus.Entry) error {
	entry.Data["level"] = entry.Level.String()
	entry.Data["time"] = entry.Time
	entry.Data["message"] = entry.Message
	if errData, ok := entry.Data[logrus.ErrorKey]; ok {
		if err, ok := errData.(error); ok && entry.Data[logrus.ErrorKey] != nil {
			entry.Data[logrus.ErrorKey] = err.Error()
		}
	}
	err := h.c.Insert(bson.M(entry.Data))
	if err != nil {
		panic(err)
	}

	return nil
}

func (h *MongoDBHooker) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}
