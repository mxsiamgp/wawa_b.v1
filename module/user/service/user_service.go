package service

import (
	"errors"

	"wawa_b.v1/module/md5"
	"wawa_b.v1/module/mobile_captcha"
	"wawa_b.v1/module/rest_json_rpc"
	"wawa_b.v1/module/rest_json_rpc/failure"
	"wawa_b.v1/module/session"
	"wawa_b.v1/module/user/business"
	"wawa_b.v1/module/user/domain/permission"
	"wawa_b.v1/module/wechat_client"

	"github.com/labstack/echo"
	"gopkg.in/mgo.v2/bson"
)

// 会话键
const (
	// 当前用户
	SESS_KEY_CURRENT_USER_ID = "USER.CURRENT_USER_ID"

	// 当前用户微信AccessToken
	SESS_KEY_CURRENT_USER_WECHAT_ACCESS_TOKEN = "USER.CURRENT_USER_WECHAT_ACCESS_TOKEN"

	// 当前用户微信用户信息
	SESS_KEY_CURRENT_USER_WECHAT_USER_INFO = "USER.CURRENT_USER_WECHAT_USER_INFO"

	// 用户注册验证码
	SESS_KEY_MOBILE_CAPTCHA_FOR_REGISTER = "USER.MOBILE_CAPTCHA_FOR_REGISTER"

	// 找回密码验证码
	SESS_KEY_MOBILE_CAPTCHA_FOR_RETAKE_PASSWORD = "USER.MOBILE_CAPTCHA_FOR_RETAKE_PASSWORD"

	// 找回密码手机
	SESS_KEY_MOBILE_FOR_RETAKE_PASSWORD = "USER.MOBILE_FOR_RETAKE_PASSWORD"

	// 找回密码用户名
	SESS_KEY_USER_ID_FOR_RETAKE_PASSWORD = "USER.NAME_FOR_RETAKE_PASSWORD"
)

// 失败代码
const (
	// 不能删除主办方管理员
	FAIL_CD_CANNOT_DELETE_SPONSOR_MANAGER = "USER.CANNOT_DELETE_SPONSOR_MANAGER"

	// 密码不正确
	FAIL_CD_INCORRECT_PASSWORD = "USER.INCORRECT_PASSWORD"

	// 用户不存在
	FAIL_CD_NO_SUCH_USER = "USER.NO_SUCH_USER"

	// 用户没有权限操作资源
	FAIL_CD_PERMISSION_DENIED = "USER.PERMISSION_DENIED"

	// 资源不隶属该用户
	FAIL_CD_RESOURCE_NOT_BELONG_TO_CURRENT_USER = "USER.RESOURCE_NOT_BELONG_TO_CURRENT_USER"

	// 用户未登录
	FAIL_CD_USER_NOT_LOGGED_IN = "USER.USER_NOT_LOGGED_IN"

	// 微信授权重定向
	FAIL_CD_WECHAT_REDIRECT = "USER.WECHAT_AUTH_REDIRECT"
)

type DeleteParam struct {
	// ID
	ID string `json:"id"`
}

// 删除一个用户
func DeleteProcessHandler(userMgr business.UserManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*DeleteParam)

		user := userMgr.Get(param.ID)
		if user != nil && permission.IsInclude(user.FlatPermissions, []string{"USER.SPONSOR_MANAGER"}) {
			panic(failure.New(FAIL_CD_CANNOT_DELETE_SPONSOR_MANAGER))
		}

		userMgr.Delete(param.ID)
		return nil
	}
}

// 确保登录
func EnsureLoggedInProcessHandler() rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, _ interface{}, ch *rest_json_rpc.ProcessChain) interface{} {
		sess := session.GetSessionByContext(ctx)
		userID := new(bson.ObjectId)
		if !sess.Get(SESS_KEY_CURRENT_USER_ID, userID) {
			panic(failure.New(FAIL_CD_USER_NOT_LOGGED_IN))
		}
		return ch.Next()
	}
}

// 确保用户有权限操作资源
// + 确保登录
func EnsureRequiredPermissionsProcessHandler(userMgr business.UserManager, perms []string) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, ch *rest_json_rpc.ProcessChain) interface{} {
		sess := session.GetSessionByContext(ctx)
		userID := new(bson.ObjectId)
		if !sess.Get(SESS_KEY_CURRENT_USER_ID, userID) {
			panic(errors.New("请确保用户已登录"))
		}

		user := userMgr.Get(userID.Hex())
		if !permission.IsInclude(user.FlatPermissions, perms) {
			panic(failure.New(FAIL_CD_PERMISSION_DENIED))
		}

		return ch.Next()
	}
}

// 资源是否属于该用户
type ResourceBelongToUser func(ctx echo.Context, param interface{}, userID string) bool

// 确保操作属于当前用户的资源
// + 确保登录
func EnsureResourceBelongToCurrentUserProcessHandler(userMgr business.UserManager, rbtu ResourceBelongToUser) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, ch *rest_json_rpc.ProcessChain) interface{} {
		sess := session.GetSessionByContext(ctx)
		userID := new(bson.ObjectId)
		if !sess.Get(SESS_KEY_CURRENT_USER_ID, userID) {
			panic(errors.New("请确保用户已登录"))
		}

		user := userMgr.Get(userID.Hex())
		if !rbtu(ctx, p, user.ID.Hex()) {
			panic(failure.New(FAIL_CD_RESOURCE_NOT_BELONG_TO_CURRENT_USER))
		}

		return ch.Next()
	}
}

