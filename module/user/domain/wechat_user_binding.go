package domain

import "gopkg.in/mgo.v2/bson"

// 微信用户绑定
type WechatUserBinding struct {
	// ID
	ID     bson.ObjectId `bson:"_id,omitempty" json:"id"`

	// 微信OpenID
	OpenID string `bson:"openId" json:"open_id"`

	// 用户ID
	UserID bson.ObjectId `bson:"userId" json:"user_id"`
}
