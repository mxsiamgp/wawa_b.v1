package wechat_pay_client

import (
	"bytes"
	"errors"
	"sort"
	"strings"

	"wawa_b.v1/module/md5"

	"github.com/Jeffail/gabs"
	"github.com/clbanning/mxj"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/valyala/fasthttp"
)

// 微信支付客户端
type WechatPayClient struct {
	// HTTP客户端
	httpClient *fasthttp.Client

	// AppID
	appID      string

	// 微信支付商户号
	mchID      string

	// 微信支付API秘钥
	partnerKey string
}

// 创建一个微信支付客户端
func NewWechatPayClient(httpCli *fasthttp.Client, appID, mchID, ptrKey string) *WechatPayClient {
	return &WechatPayClient{
		httpClient: httpCli,
		appID: appID,
		mchID: mchID,
		partnerKey: ptrKey,
	}
}

// 统一下单
func (cli *WechatPayClient) UnifiedOrder(params map[string]string) (*gabs.Container, []byte, []byte) {
	finalParams := map[string]string{}

	// 收集参数
	// - 收集公共参数
	finalParams["appid"] = cli.appID
	finalParams["mch_id"] = cli.mchID
	finalParams["nonce_str"] = strings.Replace(uuid.NewV4().String(), "-", "", -1)

	// - 收集特定参数
	for key, val := range params {
		finalParams[key] = val
	}

	finalParams["sign"] = cli.Sign(finalParams)

	// 生成请求体
	reqBody, err := mxj.Map(stringMap2InterfaceMap(finalParams)).Xml("xml")
	if err != nil {
		panic(err)
	}

	// 发送请求
	// - 构建请求响应
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethodBytes([]byte("POST"))
	req.SetRequestURI("https://api.mch.weixin.qq.com/pay/unifiedorder")
	req.Header.SetContentTypeBytes([]byte("text/xml"))
	req.BodyWriter().Write(reqBody)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	// - 执行发送
	if err := cli.httpClient.Do(req, res); err != nil {
		panic(err)
	}
	if res.StatusCode() != 200 {
		panic(errors.New("微信支付统一下单HTTP调用错误"))
	}

	// 解析响应
	// - 解析XML
	resMap, err := mxj.NewMapXml(res.Body())
	if err != nil {
		panic(err)
	}

	// - XML转JSON
	resJSON, err := resMap.Json()
	if err != nil {
		panic(err)
	}
	result, err := gabs.ParseJSON(resJSON)
	if err != nil {
		panic(err)
	}

	return result.S("xml"), reqBody, res.Body()
}

// 关闭订单
func (cli *WechatPayClient) CloseOrder(outTradeNo string) (*gabs.Container, []byte, []byte) {
	finalParams := map[string]string{}

	// 收集参数
	// - 收集公共参数
	finalParams["appid"] = cli.appID
	finalParams["mch_id"] = cli.mchID
	finalParams["out_trade_no"] = outTradeNo
	finalParams["nonce_str"] = strings.Replace(uuid.NewV4().String(), "-", "", -1)

	finalParams["sign"] = cli.Sign(finalParams)

	// 生成请求体
	reqBody, err := mxj.Map(stringMap2InterfaceMap(finalParams)).Xml("xml")
	if err != nil {
		panic(err)
	}

	// 发送请求
	// - 构建请求响应
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethodBytes([]byte("POST"))
	req.SetRequestURI("https://api.mch.weixin.qq.com/pay/closeorder")
	req.Header.SetContentTypeBytes([]byte("text/xml"))
	req.BodyWriter().Write(reqBody)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	// - 执行发送
	if err := cli.httpClient.Do(req, res); err != nil {
		panic(err)
	}
	if res.StatusCode() != 200 {
		panic(errors.New("微信支付关闭订单HTTP调用错误"))
	}

	// 解析响应
	// - 解析XML
	resMap, err := mxj.NewMapXml(res.Body())
	if err != nil {
		panic(err)
	}

	// - XML转JSON
	resJSON, err := resMap.Json()
	if err != nil {
		panic(err)
	}
	result, err := gabs.ParseJSON(resJSON)
	if err != nil {
		panic(err)
	}

	return result.S("xml"), reqBody, res.Body()
}

// 生成签名
func (cli *WechatPayClient) Sign(params map[string]string) string {
	// 生成有序的参数键序列
	orderedParamKeys := make([]string, 0, len(params))
	for key := range params {
		orderedParamKeys = append(orderedParamKeys, key)
	}
	sort.Strings(orderedParamKeys)

	// 构建签名明文
	signPlaintextBuf := bytes.NewBufferString("")
	for i, key := range orderedParamKeys {
		if i != 0 {
			signPlaintextBuf.WriteString("&")
		}
		signPlaintextBuf.WriteString(key)
		signPlaintextBuf.WriteString("=")
		signPlaintextBuf.WriteString(params[key])
	}
	if signPlaintextBuf.Len() != 0 {
		signPlaintextBuf.WriteString("&")
	}
	signPlaintextBuf.WriteString("key=")
	signPlaintextBuf.WriteString(cli.partnerKey)

	// 生成签名
	return strings.ToUpper(md5.StringDigest(signPlaintextBuf.Bytes()))
}

func stringMap2InterfaceMap(m map[string]string) map[string]interface{} {
	r := map[string]interface{}{}
	for k, v := range m {
		r[k] = v
	}
	return r
}

type PayNotifyCallbackHandler func(param *gabs.Container)

// 微信支付通知回调
func (cli *WechatPayClient) PayNotifyCallbackHandlerFunc(handler PayNotifyCallbackHandler) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		defer func() {
			// 响应请求
			// - 构建响应
			resMap := mxj.New()
			if err := resMap.SetValueForPath("SUCCESS", "return_code"); err != nil {
				panic(err)
			}
			if err := resMap.SetValueForPath("OK", "return_msg"); err != nil {
				panic(err)
			}

			// - 生成XML
			resXML, err := resMap.Xml("xml")
			if err != nil {
				panic(err)
			}

			// - 发送响应
			if err := ctx.Stream(200, "text/xml", bytes.NewReader(resXML)); err != nil {
				panic(err)
			}
		}()

		// 解析请求
		// - 解析XML
		reqMap, err := mxj.NewMapXmlReader(ctx.Request().Body())
		if err != nil {
			panic(err)
		}

		// - XML转JSON
		reqJSON, err := reqMap.Json()
		if err != nil {
			panic(err)
		}
		param, err := gabs.ParseJSON(reqJSON)
		if err != nil {
			panic(err)
		}

		handler(param.S("xml"))

		return nil
	}
}
