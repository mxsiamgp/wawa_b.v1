package business

import (
	"wawa_b.v1/module/md5"
	"wawa_b.v1/module/js_regex"
	"wawa_b.v1/module/rest_json_rpc/failure"
	"wawa_b.v1/module/user/domain"
	"wawa_b.v1/module/user/util/permission"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// 失败代码
const (
	// 用户名重复
	FAIL_CD_DUPLICATE_USER_NAME = "USER.DUPLICATE_USER_NAME"

	// 授权人没有相应的权限
	FAIL_CD_GRANTER_NO_THESE_PERMISSIONS = "USER.GRANTER_NO_THESE_PERMISSIONS"

	// 用户类型无效
	FAIL_CD_INVALID_USER_KIND = "USER.INVALID_USER_KIND"
)

// 用户管理器
type UserManager interface {
	// 绑定微信OpenID
	BindWechatOpenID(userID, wechatOpenID string)

	// 删除一个用户
	Delete(id string)

	// 获取一个用户
	Get(id string) *domain.User

	// 根据名称获取一个用户
	GetByName(name string) *domain.User

	// 根据微信OpenID获取一个用户
	GetByWechatOpenID(wechatOpenID string) *domain.User

	// 授予用户权限
	GrantFlatPermissions(granterID, granteeID string, perms []string) error

	// 注册一个新用户
	Register(kind, name, password, nickname, mobile string) error

	// 用户检索
	Retrieve(lastID *string, limit int, name string, nickname string) []*domain.User

	// 更新一个用户的基本信息
	Update(id, nickname string)

	// 修改一个用户密码
	UpdatePassword(name, password string)
}

// MongoDB用户管理器
type MongoDBUserManager struct {
	// 用户集合
	userCollection        *mgo.Collection

	// 微信用户集合
	wechatUserBindingCollection  *mgo.Collection

	// 用户类型到权限集合的映射
	flatPermissionsByKind map[string][]string
}

// 创建一个MongoDB用户管理器
func NewMongoDBUserManager(masterDB *mgo.Database, flatPermissionsByKind map[string][]string) *MongoDBUserManager {
	return &MongoDBUserManager{
		userCollection: masterDB.C("Users"),
		wechatUserBindingCollection: masterDB.C("WechatUserBindings"),
		flatPermissionsByKind: flatPermissionsByKind,
	}
}

func (mgr *MongoDBUserManager) BindWechatOpenID(userID, openID string) {
	if _, err := mgr.wechatUserBindingCollection.Upsert(bson.M{
		"openId": openID,
	}, bson.M{
		"openId": openID,
		"userId": bson.ObjectIdHex(userID),
	}); err != nil {
		panic(err)
	}
}

func (mgr *MongoDBUserManager) Delete(id string) {
	if err := mgr.userCollection.RemoveId(bson.ObjectIdHex(id)); err != nil {
		panic(err)
	}
}

func (mgr *MongoDBUserManager) Get(id string) *domain.User {
	users := make([]*domain.User, 0)
	if err := mgr.userCollection.FindId(bson.ObjectIdHex(id)).All(&users); err != nil {
		panic(err)
	}
	if len(users) == 0 {
		return nil
	}
	return users[0]
}

func (mgr *MongoDBUserManager) GetByName(name string) *domain.User {
	users := make([]*domain.User, 0)
	if err := mgr.userCollection.Find(bson.M{
		"name": name,
	}).All(&users); err != nil {
		panic(err)
	}
	if len(users) == 0 {
		return nil
	}
	return users[0]
}

func (mgr *MongoDBUserManager) GetByWechatOpenID(openID string) *domain.User {
	bindings := make([]*domain.WechatUserBinding, 0)
	if err := mgr.wechatUserBindingCollection.Find(bson.M{
		"openId": openID,
	}).All(&bindings); err != nil {
		panic(err)
	}
	if len(bindings) == 0 {
		return nil
	}
	return mgr.Get(bindings[0].UserID.Hex())
}

func (mgr *MongoDBUserManager) GrantFlatPermissions(granterID, granteeID string, perms []string) error {
	greater := mgr.Get(granterID)
	if greater == nil {
		return nil
	}

	if !permission.IsInclude(greater.FlatPermissions, perms) {
		return failure.New(FAIL_CD_GRANTER_NO_THESE_PERMISSIONS)
	}

	if err := mgr.userCollection.UpdateId(bson.ObjectIdHex(granteeID), bson.M{
		"$set": bson.M{
			"flatPermissions": perms,
		},
	}); err != nil {
		panic(err)
	}

	return nil
}

func (mgr *MongoDBUserManager) Register(kind, name, password, nickname, mobile string) error {
	perms, ok := mgr.flatPermissionsByKind[kind]
	if !ok {
		return failure.New(FAIL_CD_INVALID_USER_KIND)
	}
	if mgr.GetByName(name) != nil {
		return failure.New(FAIL_CD_DUPLICATE_USER_NAME)
	}
	if err := mgr.userCollection.Insert(&domain.User{
		Name: name,
		PasswordDigest: md5.StringDigest([]byte(password)),
		Nickname: nickname,
		Mobile: mobile,
		FlatPermissions: perms,
	}); err != nil {
		panic(err)
	}
	return nil
}

func (mgr *MongoDBUserManager) Retrieve(lastID *string, limit int, name, nickname string) []*domain.User {
	users := make([]*domain.User, 0)
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
	if len(nickname) != 0 {
		query["nickname"] = bson.RegEx{
			Pattern: js_regex.EscapeTextPattern(nickname),
		}
	}
	if err := mgr.userCollection.Find(query).Sort("_id").Limit(limit).All(&users); err != nil {
		panic(err)
	}
	return users
}

func (mgr *MongoDBUserManager) Update(id, nickname string) {
	if err := mgr.userCollection.UpdateId(bson.ObjectIdHex(id), bson.M{
		"$set": bson.M{
			"nickname": nickname,
		},
	}); err != nil {
		panic(err)
	}
}

func (mgr *MongoDBUserManager) UpdatePassword(id, password string) {
	if err := mgr.userCollection.UpdateId(bson.ObjectIdHex(id), bson.M{
		"$set": bson.M{
			"passwordDigest": md5.StringDigest([]byte(password)),
		},
	}); err != nil {
		panic(err)
	}
}
