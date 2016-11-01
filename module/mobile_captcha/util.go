package mobile_captcha

import (
	"wawa_b.v1/module/rest_json_rpc/failure"
	"wawa_b.v1/module/session"

	"github.com/labstack/echo"
)

// 核实手机验证码
func CheckMobileCaptcha(ctx echo.Context, sessKey, mobi, mobiCaptCode string) error {
	sess := session.GetSessionByContext(ctx)
	mobiCapt := &MobileCaptcha{}
	if !sess.Get(sessKey, mobiCapt) || mobiCapt.Mobile != mobi || mobiCapt.Code != mobiCaptCode {
		return failure.New(FAIL_CD_INCORRECT_MOBILE_CAPTCHA)
	}

	// 核实通过删除验证码
	sess.Remove(sessKey)

	return nil
}
