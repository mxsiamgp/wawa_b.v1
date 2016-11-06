package service

import (
	"wawa_b.v1/module/competition/business"
	"wawa_b.v1/module/competition/domain"
	"wawa_b.v1/module/rest_json_rpc"

	"github.com/labstack/echo"
)

type DeleteParam struct {
	// ID
	ID string `json:"id"`
}

// 删除一个赛事
func DeleteProcessHandler(cmptMgr business.CompetitionManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*DeleteParam)
		cmptMgr.Delete(param.ID)
		return nil
	}
}

type GetParam struct {
	// ID
	ID string `json:"id"`
}

// 获取一个赛事
func GetProcessHandler(cmptMgr business.CompetitionManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*GetParam)
		return cmptMgr.Get(param.ID)
	}
}

type AddParam struct {
	// 赛事名
	Name    string `json:"name"`

	// 门票集
	Tickets []*domain.Ticket `json:"tickets"`
}

// 添加一个新赛事
func AddProcessHandler(cmptMgr business.CompetitionManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*AddParam)
		if err := cmptMgr.Add(param.Name, param.Tickets); err != nil {
			panic(err)
		}
		return nil
	}
}

type RetrieveParam struct {
	// 最后一个ID
	LastID *string `json:"last_id"`

	// 赛事名
	Name   string `json:"name"`
}

// 检索赛事
func RetrieveProcessHandler(cmptMgr business.CompetitionManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*RetrieveParam)
		return cmptMgr.Retrieve(param.LastID, 15, param.Name)
	}
}

type UpdateTicketsParam struct {
	// ID
	ID      string `json:"id"`

	// 门票集
	Tickets []*domain.Ticket `json:"tickets"`
}

// 更新一个赛事的门票集
func UpdateTicketsProcessHandler(cmptMgr business.CompetitionManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*UpdateTicketsParam)
		cmptMgr.UpdateTickets(param.ID, param.Tickets)
		return nil
	}
}
