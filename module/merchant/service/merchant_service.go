package service

import (
	"errors"

	"wawa_b.v1/module/rest_json_rpc"
	merchant_business "wawa_b.v1/module/merchant/business"
	"wawa_b.v1/module/rest_json_rpc/failure"
	"wawa_b.v1/module/session"
	user_business "wawa_b.v1/module/user/business"
	user_service "wawa_b.v1/module/user/service"

	"github.com/labstack/echo"
	"gopkg.in/mgo.v2/bson"
)

// 失败代码
const (
	// 管理员用户不存在
	FAIL_CD_NO_SUCH_MANAGER_USER = "MERCHANT.NO_SUCH_MANAGER_USER"
)

type DeleteParam struct {
	// ID
	ID string `json:"id"`
}

// 删除一个商家
func DeleteProcessHandler(mcMgr merchant_business.MerchantManager) rest_json_rpc.ProcessHandler {
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
func GetProcessHandler(mcMgr merchant_business.MerchantManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*GetParam)
		return mcMgr.Get(param.ID)
	}
}

type RegisterParam struct {
	// 商家店名
	Name            string `json:"name"`

	// 管理员用户名
	ManagerUserName string `json:"manager_user_name"`

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
// + 确保登录
func RegisterProcessHandler(mcMgr merchant_business.MerchantManager, userMgr user_business.UserManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*RegisterParam)

		mgrUser := userMgr.GetByName(param.ManagerUserName)
		if mgrUser == nil {
			panic(failure.New(FAIL_CD_NO_SUCH_MANAGER_USER))
		}

		sess := session.GetSessionByContext(ctx)
		userID := new(bson.ObjectId)
		if !sess.Get(user_service.SESS_KEY_CURRENT_USER_ID, userID) {
			panic(errors.New("请确保用户已登录"))
		}

		if err := mcMgr.Register(userID.Hex(), param.Name, mgrUser.ID.Hex(), param.ItemsOfBusiness, param.ContactsName, param.ContactsMobile, param.ContactsIDCard, param.ContactsAddress); err != nil {
			panic(err)
		}
		return nil
	}
}

type RetrieveParam struct {
	// 最后一个ID
	LastID *string `json:"last_id"`

	// 商家店名
	Name   string `json:"name"`
}

// 检索商家
func RetrieveProcessHandler(mcMgr merchant_business.MerchantManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*RetrieveParam)
		return mcMgr.Retrieve(param.LastID, 15, param.Name)
	}
}

type UpdateParam struct {
	// ID
	ID              string `json:"id"`

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
func UpdateProcessHandler(mcMgr merchant_business.MerchantManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*UpdateParam)
		mcMgr.Update(param.ID, param.Name, param.ItemsOfBusiness, param.ContactsName, param.ContactsMobile, param.ContactsIDCard, param.ContactsAddress)
		return nil
	}
}
