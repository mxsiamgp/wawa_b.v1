package service

import (
	"wawa_b.v1/module/competition/business"
	"wawa_b.v1/module/rest_json_rpc"

	"github.com/labstack/echo"
)

type MarkInspectedParam struct {
	CompetitionID string `json:"competition_id"`
	UserID        string `json:"user_id"`
}

// 根据赛事ID与用户ID获取所有门票
func MarkInspectedProcessHandler(inspectMgr business.InspectionManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*MarkInspectedParam)
		if err := inspectMgr.MarkInspected(param.CompetitionID, param.UserID); err != nil {
			panic(err)
		}
		return nil
	}
}
