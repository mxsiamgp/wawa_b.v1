package business

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"wawa_b.v1/module/log"
	"wawa_b.v1/module/order/domain"
	"wawa_b.v1/module/rest_json_rpc/failure"
	"wawa_b.v1/module/wechat_pay_client"

	"github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// 失败代码
const (
	// 创建微信支付订单失败
	FAIL_CD_CREATE_WECHAT_PAY_ORDER_FAIL = "ORDER.CREATE_WECHAT_PAY_ORDER_FAIL"

	// 订单不存在
	FAIL_CD_NO_SUCH_ORDER = "ORDER.NO_SUCH_ORDER"

	// 订单不是未支付状态
	FAIL_CD_ORDER_STATUS_NOT_BE_UNPAID = "ORDER.ORDER_STATUS_NOT_BE_UNPAID"
)

// 订单管理器
type OrderManager interface {
	// 创建一个新订单
	Create(userID string, items []*domain.OrderItem) string

	// 根据ID获取订单
	Get(id string) *domain.Order

	// 根据用户ID获取所有订单
	GetAllOrdersByUserID(lastID *string, limit int, userID string) []*domain.Order

	// 微信H5支付
	PayByWechatH5(orderID, spbillCreateIP, notifyURL, openID string) (string, error)

	// 微信支付
	WechatPaid(orderID string)
}

// 订单项支付通知回调
type OrderItemPayNotifyCallback func(orderID, orderItemID string)

// MongoDB订单管理器
type MongoDBOrderManager struct {
	// 日志
	logger                              *logrus.Logger

	// 订单集合
	orderCollection                     *mgo.Collection

	// 订单项类型到支付通知回调的映射
	payNotifyCallbackByItemSellableType map[string]OrderItemPayNotifyCallback

	// - 微信支付相关 --BEGIN--
	// 微信支付客户端
	wechatPayClient                     *wechat_pay_client.WechatPayClient

	// 微信H5支付浏览器打开的移动网页的主页标题
	wechatH5PayWebpageTitle             *string
	// - 微信支付相关 --END--
}

// 创建一个MongoDB订单管理器
func NewMongoDBOrderManager(
tradeDB *mgo.Database, phcbist map[string]OrderItemPayNotifyCallback,
wcPayCli *wechat_pay_client.WechatPayClient, wcH5PayWpTitle *string) *MongoDBOrderManager {
	return &MongoDBOrderManager{
		logger: log.GetLogger("order.orderManager"),
		orderCollection: tradeDB.C("Orders"),
		payNotifyCallbackByItemSellableType: phcbist,
		wechatPayClient: wcPayCli,
		wechatH5PayWebpageTitle: wcH5PayWpTitle,
	}
}

func (mgr *MongoDBOrderManager) Create(userID string, items []*domain.OrderItem) string {
	for _, item := range items {
		item.ID = bson.NewObjectId()
	}

	price := 0
	for _, item := range items {
		price += item.TotalPriceFee
	}

	id := bson.NewObjectId()

	if err := mgr.orderCollection.Insert(&domain.Order{
		ID: id,
		UserID: bson.ObjectIdHex(userID),
		CreatedTime: time.Now(),
		Items: items,
		TotalPriceFee: price,
		Status: domain.ORD_STAT_UNPAID,
	}); err != nil {
		panic(err)
	}

	return id.Hex()
}

func (mgr *MongoDBOrderManager) Get(id string) *domain.Order {
	orders := make([]*domain.Order, 0)
	if err := mgr.orderCollection.FindId(bson.ObjectIdHex(id)).All(&orders); err != nil {
		panic(err)
	}
	if len(orders) == 0 {
		return nil
	}
	return orders[0]
}

func (mgr *MongoDBOrderManager) GetAllOrdersByUserID(lastID *string, limit int, userID string) []*domain.Order {
	orders := make([]*domain.Order, 0)
	query := bson.M{
		"userId": bson.ObjectIdHex(userID),
	}
	if lastID != nil {
		query["_id"] = bson.M{
			"$lt": bson.ObjectIdHex(*lastID),
		}
	}
	if err := mgr.orderCollection.Find(query).Sort("userId", "-createdTime").Limit(limit).All(&orders); err != nil {
		panic(err)
	}
	return orders
}

func (mgr *MongoDBOrderManager) PayByWechatH5(orderID, spbillCreateIP, notifyURL, openID string) (string, error) {
	order := mgr.Get(orderID)
	if order == nil {
		return "", failure.New(FAIL_CD_NO_SUCH_ORDER)
	}

	if order.Status != domain.ORD_STAT_UNPAID {
		return "", failure.New(FAIL_CD_ORDER_STATUS_NOT_BE_UNPAID)
	}

	if order.WechatPayOutTradeNo != nil {
		res, reqBody, resBody := mgr.wechatPayClient.CloseOrder(*order.WechatPayOutTradeNo)
		if res.Path("return_code").Data().(string) != "SUCCESS" || res.ExistsP("err_code") {
			mgr.logger.WithFields(logrus.Fields{
				"requestBody": string(reqBody),
				"responseBody": string(resBody),
			}).Error("关闭微信支付订单失败")
		}

		if res.Path("return_code").Data().(string) != "SUCCESS" {
			panic(errors.New("关闭微信支付订单失败"))
		}

		// 关闭订单错误
		if res.ExistsP("err_code") {
			// 订单已经支付状态修复
			if res.Path("err_code").Data().(string) == "ORDERPAID" {
				mgr.WechatPaid(orderID)
			}

			return "", failure.New(FAIL_CD_CREATE_WECHAT_PAY_ORDER_FAIL)
		}
	}

	outTradeNo := strings.Replace(uuid.NewV4().String(), "-", "", -1)

	if err := mgr.orderCollection.Update(bson.M{
		"_id": bson.ObjectIdHex(orderID),
	}, bson.M{
		"$set": bson.M{
			"wechatPayOutTradeNo": outTradeNo,
		},
	}); err != nil {
		panic(err)
	}

	res, reqBody, resBody := mgr.wechatPayClient.UnifiedOrder(map[string]string{
		"body": fmt.Sprintf("%s-%s", *mgr.wechatH5PayWebpageTitle, orderID),
		"attach": orderID,
		"out_trade_no": outTradeNo,
		"total_fee": strconv.Itoa(order.TotalPriceFee),
		"spbill_create_ip": spbillCreateIP,
		"notify_url": notifyURL,
		"trade_type": "JSAPI",
		"openid": openID,
	})

	if res.Path("return_code").Data().(string) != "SUCCESS" || res.Path("result_code").Data().(string) != "SUCCESS" {
		mgr.logger.WithFields(logrus.Fields{
			"requestBody": string(reqBody),
			"responseBody": string(resBody),
		}).Error("创建微信支付订单失败")
		return "", failure.New(FAIL_CD_CREATE_WECHAT_PAY_ORDER_FAIL)
	}

	return res.Path("prepay_id").Data().(string), nil
}

func (mgr *MongoDBOrderManager) WechatPaid(orderID string) {
	// 调用订单项支付通知回调
	order := mgr.Get(orderID)
	for _, item := range order.Items {
		if cb, ok := mgr.payNotifyCallbackByItemSellableType[item.SellableType]; ok {
			cb(orderID, item.ID.Hex())
		}
	}

	if err := mgr.orderCollection.UpdateId(bson.ObjectIdHex(orderID), bson.M{
		"$set": bson.M{
			"status": domain.ORD_STAT_PAID,
			"payApproach": domain.ORD_PAY_APPROACH_WECHAT,
		},
	}); err != nil {
		panic(err)
	}
}
