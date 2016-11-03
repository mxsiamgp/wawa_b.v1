package business

import (
	"wawa_b.v1/module/js_regex"
	"wawa_b.v1/module/merchant/domain"
	"wawa_b.v1/module/rest_json_rpc/failure"
	"wawa_b.v1/module/user/business"
	"wawa_b.v1/module/user/domain/permission"

	"github.com/fatih/set"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// 失败代码
const (
	// 管理员用户已经绑定了商家
	FAIL_CD_MANAGER_USER_IS_BOUND = "MERCHANT.MANAGER_USER_BOUND"

	// 商家店名重复
	FAIL_CD_DUPLICATE_MERCHANT_NAME = "MERCHANT.DUPLICATE_MERCHANT_NAME"
)

var MERCHANT_PERMISSIONS = []string{
	"USER.MERCHANT_STAFF.RETRIEVE",
	"USER.MERCHANT_STAFF.MODIFY",
}

// 商家管理器
type MerchantManager interface {
	// 删除一个商家
	Delete(id string)

	// 获取一个商家
	Get(id string) *domain.Merchant

	// 根据名称获取一个商家
	GetByName(name string) *domain.Merchant

	// 根据管理员用户ID获取一个商家
	GetByManagerUserID(mgrUserID string) *domain.Merchant

	// 注册一个新商家
	Register(userID, name, managerUserID, itemsOfBusiness, contactsName, contactsMobile, contactsIDCard, contactsAddress string) error

	// 商家检索
	Retrieve(lastID *string, limit int, name string) []*domain.Merchant

	// 更新一个商家的基本信息
	Update(id, name, itemsOfBusiness, contactsName, contactsMobile, contactsIDCard, contactsAddress string)
}

// MongoDB商家管理器
type MongoDBMerchantManager struct {
	// 商家集合
	merchantCollection *mgo.Collection

	// 用户管理器
	userManager        business.UserManager
}

// 创建一个MongoDB用户管理器
func NewMongoDBMerchantManager(mxsDB *mgo.Database, userMgr business.UserManager) *MongoDBMerchantManager {
	return &MongoDBMerchantManager{
		merchantCollection: mxsDB.C("Merchants"),
		userManager: userMgr,
	}
}

func (mgr *MongoDBMerchantManager) Delete(id string) {
	if err := mgr.merchantCollection.RemoveId(bson.ObjectIdHex(id)); err != nil {
		panic(err)
	}
}

func (mgr *MongoDBMerchantManager) Get(id string) *domain.Merchant {
	merchants := make([]*domain.Merchant, 0)
	if err := mgr.merchantCollection.FindId(bson.ObjectIdHex(id)).All(&merchants); err != nil {
		panic(err)
	}
	if len(merchants) == 0 {
		return nil
	}
	return merchants[0]
}

func (mgr *MongoDBMerchantManager) GetByName(name string) *domain.Merchant {
	merchants := make([]*domain.Merchant, 0)
	if err := mgr.merchantCollection.Find(bson.M{
		"name": name,
	}).All(&merchants); err != nil {
		panic(err)
	}
	if len(merchants) == 0 {
		return nil
	}
	return merchants[0]
}

func (mgr *MongoDBMerchantManager) GetByManagerUserID(managerUserID string) *domain.Merchant {
	merchants := make([]*domain.Merchant, 0)
	if err := mgr.merchantCollection.Find(bson.M{
		"managerUserId": bson.ObjectIdHex(managerUserID),
	}).All(&merchants); err != nil {
		panic(err)
	}
	if len(merchants) == 0 {
		return nil
	}
	return merchants[0]
}

func (mgr *MongoDBMerchantManager) Register(userID, name, managerUserID, itemsOfBusiness, contactsName, contactsMobile, contactsIDCard, contactsAddress string) error {
	if mgr.GetByName(name) != nil {
		return failure.New(FAIL_CD_DUPLICATE_MERCHANT_NAME)
	}

	if mgr.GetByManagerUserID(managerUserID) != nil {
		return failure.New(FAIL_CD_MANAGER_USER_IS_BOUND)
	}

	mgrUser := mgr.userManager.Get(managerUserID)

	if err := mgr.merchantCollection.Insert(&domain.Merchant{
		Name: name,
		ManagerUserID: mgrUser.ID,
		ManagerUserName: mgrUser.Name,
		ItemsOfBusiness: itemsOfBusiness,
		ContactsName: contactsName,
		ContactsMobile: contactsMobile,
		ContactsIDCard: contactsIDCard,
		ContactsAddress: contactsAddress,
	}); err != nil {
		panic(err)
	}

	permSet := permission.PermissionSet(mgrUser.FlatPermissions)
	permSet.Add("USER.MERCHANT_STAFF.RETRIEVE", "USER.MERCHANT_STAFF.MODIFY")

	if err := mgr.userManager.GrantFlatPermissions(userID, managerUserID, set.StringSlice(permSet)); err != nil {
		return err
	}

	return nil
}

func (mgr *MongoDBMerchantManager) Retrieve(lastID *string, limit int, name string) []*domain.Merchant {
	merchants := make([]*domain.Merchant, 0)
	query := bson.M{}
	if lastID != nil {
		query["_id"] = bson.M{
			"$gt": bson.ObjectIdHex(*lastID),
		}
	}
	if len(name) != 0 {
		query["name"] = bson.RegEx{
			Pattern: js_regex.EscapeTextPattern(name),
		}
	}
	if err := mgr.merchantCollection.Find(query).Sort("_id").Limit(limit).All(&merchants); err != nil {
		panic(err)
	}
	return merchants
}

func (mgr *MongoDBMerchantManager) Update(id, name, itemsOfBusiness, contactsName, contactsMobile, contactsIDCard, contactsAddress string) {
	if err := mgr.merchantCollection.UpdateId(bson.ObjectIdHex(id), bson.M{
		"$set": bson.M{
			"name": name,
			"itemsOfBusiness": itemsOfBusiness,
			"contactsName": contactsName,
			"contactsMobile": contactsMobile,
			"contactsIDCard": contactsIDCard,
			"contactsAddress": contactsAddress,
		},
	}); err != nil {
		panic(err)
	}
}
