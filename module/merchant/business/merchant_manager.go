package business

import (
	"wawa_b.v1/module/js_regex"
	"wawa_b.v1/module/merchant/domain"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// 商家管理器
type MerchantManager interface {
	// 删除一个商家
	Delete(id string)

	// 获取一个商家
	Get(id string) *domain.Merchant

	// 注册一个新商家
	Register(name, itemsOfBusiness, contactsName, contactsMobile, contactsIDCard, contactsAddress string)

	// 商家检索
	Retrieve(lastID *string, limit int, name string) []*domain.Merchant

	// 更新一个商家的基本信息
	Update(id, name, itemsOfBusiness, contactsName, contactsMobile, contactsIDCard, contactsAddress string)
}

// MongoDB商家管理器
type MongoDBMerchantManager struct {
	// 商家集合
	merchantCollection *mgo.Collection
}

// 创建一个MongoDB用户管理器
func NewMongoDBMerchantManager(mxsDB *mgo.Database) *MongoDBMerchantManager {
	return &MongoDBMerchantManager{
		merchantCollection: mxsDB.C("Merchants"),
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

func (mgr *MongoDBMerchantManager) Register(name, itemsOfBusiness, contactsName, contactsMobile, contactsIDCard, contactsAddress string) {
	if err := mgr.merchantCollection.Insert(&domain.Merchant{
		Name: name,
		ItemsOfBusiness: itemsOfBusiness,
		ContactsName: contactsName,
		ContactsMobile: contactsMobile,
		ContactsIDCard: contactsIDCard,
		ContactsAddress: contactsAddress,
	}); err != nil {
		panic(err)
	}
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
