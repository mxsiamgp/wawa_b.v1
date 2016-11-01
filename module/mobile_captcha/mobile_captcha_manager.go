package mobile_captcha

import (
	"wawa_b.v1/module/log"
	"wawa_b.v1/module/rest_json_rpc/failure"
	"wawa_b.v1/module/top"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/Sirupsen/logrus"
)

// 失败代码
const (
	// 发送验证码失败
	FAIL_CD_SEND_MOBILE_CAPTCHA_FAIL = "MOBILE_CAPTCHA.SEND_MOBILE_CAPTCHA_FAIL"

	// 验证码错误
	FAIL_CD_INCORRECT_MOBILE_CAPTCHA = "MOBILE_CAPTCHA.INCORRECT_MOBILE_CAPTCHA"
)

// 手机验证码
type MobileCaptcha struct {
	// 手机
	Mobile string `json:"mobile"`

	// 验证码
	Code   string `json:"code"`
}

// 手机验证码管理器
type MobileCaptchaManager interface {
	// 发送手机验证码
	Send(mobile string) (*MobileCaptcha, error)
}

// TOP手机验证码管理器
type TOPMobileCaptchaManager struct {
	// 日志
	logger               *logrus.Logger

	// TOP客户端
	topClient            *top.TOPClient

	// 验证码生成器
	captchaCodeGenerator CaptchaCodeGenerator

	// 短信模板ID
	// For sms_template_code
	// 短信模板中的变量：
	// ${product} 产品名称
	// ${code} 验证码
	templateCode         string

	// 产品名称
	// For sms_param.product
	product              string

	// 短信签名
	// For sms_free_sign_name
	signName             string
}

// 创建一个TOP手机验证码管理器
func NewTOPMobileCaptchaManager(topCli *top.TOPClient, captCdGen CaptchaCodeGenerator, tplCode, prod, signName string) *TOPMobileCaptchaManager {
	return &TOPMobileCaptchaManager{
		logger: log.GetLogger("mobile_captcha.mobileCaptchaManager"),
		topClient: topCli,
		captchaCodeGenerator: captCdGen,
		templateCode: tplCode,
		product: prod,
		signName: signName,
	}
}

func (mgr *TOPMobileCaptchaManager) Send(mobi string) (*MobileCaptcha, error) {
	code := mgr.captchaCodeGenerator.Generate()

	paramJSON, err := ffjson.Marshal(map[string]string{
		"product": mgr.product,
		"code": code,
	})
	if err != nil {
		panic(err)
	}

	res, reqBody, resBody := mgr.topClient.POST("alibaba.aliqin.fc.sms.num.send", map[string]string{
		"sms_type": "normal",
		"sms_free_sign_name": mgr.signName,
		"sms_param": string(paramJSON),
		"rec_num": mobi,
		"sms_template_code": mgr.templateCode,
	})

	if res.ExistsP("error_response") {
		mgr.logger.WithFields(logrus.Fields{
			"requestBody": string(reqBody),
			"responseBody": string(resBody),
		}).Error("发送手机验证码失败")
		return nil, failure.New(FAIL_CD_SEND_MOBILE_CAPTCHA_FAIL)
	}

	return &MobileCaptcha{
		Mobile: mobi,
		Code: code,
	}, nil
}
