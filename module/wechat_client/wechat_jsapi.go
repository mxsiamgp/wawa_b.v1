package wechat_client

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"wawa_b.v1/module/sha1"

	"github.com/Jeffail/gabs"
	"github.com/satori/go.uuid"
)

// 刷新微信JSAPI票据
func (cli *WechatClient) RefreshJSAPITicket() {
	cli.RefreshAccessToken()

	// 未超时
	if len(cli.jsapiTicket) != 0 && time.Now().Before(cli.jsapiTicketExpirationTime) {
		return
	}

	query := url.Values{}
	query.Set("access_token", cli.accessToken)
	query.Set("type", "jsapi")

	tokenURL := &url.URL{
		Scheme: "https",
		Host: "api.weixin.qq.com",
		Path: "/cgi-bin/ticket/getticket",
		RawQuery: query.Encode(),
	}

	statusCd, body, err := cli.httpClient.Get(nil, tokenURL.String())
	if err != nil {
		panic(err)
	}
	if statusCd != 200 {
		panic(errors.New("获取微信JSAPI票据HTTP调用错误"))
	}

	res, err := gabs.ParseJSON(body)
	if err != nil {
		panic(err)
	}
	if int(res.Path("errcode").Data().(float64)) != 0 {
		panic(errors.New(res.String()))
	}

	cli.jsapiTicket = res.Path("ticket").Data().(string)
	cli.jsapiTicketExpirationTime = time.Now().Add((7200 - 200) * time.Second)
}

type WechatJSAPIConfig struct {
	// appId
	AppID       string `json:"app_id"`

	// timestamp
	Timestamp   int64 `json:"timestamp"`

	// nonceStr
	NonceString string `json:"nonce_string"`

	//signature
	Signature   string `json:"signature"`
}

// 获取JSAPI配置
func (cli *WechatClient) GetJSAPIConfig(urlParam string) *WechatJSAPIConfig {
	cli.RefreshJSAPITicket()

	nonceStr := strings.Replace(uuid.NewV4().String(), "-", "", -1)
	ts := time.Now().Unix()

	params := make(map[string]string)

	// 收集参数
	params["noncestr"] = nonceStr
	params["jsapi_ticket"] = cli.jsapiTicket
	params["timestamp"] = fmt.Sprintf("%d", ts)
	params["url"] = urlParam

	// 生成签名
	// - 生成有序的参数键序列
	orderedParamKeys := make([]string, 0, len(params))
	for key := range params {
		orderedParamKeys = append(orderedParamKeys, key)
	}
	sort.Strings(orderedParamKeys)

	// - 构建签名明文
	signPlaintextBuf := bytes.NewBufferString("")
	for i, key := range orderedParamKeys {
		if i != 0 {
			signPlaintextBuf.WriteString("&")
		}
		signPlaintextBuf.WriteString(key)
		signPlaintextBuf.WriteString("=")
		signPlaintextBuf.WriteString(params[key])
	}

	// - 生成签名
	sign := strings.ToLower(sha1.StringDigest(signPlaintextBuf.Bytes()))

	return &WechatJSAPIConfig{
		AppID: cli.appID,
		Timestamp: ts,
		NonceString: nonceStr,
		Signature: sign,
	}
}
