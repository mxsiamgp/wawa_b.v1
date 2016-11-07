package domain

import "gopkg.in/mgo.v2/bson"

// 售出的门票
type DrawnTicket struct {
	// 订单ID
	OrderID         bson.ObjectId `bson:"orderId" json:"order_id"`

	// 订单项ID
	OrderItemID     bson.ObjectId `bson:"orderItemId" json:"order_item_id"`

	// 用户ID
	UserID          bson.ObjectId `bson:"userId" json:"user_id"`

	// 赛事ID
	CompetitionID   bson.ObjectId `bson:"competitionId" json:"competition_id"`

	// 赛事名
	CompetitionName string `bson:"competitionName" json:"competition_name"`

	// 门票ID
	TicketID        bson.ObjectId `bson:"ticketId" json:"ticket_id"`

	// 门票名
	TicketName      string `bson:"ticketName" json:"ticket_name"`

	// 价格
	PriceFee        int `bson:"priceFee" json:"price_fee"`

	// 数量
	Quantity        int `bson:"quantity" json:"quantity"`
}
