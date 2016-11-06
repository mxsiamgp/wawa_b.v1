package domain

import "gopkg.in/mgo.v2/bson"

// 门票
type Ticket struct {
	// ID
	ID          bson.ObjectId `bson:"_id,omitempty" json:"id"`

	// 门票名
	Name        string `bson:"name" json:"name"`

	// 描述
	Description string `bson:"description" json:"description"`

	// 价格
	PriceFee    int `bson:"priceFee" json:"price_fee"`
}
