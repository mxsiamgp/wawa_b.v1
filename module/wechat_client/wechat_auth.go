package wechat_client

import (
	"errors"
	"net/url"

	"github.com/Jeffail/gabs"
)

// 获取微信授权URL
func (cli *WechatClient) GetAuthorizeURL(redirectURI, scope, state string) string {
	query := url.Values{}
	query.Set("appid", cli.appID)
	query.Set("redirect_uri", redirectURI)
	query.Set("response_type", "code")
	query.Set("scope", scope)
	query.Set("state", state)

	authURL := &url.URL{
		Scheme: "https",
		Host: "open.weixin.qq.com",
		Path: "/connect/oauth2/authorize",
		RawQuery: query.Encode(),
		Fragment: "wechat_redirect",
	}

	return authURL.String()
}

type WechatAccessToken struct {
	// access_token
	AccessToken  string `json:"access_token"`

	// expires_in
	ExpiresIn    int `json:"expires_in"`

	// refresh_token
	RefreshToken string `json:"refresh_token"`

	// openid
	OpenID       string `json:"open_id"`

	// scope
	Scope        string `json:"scope"`
}

// 获取微信网页授权接口调用权限
func (cli *WechatClient) GetAccessToken(code string) *WechatAccessToken {
	query := url.Values{}
	query.Set("appid", cli.appID)
	query.Set("secret", cli.appSecret)
	query.Set("code", code)
	query.Set("grant_type", "authorization_code")

	accTokenURL := url.URL{
		Scheme: "https",
		Host: "api.weixin.qq.com",
		Path: "/sns/oauth2/access_token",
		RawQuery: query.Encode(),
	}

	statusCd, body, err := cli.httpClient.Get(nil, accTokenURL.String())
	if err != nil {
		panic(err)
	}
	if statusCd != 200 {
		panic(errors.New("获取微信网页授权接口调用权限HTTP调用错误"))
	}

	res, err := gabs.ParseJSON(body)
	if err != nil {
		panic(err)
	}
	if res.ExistsP("errcode") {
		panic(errors.New(res.String()))
	}

	return &WechatAccessToken{
		AccessToken: res.Path("access_token").Data().(string),
		ExpiresIn: int(res.Path("expires_in").Data().(float64)),
		RefreshToken: res.Path("refresh_token").Data().(string),
		OpenID: res.Path("openid").Data().(string),
		Scope: res.Path("scope").Data().(string),
	}
}

type WechatUserInfo struct {
	// openid
	OpenID       string `json:"open_id"`

	// nickname
	Nickname     string `json:"nickname"`

	// sex
	Sex          int `json:"sex"`

	// province
	Province     string `json:"province"`

	// city
	City         string `json:"city"`

	// country
	Country      string `json:"country"`

	// headimgurl
	HeadImageURL string `json:"head_image_url"`

	// privilege
	Privilege    []string `json:"privilege"`

	// unionid
	UnionID      *string `json:"union_id"`
}

// 获取微信用户信息
func (cli *WechatClient) GetUserInfo(accToken, openID string) *WechatUserInfo {
	query := url.Values{}
	query.Set("access_token", accToken)
	query.Set("openid", openID)
	query.Set("lang", "zh_CN")

	getUserInfoURL := url.URL{
		Scheme: "https",
		Host: "api.weixin.qq.com",
		Path: "/sns/userinfo",
		RawQuery: query.Encode(),
	}

	statusCd, body, err := cli.httpClient.Get(nil, getUserInfoURL.String())
	if err != nil {
		panic(err)
	}
	if statusCd != 200 {
		panic(errors.New("获取微信用户信息HTTP调用错误"))
	}

	res, err := gabs.ParseJSON(body)
	if err != nil {
		panic(err)
	}
	if res.ExistsP("errcode") {
		panic(errors.New(res.String()))
	}

	userInfo := &WechatUserInfo{
		OpenID: res.Path("openid").Data().(string),
		Nickname: res.Path("nickname").Data().(string),
		Sex: int(res.Path("sex").Data().(float64)),
		Province: res.Path("province").Data().(string),
		City: res.Path("city").Data().(string),
		Country: res.Path("country").Data().(string),
		HeadImageURL: res.Path("headimgurl").Data().(string),
		Privilege: interfaceArray2StringArray(res.Path("privilege").Data().([]interface{})),
	}
	if res.ExistsP("unionid") {
		unionID := res.Path("unionid").Data().(string)
		userInfo.UnionID = &unionID
	}

	return userInfo
}

func interfaceArray2StringArray(a []interface{}) []string {
	r := make([]string, len(a))
	for _, e := range a {
		r = append(r, e.(string))
	}
	return r
}
