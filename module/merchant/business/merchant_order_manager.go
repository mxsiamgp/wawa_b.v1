package business

import (
	order_business "wawa_b.v1/module/order/business"
	order_domain "wawa_b.v1/module/order/domain"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// 商家订单管理器
type MerchantOrderManager interface {
	// 创建商家订单
	CreateOrder(userID, competitionID, merchantID string, priceFee int)

	// 订单项目支付通知回调
	OrderItemPayNotifyCallback() order_business.OrderItemPayNotifyCallback
}

// MongoDB商家订单管理器
type MongoDBMerchantOrderManager struct {
	// 商家管理器
	merchantManager MerchantManager

	// 订单集合
	orderCollection *mgo.Collection

	// 订单管理器
	orderManager    order_business.OrderManager
}

// 创建一个MongoDB商家订单管理器
func NewMongoDBMerchantOrderManager(tradeDB *mgo.Database, mcMgr MerchantManager, orderMgr order_business.OrderManager) *MongoDBMerchantOrderManager {
	return &MongoDBMerchantOrderManager{
		merchantManager: mcMgr,
		orderCollection: tradeDB.C("Orders"),
		orderManager: orderMgr,
	}
}

func (mgr *MongoDBMerchantOrderManager) CreateOrder(userID, competitionID, merchantID string, priceFee int) {
	mc := mgr.merchantManager.Get(merchantID)

	orderID := mgr.orderManager.Create(userID, []*order_domain.OrderItem{
		{
			SellableType: "PAYMENT_TO_MERCHANT",
			SellableValue: mc.Name,
			Quantity: 1,
			TotalPriceFee: priceFee,
		},
	})

	if err := mgr.orderCollection.UpdateId(bson.ObjectIdHex(orderID), bson.M{
		"$set": bson.M{
			"competitionId": bson.ObjectIdHex(competitionID),
			"merchantId": mc.ID,
		},
	}); err != nil {
		panic(err)
	}
}

func (mgr *MongoDBMerchantOrderManager) OrderItemPayNotifyCallback() order_business.OrderItemPayNotifyCallback {
	return func(orderID, orderItemID string) {
	}
}
