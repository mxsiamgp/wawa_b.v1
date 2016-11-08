package business

import (
	"errors"

	competition_domain "wawa_b.v1/module/competition/domain"
	order_business "wawa_b.v1/module/order/business"
	order_domain "wawa_b.v1/module/order/domain"

	"github.com/pquerna/ffjson/ffjson"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// 门票管理器
type DrawnTicketManager interface {
	// 创建订单
	CreateOrder(userID string, competitionID string, tickets []*CreateOrderTicket)

	// 根据赛事ID与用户ID获取所有门票
	GetAllByCompetitionIDAndUserID(competitionID, userID string) []*competition_domain.DrawnTicket

	// 订单项目支付通知回调
	OrderItemPayNotifyCallback() order_business.OrderItemPayNotifyCallback
}

type CreateOrderTicket struct {
	TicketID string `json:"ticket_id"`
	Quantity int `json:"quantity"`
}

// MongoDB门票管理器
type MongoDBDrawnTicketManager struct {
	// 门票集合
	drawnTicketCollection *mgo.Collection

	// 赛事管理器
	competitionManager    CompetitionManager

	// 订单管理器
	orderManager          order_business.OrderManager
}

// 创建一个MongoDB门票管理器
func NewMongoDBDrawnTicketManager(mxsDB *mgo.Database, cmptMgr CompetitionManager, orderMgr order_business.OrderManager) *MongoDBDrawnTicketManager {
	return &MongoDBDrawnTicketManager{
		drawnTicketCollection: mxsDB.C("DrawnTickets"),
		competitionManager: cmptMgr,
		orderManager: orderMgr,
	}
}

func (mgr *MongoDBDrawnTicketManager) CreateOrder(userID string, cmptID string, ts []*CreateOrderTicket) {
	orderItems := make([]*order_domain.OrderItem, 0, len(ts))
	cmpt := mgr.competitionManager.Get(cmptID)
	if cmpt == nil {
		panic(errors.New("无效的赛事ID"))
	}

	tickets := map[string]*competition_domain.Ticket{}

	for _, ticket := range cmpt.Tickets {
		tickets[ticket.ID.Hex()] = ticket
	}

	for _, item := range ts {
		t, ok := tickets[item.TicketID]
		if !ok {
			panic(errors.New("无效的门票ID"))
		}

		valBytes, err := ffjson.Marshal(&competition_domain.DrawnTicket{
			CompetitionID: cmpt.ID,
			CompetitionName: cmpt.Name,
			TicketID: t.ID,
			TicketName: t.Name,
			PriceFee: t.PriceFee,
			Quantity: item.Quantity,
		})
		if err != nil {
			panic(err)
		}

		orderItems = append(orderItems, &order_domain.OrderItem{
			SellableType: "COMPETITION.TICKET",
			SellableValue: string(valBytes),
			Quantity: item.Quantity,
			TotalPriceFee: t.PriceFee * item.Quantity,
		})
	}

	mgr.orderManager.Create(userID, orderItems)
}

func (mgr *MongoDBDrawnTicketManager) GetAllByCompetitionIDAndUserID(cmptID, userID string) []*competition_domain.DrawnTicket {
	tickets := make([]*competition_domain.DrawnTicket, 0)
	query := bson.M{
		"competitionId": bson.ObjectIdHex(cmptID),
		"userId": bson.ObjectIdHex(userID),
	}
	if err := mgr.drawnTicketCollection.Find(query).Sort("_id").All(&tickets); err != nil {
		panic(err)
	}
	return tickets
}

func (mgr *MongoDBDrawnTicketManager) OrderItemPayNotifyCallback() order_business.OrderItemPayNotifyCallback {
	return func(orderID, orderItemID string) {
		order := mgr.orderManager.Get(orderID)
		if order == nil {
			return
		}

		var orderItem *order_domain.OrderItem
		for _, item := range order.Items {
			if item.ID.Hex() == orderItemID {
				orderItem = item
				break
			}
		}
		if orderItem == nil || orderItem.SellableType != "COMPETITION.TICKET" {
			return
		}

		ticket := &competition_domain.DrawnTicket{}
		if err := ffjson.Unmarshal([]byte(orderItem.SellableValue), ticket); err != nil {
			return
		}
		ticket.OrderID = order.ID
		ticket.OrderItemID = orderItem.ID
		ticket.UserID = order.UserID

		insErr := mgr.drawnTicketCollection.Insert(ticket)
		if insErr != nil {
			n, findErr := mgr.drawnTicketCollection.Find(bson.M{
				"orderId": ticket.OrderID,
				"orderItemId": ticket.OrderItemID,
			}).Count()
			if findErr != nil {
				panic(findErr)
			}

			if n == 0 {
				panic(insErr)
			}
		}
	}
}
