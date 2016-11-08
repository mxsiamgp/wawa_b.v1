package main

import (
	"flag"
	"fmt"
	"net/url"
	"time"

	competition_business "wawa_b.v1/module/competition/business"
	competition_service "wawa_b.v1/module/competition/service"
	"wawa_b.v1/module/config"
	merchant_business "wawa_b.v1/module/merchant/business"
	merchant_service "wawa_b.v1/module/merchant/service"
	"wawa_b.v1/module/mobile_captcha"
	order_business "wawa_b.v1/module/order/business"
	order_service "wawa_b.v1/module/order/service"
	"wawa_b.v1/module/rest_json_rpc"
	"wawa_b.v1/module/session"
	"wawa_b.v1/module/top"
	user_business "wawa_b.v1/module/user/business"
	user_service "wawa_b.v1/module/user/service"
	wechat_service "wawa_b.v1/module/wechat/service"
	"wawa_b.v1/module/wechat_client"
	wechat_pay_service "wawa_b.v1/module/wechat_pay/service"
	"wawa_b.v1/module/wechat_pay_client"

	"github.com/labstack/echo"
	echo_fasthttp "github.com/labstack/echo/engine/fasthttp"
	"github.com/labstack/echo/middleware"
	"github.com/valyala/fasthttp"
	"gopkg.in/mgo.v2"
	"gopkg.in/redis.v4"
)

