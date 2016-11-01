package domain

import "gopkg.in/mgo.v2/bson"

// 用户
type User struct {
	// ID
	ID              bson.ObjectId `bson:"_id,omitempty" json:"id"`

	// 用户名
	Name            string `bson:"name" json:"name"`

	// 密码摘要
	PasswordDigest  string `bson:"passwordDigest" json:"password_digest"`

	// 昵称
	Nickname        string `bson:"nickname" json:"nickname"`

	// 手机
	Mobile          string `bson:"mobile" json:"mobile"`

	// 权限集
	FlatPermissions []string `bson:"flatPermissions" json:"flat_permissions"`
}
