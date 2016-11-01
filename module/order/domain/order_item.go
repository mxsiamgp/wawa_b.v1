package domain

import "gopkg.in/mgo.v2/bson"

// 订单项
type OrderItem struct {
	// ID
	ID            bson.ObjectId `bson:"_id,omitempty" json:"id"`

	// 商品类型
	SellableType  string `bson:"sellableType" json:"sellable_type"`

	// 商品
	SellableValue string `bson:"sellableValue" json:"sellable_value"`

	// 数量
	Quantity      int `bson:"quantity" json:"quantity"`

	// 总价（分）
	TotalPriceFee int `bson:"totalPriceFee" json:"total_price_fee"`
}