func main() {
	argConfDir := flag.String("confDir", "/etc/mxsiamgp", "配置文件目录")
	argProf := flag.String("profile", "production", "配置环境")
	flag.Parse()

	v := config.NewConfig(*argConfDir, *argProf)

	e := echo.New()

	originURL := &url.URL{
		Scheme: "http",
		Host: v.GetString("frontend.host"),
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			originURL.String(),
		},
		AllowMethods: []string{echo.OPTIONS, echo.POST},
		AllowCredentials: true,
		MaxAge: 86400,
	}))

	e.Use(middleware.Recover())

	sessMgr := session.NewManager("mxsiamgp.sid", session.NewRedisSessionStore(redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", v.GetString("session.storage.redis.host"), v.GetInt("session.storage.redis.port")),
	}), time.Duration(v.GetInt("session.storage.expiration")) * time.Millisecond))
	e.Use(sessMgr.HandlerFunc())

	rpc := rest_json_rpc.NewRPC()
	e.POST("/rest_json_rpc", rpc.HandlerFunc())

	mgoConn, err := mgo.Dial(fmt.Sprintf("mongodb://%s:%d", v.GetString("database.mongodb.host"), v.GetInt("database.mongodb.port")))
	if err != nil {
		panic(err)
	}

	topCli := top.NewTOPClient(&fasthttp.Client{}, top.HTTP_OFFICIAL, v.GetString("top.appKey"), v.GetString("top.appSecret"))
	captCdGen := mobile_captcha.NewRandDigitalCaptchaGenerator(6)

	wcCli := wechat_client.NewWechatClient(&fasthttp.Client{}, v.GetString("wechat.appId"), v.GetString("wechat.appSecret"))

	userMgr := user_business.NewMongoDBUserManager(mgoConn.DB("mxsiamgp"), map[string][]string{
		"ANONYMOUS_USER": []string{},
	})

	// 用户模块过程
	rpc.RegisterProcess("user.get", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.GetProcessHandler(userMgr),
		},
		ParamFactory: func() interface{} {
			return &user_service.GetParam{}
		},
	})

	rpc.RegisterProcess("user.send_mobile_captcha_for_register", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.SendMobileCaptchaForRegisterProcessHandler(
				mobile_captcha.NewTOPMobileCaptchaManager(topCli, captCdGen,
					v.GetString("top.mobileCaptchaManager.user.register.templateCode"),
					v.GetString("top.mobileCaptchaManager.user.register.product"),
					v.GetString("top.mobileCaptchaManager.user.register.sign"))),
		},
		ParamFactory: func() interface{} {
			return &user_service.SendMobileCaptchaForRegisterParam{}
		},
	})
	rpc.RegisterProcess("user.register", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.RegisterProcessHandler(userMgr),
		},
		ParamFactory: func() interface{} {
			return &user_service.RegisterParam{}
		},
	})
	rpc.RegisterProcess("user.retrieve", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureRequiredPermissionsProcessHandler(userMgr, []string{
				"USER.RETRIEVE",
			}),
			user_service.RetrieveProcessHandler(userMgr),
		},
		ParamFactory: func() interface{} {
			return &user_service.RetrieveParam{}
		},
	})
	rpc.RegisterProcess("user.send_mobile_captcha_for_retake_password", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.SendMobileCaptchaForRetakePasswordProcessHandler(
				userMgr, mobile_captcha.NewTOPMobileCaptchaManager(topCli, captCdGen,
					v.GetString("top.mobileCaptchaManager.user.retakePassword.templateCode"),
					v.GetString("top.mobileCaptchaManager.user.retakePassword.product"),
					v.GetString("top.mobileCaptchaManager.user.retakePassword.sign"))),
		},
		ParamFactory: func() interface{} {
			return &user_service.SendMobileCaptchaForRetakePasswordParam{}
		},
	})
	rpc.RegisterProcess("user.retake_password", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.RetakePasswordProcessHandler(userMgr),
		},
		ParamFactory: func() interface{} {
			return &user_service.RetakePasswordParam{}
		},
	})
	rpc.RegisterProcess("user.update", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureResourceBelongToCurrentUserProcessHandler(userMgr, func(_ echo.Context, p interface{}, userID string) bool {
				param := p.(*user_service.UpdateParam)
				return param.ID == userID
			}),
			user_service.UpdateProcessHandler(userMgr),
		},
		ParamFactory: func() interface{} {
			return &user_service.UpdateParam{}
		},
	})
	rpc.RegisterProcess("user.update_password", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureResourceBelongToCurrentUserProcessHandler(userMgr, func(_ echo.Context, p interface{}, userID string) bool {
				param := p.(*user_service.UpdatePasswordParam)
				return param.ID == userID
			}),
			user_service.UpdatePasswordProcessHandler(userMgr),
		},
		ParamFactory: func() interface{} {
			return &user_service.UpdatePasswordParam{}
		},
	})
	rpc.RegisterProcess("user.delete", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureRequiredPermissionsProcessHandler(userMgr, []string{
				"USER.MODIFY",
			}),
			user_service.DeleteProcessHandler(userMgr),
		},
		ParamFactory: func() interface{} {
			return &user_service.DeleteParam{}
		},
	})
	rpc.RegisterProcess("user.grant_flat_permissions", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureResourceBelongToCurrentUserProcessHandler(userMgr, func(_ echo.Context, p interface{}, userID string) bool {
				param := p.(*user_service.GrantFlatPermissionsParam)
				return param.GranterID == userID
			}),
			user_service.GrantFlatPermissionsProcessHandler(userMgr),
		},
		ParamFactory: func() interface{} {
			return &user_service.GrantFlatPermissionsParam{}
		},
	})
	rpc.RegisterProcess("user.get_current_user", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.GetCurrentUserProcessHandler(userMgr),
		},
		ParamFactory: func() interface{} {
			return &user_service.GetCurrentUserParam{}
		},
	})

	wcAuthURL := &url.URL{
		Scheme: "http",
		Host: v.GetString("backend.host"),
		Path: "/user/wechat_auth",
	}

	rpc.RegisterProcess("user.get_current_wechat_user", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureWechatAuthorizedProcessHandler(wcCli, wcAuthURL.String(), "snsapi_userinfo", ""),
			user_service.EnsureLoggedInProcessHandler(),
			user_service.GetCurrentUserProcessHandler(userMgr),
		},
		ParamFactory: func() interface{} {
			return &user_service.GetCurrentUserParam{}
		},
	})
	rpc.RegisterProcess("user.login", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.LoginProcessHandler(userMgr),
		},
		ParamFactory: func() interface{} {
			return &user_service.LoginParam{}
		},
	})
	wcLoginURL := &url.URL{
		Scheme: "http",
		Host: v.GetString("frontend.host"),
		Path: "/module/wechat/html/index.html",
		Fragment: "/user/wechat/login",
	}
	e.GET("/user/wechat_auth", user_service.WechatAuthHandlerFunc(userMgr, wcCli, wcLoginURL.String()))
	rpc.RegisterProcess("user.logout", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.LogoutProcessHandler(),
		},
		ParamFactory: func() interface{} {
			return &user_service.LogoutParam{}
		},
	})
	rpc.RegisterProcess("user.logout_for_wechat", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureWechatAuthorizedProcessHandler(wcCli, wcAuthURL.String(), "snsapi_userinfo", ""),
			user_service.LogoutForWechatProcessHandler(userMgr),
		},
		ParamFactory: func() interface{} {
			return &user_service.LogoutForWechatParam{}
		},
	})

	// 赛事模块过程
	cmptMgr := competition_business.NewMongoDBCompetitionManager(mgoConn.DB("mxsiamgp"))
	rpc.RegisterProcess("competition.add", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureRequiredPermissionsProcessHandler(userMgr, []string{
				"COMPETITION.MODIFY",
			}),
			competition_service.AddProcessHandler(cmptMgr),
		},
		ParamFactory: func() interface{} {
			return &competition_service.AddParam{}
		},
	})

	wcPayCli := wechat_pay_client.NewWechatPayClient(&fasthttp.Client{}, v.GetString("wechat.appId"), v.GetString("wechat.mchId"), v.GetString("wechat.partnerKey"))

	wcH5PayWpTitle := "海南超跑赛车收费服务平台"

	pncbist := map[string]order_business.OrderItemPayNotifyCallback{}
	orderMgr := order_business.NewMongoDBOrderManager(mgoConn.DB("mxsiamgp"), pncbist, wcPayCli, &wcH5PayWpTitle)

	ticketMgr := competition_business.NewMongoDBDrawnTicketManager(mgoConn.DB("mxsiamgp"), cmptMgr, orderMgr)

	pncbist["COMPETITION.TICKET"] = ticketMgr.OrderItemPayNotifyCallback()

	rpc.RegisterProcess("competition.create_order", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			competition_service.CreateOrderProcessHandler(ticketMgr),
		},
		ParamFactory: func() interface{} {
			return &competition_service.CreateOrderParam{}
		},
	})
	rpc.RegisterProcess("competition.delete", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureRequiredPermissionsProcessHandler(userMgr, []string{
				"COMPETITION.MODIFY",
			}),
			competition_service.DeleteProcessHandler(cmptMgr),
		},
		ParamFactory: func() interface{} {
			return &competition_service.DeleteParam{}
		},
	})
	rpc.RegisterProcess("competition.finish", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureRequiredPermissionsProcessHandler(userMgr, []string{
				"COMPETITION.MODIFY",
			}),
			competition_service.FinishProcessHandler(cmptMgr),
		},
		ParamFactory: func() interface{} {
			return &competition_service.FinishParam{}
		},
	})
	rpc.RegisterProcess("competition.get", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			competition_service.GetProcessHandler(cmptMgr),
		},
		ParamFactory: func() interface{} {
			return &competition_service.GetParam{}
		},
	})
	rpc.RegisterProcess("competition.list_in_progress", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			competition_service.ListInProgressProcessHandler(cmptMgr),
		},
		ParamFactory: func() interface{} {
			return &competition_service.ListInProgressParam{}
		},
	})
	rpc.RegisterProcess("competition.retrieve", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureRequiredPermissionsProcessHandler(userMgr, []string{
				"COMPETITION.RETRIEVE",
			}),
			competition_service.RetrieveProcessHandler(cmptMgr),
		},
		ParamFactory: func() interface{} {
			return &competition_service.RetrieveParam{}
		},
	})
	rpc.RegisterProcess("competition.update_tickets", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureRequiredPermissionsProcessHandler(userMgr, []string{
				"COMPETITION.MODIFY",
			}),
			competition_service.UpdateTicketsProcessHandler(cmptMgr),
		},
		ParamFactory: func() interface{} {
			return &competition_service.UpdateTicketsParam{}
		},
	})
	rpc.RegisterProcess("competition.drawn_ticket.get_all_by_competition_id_and_user_id", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureRequiredPermissionsProcessHandler(userMgr, []string{
				"COMPETITION.DRAWN_TICKET.INSPECT",
			}),
			competition_service.GetAllDrawnTicketsByCompetitionIDAndUserIDProcessHandler(ticketMgr),
		},
		ParamFactory: func() interface{} {
			return &competition_service.GetAllDrawnTicketsByCompetitionIDAndUserIDParam{}
		},
	})

	// 商家模块过程
	mcMgr := merchant_business.NewMongoDBMerchantManager(mgoConn.DB("mxsiamgp"), userMgr)
	rpc.RegisterProcess("merchant.delete", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureRequiredPermissionsProcessHandler(userMgr, []string{
				"MERCHANT.MODIFY",
			}),
			merchant_service.DeleteProcessHandler(mcMgr),
		},
		ParamFactory: func() interface{} {
			return &merchant_service.DeleteParam{}
		},
	})
	rpc.RegisterProcess("merchant.get", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			merchant_service.GetProcessHandler(mcMgr),
		},
		ParamFactory: func() interface{} {
			return &merchant_service.GetParam{}
		},
	})
	rpc.RegisterProcess("merchant.get_current_merchant", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			merchant_service.GetCurrentMerchantProcessHandler(mcMgr),
		},
		ParamFactory: func() interface{} {
			return &merchant_service.GetCurrentMerchantParam{}
		},
	})
	rpc.RegisterProcess("merchant.kick_out_staff", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureRequiredPermissionsProcessHandler(userMgr, []string{
				"MERCHANT.STAFF.MODIFY",
			}),
			merchant_service.KickOutStaffProcessHandler(mcMgr),
		},
		ParamFactory: func() interface{} {
			return &merchant_service.KickOutStaffParam{}
		},
	})
	rpc.RegisterProcess("merchant.pull_in_staff", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureRequiredPermissionsProcessHandler(userMgr, []string{
				"MERCHANT.STAFF.MODIFY",
			}),
			merchant_service.EnsureResourceBelongToCurrentMerchantProcessHandler(mcMgr, func(_ echo.Context, p interface{}, merchantID string) bool {
				param := p.(*merchant_service.PullInStaffParam)
				return param.MerchantID == merchantID
			}),
			merchant_service.PullInStaffProcessHandler(mcMgr, userMgr),
		},
		ParamFactory: func() interface{} {
			return &merchant_service.PullInStaffParam{}
		},
	})
	rpc.RegisterProcess("merchant.register", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureRequiredPermissionsProcessHandler(userMgr, []string{
				"MERCHANT.MODIFY",
			}),
			merchant_service.RegisterProcessHandler(mcMgr, userMgr),
		},
		ParamFactory: func() interface{} {
			return &merchant_service.RegisterParam{}
		},
	})
	rpc.RegisterProcess("merchant.retrieve", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureRequiredPermissionsProcessHandler(userMgr, []string{
				"MERCHANT.RETRIEVE",
			}),
			merchant_service.RetrieveProcessHandler(mcMgr),
		},
		ParamFactory: func() interface{} {
			return &merchant_service.RetrieveParam{}
		},
	})
	rpc.RegisterProcess("merchant.retrieve_staffs", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureRequiredPermissionsProcessHandler(userMgr, []string{
				"MERCHANT.STAFF.RETRIEVE",
			}),
			merchant_service.EnsureResourceBelongToCurrentMerchantProcessHandler(mcMgr, func(_ echo.Context, p interface{}, merchantID string) bool {
				param := p.(*merchant_service.RetrieveStaffsParam)
				return param.MerchantID == merchantID
			}),
			merchant_service.RetrieveStaffsProcessHandler(mcMgr),
		},
		ParamFactory: func() interface{} {
			return &merchant_service.RetrieveStaffsParam{}
		},
	})
	rpc.RegisterProcess("merchant.update", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureRequiredPermissionsProcessHandler(userMgr, []string{
				"MERCHANT.MODIFY",
			}),
			merchant_service.UpdateProcessHandler(mcMgr),
		},
		ParamFactory: func() interface{} {
			return &merchant_service.UpdateParam{}
		},
	})

	// 微信模块过程
	rpc.RegisterProcess("wechat.get_wechat_jssdk_config", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			wechat_service.GetWechatJSSDKConfigProcessHandler(wcCli),
		},
		ParamFactory: func() interface{} {
			return &wechat_service.GetWechatJSSDKConfigParam{}
		},
	})

	// 微信支付模块过程
	rpc.RegisterProcess("wechat_pay.get_wechat_pay_jssdk_config", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			wechat_pay_service.GetWechatPayJSSDKConfigProcessHandler(wcPayCli),
		},
		ParamFactory: func() interface{} {
			return &wechat_pay_service.GetWechatPayJSSDKConfigParam{}
		},
	})

	// 订单模块过程
	rpc.RegisterProcess("order.get_all_orders_by_user_id", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureWechatAuthorizedProcessHandler(wcCli, wcAuthURL.String(), "snsapi_userinfo", ""),
			user_service.EnsureLoggedInProcessHandler(),
			user_service.EnsureResourceBelongToCurrentUserProcessHandler(userMgr, func(_ echo.Context, p interface{}, userID string) bool {
				param := p.(*order_service.GetAllOrdersByUserIDParam)
				return param.UserID == userID
			}),
			order_service.GetAllOrdersByUserIDProcessHandler(orderMgr),
		},
		ParamFactory: func() interface{} {
			return &order_service.GetAllOrdersByUserIDParam{}
		},
	})

	wcPayNotifyCbURL := &url.URL{
		Scheme: "http",
		Host: v.GetString("backend.host"),
		Path: "/order/wechat_pay_notify_callback",
	}
	rpc.RegisterProcess("order.pay_by_wechat_h5", &rest_json_rpc.Process{
		Handlers: []rest_json_rpc.ProcessHandler{
			user_service.EnsureWechatAuthorizedProcessHandler(wcCli, wcAuthURL.String(), "snsapi_userinfo", ""),
			user_service.EnsureLoggedInProcessHandler(),
			order_service.PayByWechatH5ProcessHandler(orderMgr, wcPayNotifyCbURL.String()),
		},
		ParamFactory: func() interface{} {
			return &order_service.PayByWechatH5Param{}
		},
	})
	e.POST("/order/wechat_pay_notify_callback", order_service.WechatPayNotifyCallbackHandlerFunc(orderMgr, wcPayCli))

	e.Run(echo_fasthttp.New(fmt.Sprintf("127.0.0.1:%d", v.GetInt("listen.port"))))
}