type WechatAuthRedirectFailDetail struct {
	// 重定向URL
	Location string `json:"location"`
}

// 确保微信授权
func EnsureWechatAuthorizedProcessHandler(wcCli *wechat_client.WechatClient, redirectURI, scope, state string) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, _ interface{}, ch *rest_json_rpc.ProcessChain) interface{} {
		if len(state) == 0 {
			state = ctx.Request().Referer()
		}

		sess := session.GetSessionByContext(ctx)

		accToken := &wechat_client.WechatAccessToken{}
		if !sess.Get(SESS_KEY_CURRENT_USER_WECHAT_ACCESS_TOKEN, accToken) {
			panic(failure.NewWithDetail(FAIL_CD_WECHAT_REDIRECT, &WechatAuthRedirectFailDetail{
				Location: wcCli.GetAuthorizeURL(redirectURI, scope, state),
			}))
		}

		return ch.Next()
	}
}

type GetParam struct {
	// ID
	ID string `json:"id"`
}

// 获取一个用户
func GetProcessHandler(userMgr business.UserManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*GetParam)
		return userMgr.Get(param.ID)
	}
}

type GetCurrentUserParam struct{}

// 获取当前用户
// + 确保登录
func GetCurrentUserProcessHandler(userMgr business.UserManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		sess := session.GetSessionByContext(ctx)
		userID := new(bson.ObjectId)
		if !sess.Get(SESS_KEY_CURRENT_USER_ID, userID) {
			panic(errors.New("请确保用户已登录"))
		}
		return userMgr.Get(userID.Hex())
	}
}

type GrantFlatPermissionsParam struct {
	// 授权用户ID
	GranterID       string `json:"granter_id"`

	// 被授权用户ID
	GranteeID       string `json:"grantee_id"`

	// 权限集
	FlatPermissions []string `json:"flat_permissions"`
}

// 授予用户权限
func GrantFlatPermissionsProcessHandler(userMgr business.UserManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*GrantFlatPermissionsParam)

		if err := userMgr.GrantFlatPermissions(param.GranterID, param.GranteeID, param.FlatPermissions); err != nil {
			panic(err)
		}

		return nil
	}
}

type LoginParam struct {
	// 名称
	Name     string `json:"name"`

	// 密码
	Password string `json:"password"`
}

// 登录
func LoginProcessHandler(userMgr business.UserManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*LoginParam)
		user := userMgr.GetByName(param.Name)
		if user == nil {
			panic(failure.New(FAIL_CD_NO_SUCH_USER))
		}

		if md5.StringDigest([]byte(param.Password)) != user.PasswordDigest {
			panic(failure.New(FAIL_CD_INCORRECT_PASSWORD))
		}

		sess := session.GetSessionByContext(ctx)
		sess.Set(SESS_KEY_CURRENT_USER_ID, user.ID)

		accToken := &wechat_client.WechatAccessToken{}
		if sess.Get(SESS_KEY_CURRENT_USER_WECHAT_ACCESS_TOKEN, accToken) {
			userMgr.BindWechatOpenID(user.ID.Hex(), accToken.OpenID)
		}

		return user
	}
}

type LogoutParam struct{}

// 注销
func LogoutProcessHandler() rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		sess := session.GetSessionByContext(ctx)
		sess.Remove(SESS_KEY_CURRENT_USER_ID)
		return nil
	}
}

type LogoutForWechatParam struct{}

// 微信注销
// + 确保微信授权
func LogoutForWechatProcessHandler(userMgr business.UserManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		sess := session.GetSessionByContext(ctx)

		accToken := &wechat_client.WechatAccessToken{}
		if !sess.Get(SESS_KEY_CURRENT_USER_WECHAT_ACCESS_TOKEN, accToken) {
			panic(errors.New("请确保微信已授权"))
		}
		userMgr.UnbindWechatOpenID(accToken.OpenID)

		sess.Remove(SESS_KEY_CURRENT_USER_ID)
		return nil
	}
}

type RegisterParam struct {
	// 用户类型
	Kind              string `json:"kind"`

	// 用户名
	Name              string `json:"name"`

	// 密码
	Password          string `json:"password"`

	// 昵称
	Nickname          string `json:"nickname"`

	// 手机
	Mobile            string `json:"mobile"`

	// 手机验证码
	MobileCaptchaCode string `json:"mobile_captcha_code"`
}

// 注册一个新用户
func RegisterProcessHandler(userMgr business.UserManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*RegisterParam)
		if err := mobile_captcha.CheckMobileCaptcha(ctx, SESS_KEY_MOBILE_CAPTCHA_FOR_REGISTER, param.Mobile, param.MobileCaptchaCode); err != nil {
			panic(err)
		}
		if err := userMgr.Register(param.Kind, param.Name, param.Password, param.Nickname, param.Mobile); err != nil {
			panic(err)
		}
		return nil
	}
}

