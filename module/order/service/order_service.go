package service

import (
	"errors"

	"wawa_b.v1/module/order/business"
	"wawa_b.v1/module/rest_json_rpc"
	"wawa_b.v1/module/session"
	user_service "wawa_b.v1/module/user/service"
	"wawa_b.v1/module/wechat_client"
	"wawa_b.v1/module/wechat_pay_client"

	"github.com/Jeffail/gabs"
	"github.com/labstack/echo"
)

type GetAllOrdersByUserIDParam struct {
	// 最后一个ID
	LastID *string `json:"last_id"`

	// 用户ID
	UserID string `json:"user_id"`
}

// 根据用户ID获取所有订单
func GetAllOrdersByUserIDProcessHandler(orderMgr business.OrderManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*GetAllOrdersByUserIDParam)
		return orderMgr.GetAllOrdersByUserID(param.LastID, 15, param.UserID)
	}
}

type PayByWechatH5Param struct {
	// 订单ID
	OrderID string `json:"order_id"`
}

// 微信H5支付
// + 确保微信授权
func PayByWechatH5ProcessHandler(orderMgr business.OrderManager, notifyURL string) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*PayByWechatH5Param)

		sess := session.GetSessionByContext(ctx)
		accToken := &wechat_client.WechatAccessToken{}
		if !sess.Get(user_service.SESS_KEY_CURRENT_USER_WECHAT_ACCESS_TOKEN, accToken) {
			panic(errors.New("请确保微信授权"))
		}

		prepayID, err := orderMgr.PayByWechatH5(param.OrderID, ctx.Request().RealIP(), notifyURL, accToken.OpenID)
		if err != nil {
			panic(err)
		}
		return prepayID
	}
}

// 微信支付通知回调
func WechatPayNotifyCallbackHandlerFunc(orderMgr business.OrderManager, wcPayCli *wechat_pay_client.WechatPayClient) echo.HandlerFunc {
	return wcPayCli.PayNotifyCallbackHandlerFunc(func(param *gabs.Container) {
		orderMgr.WechatPaid(param.Path("attach").Data().(string))
	})
}
