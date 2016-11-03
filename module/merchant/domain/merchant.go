package domain

import "gopkg.in/mgo.v2/bson"

// 商家
type Merchant struct {
	// ID
	ID              bson.ObjectId `bson:"_id,omitempty" json:"id"`

	// 商家店名
	Name            string `bson:"name" json:"name"`

	// 管理员用户ID
	ManagerUserID   bson.ObjectId `bson:"managerUserId" json:"manager_user_id"`

	// 管理员用户名（冗余）
	ManagerUserName string `bson:"managerUserName" json:"manager_user_name"`

	// 经营项目
	ItemsOfBusiness string `bson:"itemsOfBusiness" json:"items_of_business"`

	// 联系人姓名
	ContactsName    string `bson:"contactsName" json:"contacts_name"`

	// 联系人手机号码
	ContactsMobile  string `bson:"contactsMobile" json:"contacts_mobile"`

	// 联系人身份证号码
	ContactsIDCard  string `bson:"contactsIdCard" json:"contacts_id_card"`

	// 联系人地址
	ContactsAddress string `bson:"contactsAddress" json:"contacts_address"`
}
