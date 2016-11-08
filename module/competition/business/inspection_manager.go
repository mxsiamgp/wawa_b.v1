package business

import (
	"wawa_b.v1/module/rest_json_rpc/failure"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// 失败代码
const (
	// 已验票
	FAIL_CD_INSPECTION_INSPECTED = "COMPETITION.INSPECTION.INSPECTED"
)

// 验票管理器
type InspectionManager interface {
	// 标记已验票
	MarkInspected(competitionID, userId string) error
}

// MongoDB验票管理器
type MongoDBInspectionManager struct {
	// 验票集合
	inspectionCollection *mgo.Collection
}

// 创建一个MongoDB验票管理器
func NewMongoDBInspectionManager(mxsDB *mgo.Database) *MongoDBInspectionManager {
	return &MongoDBInspectionManager{
		inspectionCollection: mxsDB.C("Inspections"),
	}
}

func (mgr *MongoDBInspectionManager) MarkInspected(competitionID, userID string) error {
	n, err := mgr.inspectionCollection.Find(bson.M{
		"competitionId": bson.ObjectIdHex(competitionID),
		"userId": bson.ObjectIdHex(userID),
	}).Count()
	if err != nil {
		panic(err)
	}

	// 已验票
	if n != 0 {
		return failure.New(FAIL_CD_INSPECTION_INSPECTED)
	}

	if err := mgr.inspectionCollection.Insert(bson.M{
		"competitionId": bson.ObjectIdHex(competitionID),
		"userId": bson.ObjectIdHex(userID),
	}); err != nil {
		panic(err)
	}

	return nil
}
