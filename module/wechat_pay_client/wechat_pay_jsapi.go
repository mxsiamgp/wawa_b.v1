package wechat_pay_client

import (
	"fmt"
	"strings"
	"time"

	"github.com/satori/go.uuid"
)

type WechatPayJSAPIConfig struct {
	// timestamp
	Timestamp   int64 `json:"timestamp"`

	// nonceStr
	NonceString string `json:"nonce_string"`

	// paySign
	Signature   string `json:"signature"`
}

// 获取JSAPI配置
func (cli *WechatPayClient) GetJSAPIConfig(prepayID string) *WechatPayJSAPIConfig {
	ts := time.Now().Unix()
	nonceStr := strings.Replace(uuid.NewV4().String(), "-", "", -1)

	params := make(map[string]string)

	// 收集参数
	params["appId"] = cli.appID
	params["timeStamp"] = fmt.Sprintf("%d", ts)
	params["nonceStr"] = nonceStr
	params["package"] = fmt.Sprintf("prepay_id=%s", prepayID)
	params["signType"] = "MD5"

	sign := cli.Sign(params)

	return &WechatPayJSAPIConfig{
		Timestamp: ts,
		NonceString: nonceStr,
		Signature: sign,
	}
}
