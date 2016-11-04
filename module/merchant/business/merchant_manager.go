package business

import (
	"wawa_b.v1/module/js_regex"
	merchant_domain "wawa_b.v1/module/merchant/domain"
	"wawa_b.v1/module/rest_json_rpc/failure"
	"wawa_b.v1/module/user/business"
	user_domain "wawa_b.v1/module/user/domain"
	"wawa_b.v1/module/user/domain/permission"

	"github.com/fatih/set"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// 失败代码
const (
	// 不能踢出管理员
	FAIL_CD_CANNOT_KICK_OUT_MANAGER = "MERCHANT.CANNOT_KICK_OUT_MANAGER"

	// 商家店名重复
	FAIL_CD_DUPLICATE_MERCHANT_NAME = "MERCHANT.DUPLICATE_MERCHANT_NAME"

	// 用户已经绑定了商家
	FAIL_CD_USER_IS_BOUND = "MERCHANT.USER_IS_BOUND"
)

// 商家管理器
type MerchantManager interface {
	// 删除一个商家
	Delete(id string)

	// 获取一个商家
	Get(id string) *merchant_domain.Merchant

	// 根据名称获取一个商家
	GetByName(name string) *merchant_domain.Merchant

	// 根据用户ID获取一个商家
	GetByUserID(userID string) *merchant_domain.Merchant

	// 踢出员工
	KickOutStaff(userID string) error

	// 拉入员工
	PullInStaff(merchantID, userID string) error

	// 注册一个新商家
	Register(userID, name, managerUserID, itemsOfBusiness, contactsName, contactsMobile, contactsIDCard, contactsAddress string) error

	// 检索商家
	Retrieve(lastID *string, limit int, name string) []*merchant_domain.Merchant

	// 检索指定商家员工
	RetrieveStaffs(lastID *string, limit int, merchantID string, name string) []*user_domain.User

	// 更新一个商家的基本信息
	Update(id, name, itemsOfBusiness, contactsName, contactsMobile, contactsIDCard, contactsAddress string)
}

// MongoDB商家管理器
type MongoDBMerchantManager struct {
	// 商家集合
	merchantCollection *mgo.Collection

	// 用户集合
	userCollection     *mgo.Collection

	// 用户管理器
	userManager        business.UserManager
}

// 创建一个MongoDB用户管理器
func NewMongoDBMerchantManager(mxsDB *mgo.Database, userMgr business.UserManager) *MongoDBMerchantManager {
	return &MongoDBMerchantManager{
		merchantCollection: mxsDB.C("Merchants"),
		userCollection: mxsDB.C("Users"),
		userManager: userMgr,
	}
}

func (mgr *MongoDBMerchantManager) Delete(id string) {
	if err := mgr.merchantCollection.RemoveId(bson.ObjectIdHex(id)); err != nil {
		panic(err)
	}
}

func (mgr *MongoDBMerchantManager) Get(id string) *merchant_domain.Merchant {
	merchants := make([]*merchant_domain.Merchant, 0)
	if err := mgr.merchantCollection.FindId(bson.ObjectIdHex(id)).All(&merchants); err != nil {
		panic(err)
	}
	if len(merchants) == 0 {
		return nil
	}
	return merchants[0]
}

func (mgr *MongoDBMerchantManager) GetByName(name string) *merchant_domain.Merchant {
	merchants := make([]*merchant_domain.Merchant, 0)
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

func (mgr *MongoDBMerchantManager) GetByUserID(userID string) *merchant_domain.Merchant {
	merchants := make([]*merchant_domain.Merchant, 0)
	if err := mgr.merchantCollection.Find(bson.M{
		"$or": []interface{}{
			bson.M{
				"managerUserId": bson.ObjectIdHex(userID),
			},
			bson.M{
				"staffUserIds": bson.ObjectIdHex(userID),
			},
		},
	}).All(&merchants); err != nil {
		panic(err)
	}
	if len(merchants) == 0 {
		return nil
	}
	return merchants[0]
}

func (mgr *MongoDBMerchantManager) KickOutStaff(userID string) error {
	mc := mgr.GetByUserID(userID)
	if mc == nil {
		return nil
	}

	if mc.ManagerUserID.Hex() == userID {
		return failure.New(FAIL_CD_CANNOT_KICK_OUT_MANAGER)
	}

	if err := mgr.merchantCollection.Update(bson.M{
		"_id": mc.ID,
	}, bson.M{
		"$pull": bson.M{
			"staffUserIds": bson.ObjectIdHex(userID),
		},
	}); err != nil {
		panic(err)
	}

	return nil
}

func (mgr *MongoDBMerchantManager) PullInStaff(merchantID, userID string) error {
	if mgr.GetByUserID(userID) != nil {
		return failure.New(FAIL_CD_USER_IS_BOUND)
	}

	if err := mgr.merchantCollection.Update(bson.M{
		"_id": bson.ObjectIdHex(merchantID),
	}, bson.M{
		"$addToSet": bson.M{
			"staffUserIds": bson.ObjectIdHex(userID),
		},
	}); err != nil {
		panic(err)
	}

	return nil
}

func (mgr *MongoDBMerchantManager) Register(userID, name, managerUserID, itemsOfBusiness, contactsName, contactsMobile, contactsIDCard, contactsAddress string) error {
	if mgr.GetByName(name) != nil {
		return failure.New(FAIL_CD_DUPLICATE_MERCHANT_NAME)
	}

	if mgr.GetByUserID(managerUserID) != nil {
		return failure.New(FAIL_CD_USER_IS_BOUND)
	}

	mgrUser := mgr.userManager.Get(managerUserID)

	if err := mgr.merchantCollection.Insert(&merchant_domain.Merchant{
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
	permSet.Add("MERCHANT.STAFF.RETRIEVE", "MERCHANT.STAFF.MODIFY")

	if err := mgr.userManager.GrantFlatPermissions(userID, managerUserID, set.StringSlice(permSet)); err != nil {
		return err
	}

	return nil
}

func (mgr *MongoDBMerchantManager) Retrieve(lastID *string, limit int, name string) []*merchant_domain.Merchant {
	merchants := make([]*merchant_domain.Merchant, 0)
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

func (mgr *MongoDBMerchantManager) RetrieveStaffs(lastID *string, limit int, merchantID string, name string) []*user_domain.User {
	users := make([]*user_domain.User, 0)
	query := bson.M{}

	mc := mgr.Get(merchantID)

	idFilter := bson.M{}
	query["_id"] = idFilter

	staffIds := make([]bson.ObjectId, 0, len(mc.StaffUsersIds) + 1)
	staffIds = append(staffIds, mc.ManagerUserID)
	staffIds = append(staffIds, mc.StaffUsersIds...)
	idFilter["$in"] = staffIds

	if lastID != nil {
		idFilter["$gt"] = bson.ObjectIdHex(*lastID)
	}
	if len(name) != 0 {
		query["name"] = bson.RegEx{
			Pattern: js_regex.EscapeTextPattern(name),
		}
	}
	if err := mgr.userCollection.Find(query).Sort("_id").Limit(limit).All(&users); err != nil {
		panic(err)
	}
	return users
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
