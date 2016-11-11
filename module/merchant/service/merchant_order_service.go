package service

import (
	"wawa_b.v1/module/merchant/business"
	"wawa_b.v1/module/rest_json_rpc"

	"github.com/labstack/echo"
)

type CreateMerchantOrderParam struct {
	CompetitionID string `json:"competition_id"`
	MerchantID    string `json:"merchant_id"`
	PriceFee      int `json:"price_fee"`
	UserID        string `json:"user_id"`
}

// 创建商家订单
func CreateMerchantOrderProcessHandler(mcOrderMgr business.MerchantOrderManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*CreateMerchantOrderParam)
		mcOrderMgr.CreateOrder(param.UserID, param.CompetitionID, param.MerchantID, param.PriceFee)
		return nil
	}
}
