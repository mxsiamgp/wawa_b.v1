package service

import (
	"wawa_b.v1/module/rest_json_rpc"
	"wawa_b.v1/module/wechat_client"

	"github.com/labstack/echo"
)

type GetWechatJSSDKConfigParam struct {
	// URL
	URL *string `json:"url"`
}

// 获取微信JSSDK配置
func GetWechatJSSDKConfigProcessHandler(wcCli *wechat_client.WechatClient) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*GetWechatJSSDKConfigParam)

		url := ctx.Request().Referer()
		if param.URL != nil {
			url = *param.URL
		}

		return wcCli.GetJSAPIConfig(url)
	}
}
