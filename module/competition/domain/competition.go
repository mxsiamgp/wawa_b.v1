package domain

import "gopkg.in/mgo.v2/bson"

// 赛事
type Competition struct {
	// ID
	ID         bson.ObjectId `bson:"_id,omitempty" json:"id"`

	// 赛事名
	Name       string `bson:"name" json:"name"`

	// 是否完成
	IsFinished bool `bson:"isFinished" json:"is_finished"`

	// 门票集
	Tickets    []*Ticket `bson:"tickets" json:"tickets"`
}
