package service

import (
	"wawa_b.v1/module/rest_json_rpc"
	"wawa_b.v1/module/wechat_pay_client"

	"github.com/labstack/echo"
)

type GetWechatPayJSSDKConfigParam struct {
	// 预支付ID
	PrepayID string `json:"prepay_id"`
}

// 获取微信JSSDK配置
func GetWechatPayJSSDKConfigProcessHandler(wcPayCli *wechat_pay_client.WechatPayClient) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*GetWechatPayJSSDKConfigParam)

		return wcPayCli.GetJSAPIConfig(param.PrepayID)
	}
}
