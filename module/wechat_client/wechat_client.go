package wechat_client

import (
	"errors"
	"net/url"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/valyala/fasthttp"
)

// 微信客户端
type WechatClient struct {
	// HTTP客户端
	httpClient                *fasthttp.Client

	// AppID
	appID                     string

	// AppSecret
	appSecret                 string

	// 微信接口调用凭据
	accessToken               string

	// 微信接口调用凭据过期时间
	accessTokenExpirationTime time.Time

	// 微信JSAPI票据
	jsapiTicket               string

	// 微信JSAPI票据过期时间
	jsapiTicketExpirationTime time.Time
}

// 创建一个微信客户端
func NewWechatClient(httpCli *fasthttp.Client, appID string, appSecret string) *WechatClient {
	return &WechatClient{
		httpClient: httpCli,
		appID: appID,
		appSecret: appSecret,
	}
}

// 刷新AccessToken
func (cli *WechatClient) RefreshAccessToken() {
	// 未超时
	if len(cli.accessToken) != 0 && time.Now().Before(cli.accessTokenExpirationTime) {
		return
	}

	query := url.Values{}
	query.Set("grant_type", "client_credential")
	query.Set("appid", cli.appID)
	query.Set("secret", cli.appSecret)

	tokenURL := &url.URL{
		Scheme: "https",
		Host: "api.weixin.qq.com",
		Path: "/cgi-bin/token",
		RawQuery: query.Encode(),
	}

	statusCd, body, err := cli.httpClient.Get(nil, tokenURL.String())
	if err != nil {
		panic(err)
	}
	if statusCd != 200 {
		panic(errors.New("获取微信接口调用凭据HTTP调用错误"))
	}

	res, err := gabs.ParseJSON(body)
	if err != nil {
		panic(err)
	}
	if res.ExistsP("errcode") {
		panic(errors.New(res.String()))
	}

	cli.accessToken = res.Path("access_token").Data().(string)
	cli.accessTokenExpirationTime = time.Now().Add((7200 - 200) * time.Second)
}
