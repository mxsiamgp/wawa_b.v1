package service

import (
	competition_business "wawa_b.v1/module/competition/business"
	competition_domain "wawa_b.v1/module/competition/domain"
	"wawa_b.v1/module/rest_json_rpc"
	"wawa_b.v1/module/session"
	user_service "wawa_b.v1/module/user/service"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

type AddParam struct {
	// 赛事名
	Name    string `json:"name"`

	// 门票集
	Tickets []*competition_domain.Ticket `json:"tickets"`
}

// 添加一个新赛事
func AddProcessHandler(cmptMgr competition_business.CompetitionManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*AddParam)
		if err := cmptMgr.Add(param.Name, param.Tickets); err != nil {
			panic(err)
		}
		return nil
	}
}

type CreateOrderParam struct {
	CompetitionID string `json:"competition_id"`
	Tickets       []*competition_business.CreateOrderTicket `json:"tickets"`
}

// 创建一个订单
// + 确保登录
func CreateOrderProcessHandler(ticketMgr competition_business.DrawnTicketManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*CreateOrderParam)

		sess := session.GetSessionByContext(ctx)
		userID := new(bson.ObjectId)
		if !sess.Get(user_service.SESS_KEY_CURRENT_USER_ID, userID) {
			panic(errors.New("请确保用户已登录"))
		}

		ticketMgr.CreateOrder(userID.Hex(), param.CompetitionID, param.Tickets)

		return nil
	}
}

type DeleteParam struct {
	// ID
	ID string `json:"id"`
}

// 删除一个赛事
func DeleteProcessHandler(cmptMgr competition_business.CompetitionManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*DeleteParam)
		cmptMgr.Delete(param.ID)
		return nil
	}
}

type FinishParam struct {
	// ID
	ID string `json:"id"`
}

// 完成一个赛事
func FinishProcessHandler(cmptMgr competition_business.CompetitionManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*FinishParam)
		cmptMgr.Finish(param.ID)
		return nil
	}
}

type GetParam struct {
	// ID
	ID string `json:"id"`
}

// 获取一个赛事
func GetProcessHandler(cmptMgr competition_business.CompetitionManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*GetParam)
		return cmptMgr.Get(param.ID)
	}
}

type ListInProgressParam struct{}

// 列出正在进行的赛事
func ListInProgressProcessHandler(cmptMgr competition_business.CompetitionManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		return cmptMgr.ListInProgress()
	}
}

type RetrieveParam struct {
	// 最后一个ID
	LastID *string `json:"last_id"`

	// 赛事名
	Name   string `json:"name"`
}

// 检索赛事
func RetrieveProcessHandler(cmptMgr competition_business.CompetitionManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*RetrieveParam)
		return cmptMgr.Retrieve(param.LastID, 15, param.Name)
	}
}

type UpdateTicketsParam struct {
	// ID
	ID      string `json:"id"`

	// 门票集
	Tickets []*competition_domain.Ticket `json:"tickets"`
}

// 更新一个赛事的门票集
func UpdateTicketsProcessHandler(cmptMgr competition_business.CompetitionManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*UpdateTicketsParam)
		cmptMgr.UpdateTickets(param.ID, param.Tickets)
		return nil
	}
}
