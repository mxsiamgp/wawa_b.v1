package domain

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	ORD_STAT_UNPAID = "UNPAID"
	ORD_STAT_PAID = "PAID"
)

const (
	ORD_PAY_APPROACH_WECHAT = "WECHAT"
)

// 订单
type Order struct {
	// ID
	ID                  bson.ObjectId `bson:"_id,omitempty" json:"id"`

	// 用户ID
	UserID              bson.ObjectId `bson:"userId" json:"user_id"`

	// 创建时间
	CreatedTime         time.Time `bson:"createdTime" json:"created_time"`

	// 订单项目集合
	Items               []*OrderItem `bson:"items" json:"items"`

	// 总价（分）
	TotalPriceFee       int `bson:"totalPriceFee" json:"total_price_fee"`

	// 状态
	Status              string `bson:"status" json:"status"`

	// 支付方式
	PayApproach         *string `bson:"payApproach" json:"pay_approach"`

	// 微信支付相关 --BEGIN--

	// 商户订单号
	WechatPayOutTradeNo *string `bson:"wechatPayOutTradeNo" json:"wechat_pay_out_trade_no"`

	// 微信支付相关 --END--
}