type RetakePasswordParam struct {
	// 手机验证码
	MobileCaptchaCode string `json:"mobile_captcha_code"`

	// 新密码
	Password          string `json:"password"`
}

// 修改一个用户密码
func RetakePasswordProcessHandler(userMgr business.UserManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*RetakePasswordParam)

		sess := session.GetSessionByContext(ctx)
		var userID, mobi string
		if !sess.Get(SESS_KEY_USER_ID_FOR_RETAKE_PASSWORD, &userID) || !sess.Get(SESS_KEY_MOBILE_FOR_RETAKE_PASSWORD, &mobi) {
			panic(failure.New(mobile_captcha.FAIL_CD_INCORRECT_MOBILE_CAPTCHA))
		}

		if err := mobile_captcha.CheckMobileCaptcha(ctx, SESS_KEY_MOBILE_CAPTCHA_FOR_RETAKE_PASSWORD, mobi, param.MobileCaptchaCode); err != nil {
			panic(err)
		}

		userMgr.UpdatePassword(userID, param.Password)
		return nil
	}
}

type RetrieveParam struct {
	// 最后一个ID
	LastID   *string `json:"last_id"`

	// 用户名
	Name     string `json:"name"`

	// 昵称
	Nickname string `json:"nickname"`
}

// 检索用户
func RetrieveProcessHandler(userMgr business.UserManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*RetrieveParam)
		return userMgr.Retrieve(param.LastID, 15, param.Name, param.Nickname)
	}
}

type SendMobileCaptchaForRegisterParam struct {
	// 手机
	Mobile string `json:"mobile"`
}

// 发送用户注册手机验证码
func SendMobileCaptchaForRegisterProcessHandler(regMobiCaptMgr mobile_captcha.MobileCaptchaManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*SendMobileCaptchaForRegisterParam)
		mobiCapt, err := regMobiCaptMgr.Send(param.Mobile)
		if err != nil {
			panic(err)
		}
		session.GetSessionByContext(ctx).Set(SESS_KEY_MOBILE_CAPTCHA_FOR_REGISTER, mobiCapt)
		return nil
	}
}

type SendMobileCaptchaForRetakePasswordParam struct {
	// 用户名
	Name string `json:"name"`
}

// 发送用户注册手机验证码
func SendMobileCaptchaForRetakePasswordProcessHandler(userMgr business.UserManager, rtPwdMobiCaptMgr mobile_captcha.MobileCaptchaManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*SendMobileCaptchaForRetakePasswordParam)

		user := userMgr.GetByName(param.Name)
		if user == nil {
			return nil
		}

		mobiCapt, err := rtPwdMobiCaptMgr.Send(user.Mobile)
		if err != nil {
			panic(err)
		}

		sess := session.GetSessionByContext(ctx)
		sess.Set(SESS_KEY_USER_ID_FOR_RETAKE_PASSWORD, user.ID)
		sess.Set(SESS_KEY_MOBILE_FOR_RETAKE_PASSWORD, user.Mobile)
		sess.Set(SESS_KEY_MOBILE_CAPTCHA_FOR_RETAKE_PASSWORD, mobiCapt)
		return nil
	}
}

type UpdateParam struct {
	// ID
	ID       string `json:"id"`

	// 昵称
	Nickname string `json:"nickname"`
}

// 更新一个用户的基本信息
func UpdateProcessHandler(userMgr business.UserManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*UpdateParam)
		userMgr.Update(param.ID, param.Nickname)
		return nil
	}
}

type UpdatePasswordParam struct {
	// ID
	ID       string `json:"id"`

	// 密码
	Password string `json:"password"`
}

// 修改一个用户密码
func UpdatePasswordProcessHandler(userMgr business.UserManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*UpdatePasswordParam)
		userMgr.UpdatePassword(param.ID, param.Password)
		return nil
	}
}

// 微信网页授权登录
func WechatAuthHandlerFunc(userMgr business.UserManager, wcCli *wechat_client.WechatClient, loginURL string) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		code := ctx.QueryParam("code")
		if len(code) == 0 {
			return nil
		}

		rdURL := ctx.QueryParam("state")

		accToken := wcCli.GetAccessToken(code)

		sess := session.GetSessionByContext(ctx)
		sess.Set(SESS_KEY_CURRENT_USER_WECHAT_ACCESS_TOKEN, accToken)

		if accToken.Scope == "snsapi_userinfo" {
			userInfo := wcCli.GetUserInfo(accToken.AccessToken, accToken.OpenID)
			sess.Set(SESS_KEY_CURRENT_USER_WECHAT_USER_INFO, userInfo)
		} else {
			sess.Remove(SESS_KEY_CURRENT_USER_WECHAT_USER_INFO)
		}

		user := userMgr.GetByWechatOpenID(accToken.OpenID)
		if user == nil {
			ctx.Redirect(302, loginURL)
			return nil
		}
		sess.Set(SESS_KEY_CURRENT_USER_ID, user.ID)

		ctx.Redirect(302, rdURL)
		return nil
	}
}
