package service

import (
	"wawa_b.v1/module/rest_json_rpc"
	"wawa_b.v1/module/merchant/business"

	"github.com/labstack/echo"
)

type DeleteParam struct {
	// ID
	ID string `json:"id"`
}

// 删除一个商家
func DeleteProcessHandler(mcMgr business.MerchantManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*DeleteParam)
		mcMgr.Delete(param.ID)
		return nil
	}
}

type GetParam struct {
	// ID
	ID string `json:"id"`
}

// 获取一个用户
func GetProcessHandler(mcMgr business.MerchantManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*GetParam)
		return mcMgr.Get(param.ID)
	}
}

type RegisterParam struct {
	// 商家店名
	Name            string `json:"name"`

	// 经营项目
	ItemsOfBusiness string `json:"items_of_business"`

	// 联系人姓名
	ContactsName    string `json:"contacts_name"`

	// 联系人手机号码
	ContactsMobile  string `json:"contacts_mobile"`

	// 联系人身份证号码
	ContactsIDCard  string `json:"contacts_id_card"`

	// 联系人地址
	ContactsAddress string `json:"contacts_address"`
}

// 注册一个新商家
func RegisterProcessHandler(mcMgr business.MerchantManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*RegisterParam)
		mcMgr.Register(param.Name, param.ItemsOfBusiness, param.ContactsName, param.ContactsMobile, param.ContactsIDCard, param.ContactsAddress)
		return nil
	}
}

type RetrieveParam struct {
	// 最后一个ID
	LastID   *string `json:"last_id"`

	// 商家店名
	Name     string `json:"name"`
}

// 检索商家
func RetrieveProcessHandler(mcMgr business.MerchantManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*RetrieveParam)
		return mcMgr.Retrieve(param.LastID, 15, param.Name)
	}
}

type UpdateParam struct {
	// ID
	ID       string `json:"id"`

	// 商家店名
	Name            string `json:"name"`

	// 经营项目
	ItemsOfBusiness string `json:"items_of_business"`

	// 联系人姓名
	ContactsName    string `json:"contacts_name"`

	// 联系人手机号码
	ContactsMobile  string `json:"contacts_mobile"`

	// 联系人身份证号码
	ContactsIDCard  string `json:"contacts_id_card"`

	// 联系人地址
	ContactsAddress string `json:"contacts_address"`
}

// 更新一个商家的基本信息
func UpdateProcessHandler(mcMgr business.MerchantManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*UpdateParam)
		mcMgr.Update(param.ID, param.Name, param.ItemsOfBusiness, param.ContactsName, param.ContactsMobile, param.ContactsIDCard, param.ContactsAddress)
		return nil
	}
}
