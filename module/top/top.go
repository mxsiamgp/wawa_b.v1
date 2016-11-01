package top

import (
	"bytes"
	"errors"
	"sort"
	"strings"
	"time"

	"wawa_b.v1/module/md5"

	"github.com/Jeffail/gabs"
	"github.com/valyala/fasthttp"
)

// 调用环境URL
const (
	HTTP_SANDBOX = "http://gw.api.tbsandbox.com/router/rest"
	HTTP_OFFICIAL = "http://gw.api.taobao.com/router/rest"
	HTTP_INTERNATIONAL = "http://api.taobao.com/router/rest"
)

// TOP客户端
type TOPClient struct {
	// HTTP客户端
	httpClient *fasthttp.Client

	// 调用环境URL
	gateway    string

	// AppKey
	appKey     string

	// AppSecret
	appSecret  string
}

// 创建一个TOP客户端
func NewTOPClient(httpCli *fasthttp.Client, gw, appKey, appSecret string) *TOPClient {
	cli := &TOPClient{
		httpClient: httpCli,
		gateway: gw,
		appKey: appKey,
		appSecret: appSecret,
	}
	return cli
}

// 以POST调用方法
func (cli *TOPClient) POST(method string, specificParams map[string]string) (*gabs.Container, []byte, []byte) {
	params := make(map[string]string)

	// 收集参数
	//  - 收集公共参数
	params["method"] = method
	params["app_key"] = cli.appKey
	params["timestamp"] = time.Now().Format("2006-01-02 15:04:05")
	params["format"] = "json"
	params["v"] = "2.0"
	params["sign_method"] = "md5"

	// - 收集特定参数
	for key, val := range specificParams {
		params[key] = val
	}

	// 生成签名
	// - 生成有序的参数键序列
	orderedParamKeys := make([]string, 0, len(params))
	for key := range params {
		orderedParamKeys = append(orderedParamKeys, key)
	}
	sort.Strings(orderedParamKeys)

	// - 构建签名明文
	signPlaintextBuf := bytes.NewBufferString(cli.appSecret)
	for _, key := range orderedParamKeys {
		signPlaintextBuf.WriteString(key)
		signPlaintextBuf.WriteString(params[key])
	}
	signPlaintextBuf.WriteString(cli.appSecret)

	// - 生成签名
	sign := strings.ToUpper(md5.StringDigest(signPlaintextBuf.Bytes()))
	params["sign"] = sign

	// 生成表单
	postArgs := fasthttp.AcquireArgs()
	defer fasthttp.ReleaseArgs(postArgs)
	for key, val := range params {
		postArgs.Add(key, val)
	}

	// 发送请求
	statusCd, body, err := cli.httpClient.Post(nil, cli.gateway, postArgs)
	if err != nil {
		panic(err)
	}
	if statusCd != 200 {
		panic(errors.New("TOP客户端HTTP调用错误"))
	}

	result, err := gabs.ParseJSON(body)
	if err != nil {
		panic(err)
	}
	return result, postArgs.QueryString(), body
}
