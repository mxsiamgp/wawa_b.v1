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
	// 当前用户未绑定商家
	FAIL_CD_CURRENT_USER_UNBOUNDED_MERCHANTS = "MERCHANT.CURRENT_USER_UNBOUNDED_MERCHANTS"

	// 用户不存在
	FAIL_CD_NO_SUCH_USER = "MERCHANT.NO_SUCH_USER"

	// 资源不隶属该商家
	FAIL_CD_RESOURCE_NOT_BELONG_TO_CURRENT_MERCHANT = "MERCHANT.RESOURCE_NOT_BELONG_TO_CURRENT_MERCHANT"
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

// 资源是否属于该商家
type ResourceBelongToMerchant func(ctx echo.Context, param interface{}, merchantID string) bool

// 确保操作属于当前商家的资源
// + 确保登录
func EnsureResourceBelongToCurrentMerchantProcessHandler(mcMgr merchant_business.MerchantManager, rbtm ResourceBelongToMerchant) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, ch *rest_json_rpc.ProcessChain) interface{} {
		sess := session.GetSessionByContext(ctx)
		userID := new(bson.ObjectId)
		if !sess.Get(user_service.SESS_KEY_CURRENT_USER_ID, userID) {
			panic(errors.New("请确保用户已登录"))
		}

		merchant := mcMgr.GetByUserID(userID.Hex())
		if merchant == nil {
			panic(failure.New(FAIL_CD_CURRENT_USER_UNBOUNDED_MERCHANTS))
		}

		if !rbtm(ctx, p, merchant.ID.Hex()) {
			panic(failure.New(FAIL_CD_RESOURCE_NOT_BELONG_TO_CURRENT_MERCHANT))
		}

		return ch.Next()
	}
}

type GetCurrentMerchantParam struct{}

// 获取当前商家
// + 确保登录
func GetCurrentMerchantProcessHandler(mcMgr merchant_business.MerchantManager) rest_json_rpc.ProcessHandler {
	return func(ctx echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		sess := session.GetSessionByContext(ctx)
		userID := new(bson.ObjectId)
		if !sess.Get(user_service.SESS_KEY_CURRENT_USER_ID, userID) {
			panic(errors.New("请确保用户已登录"))
		}

		merchant := mcMgr.GetByUserID(userID.Hex())
		if merchant == nil {
			panic(failure.New(FAIL_CD_CURRENT_USER_UNBOUNDED_MERCHANTS))
		}

		return merchant
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

type KickOutStaffParam struct {
	// 用户ID
	UserID string `json:"user_id"`
}

// 踢出员工
func KickOutStaffProcessHandler(mcMgr merchant_business.MerchantManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*KickOutStaffParam)
		if err := mcMgr.KickOutStaff(param.UserID); err != nil {
			panic(err)
		}
		return nil
	}
}

type PullInStaffParam struct {
	// 商家ID
	MerchantID string `json:"merchant_id"`

	// 用户名
	Name       string `json:"name"`
}

// 拉入员工
func PullInStaffProcessHandler(mcMgr merchant_business.MerchantManager, userMgr user_business.UserManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*PullInStaffParam)

		user := userMgr.GetByName(param.Name)
		if user == nil {
			panic(failure.New(FAIL_CD_NO_SUCH_USER))
		}

		if err := mcMgr.PullInStaff(param.MerchantID, user.ID.Hex()); err != nil {
			panic(err)
		}
		return nil
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
			panic(failure.New(FAIL_CD_NO_SUCH_USER))
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

type RetrieveStaffsParam struct {
	// 最后一个ID
	LastID     *string `json:"last_id"`

	// 商家ID
	MerchantID string `json:"merchant_id"`

	// 员工用户名
	Name       string `json:"name"`
}

// 检索指定商家员工
func RetrieveStaffsProcessHandler(mcMgr merchant_business.MerchantManager) rest_json_rpc.ProcessHandler {
	return func(_ echo.Context, p interface{}, _ *rest_json_rpc.ProcessChain) interface{} {
		param := p.(*RetrieveStaffsParam)
		return mcMgr.RetrieveStaffs(param.LastID, 15, param.MerchantID, param.Name)
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
