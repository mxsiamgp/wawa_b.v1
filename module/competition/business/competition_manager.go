package business

import (
	"wawa_b.v1/module/competition/domain"
	"wawa_b.v1/module/js_regex"
	"wawa_b.v1/module/rest_json_rpc/failure"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// 失败代码
const (
	// 赛事名重复
	FAIL_CD_DUPLICATE_COMPETITION_NAME = "COMPETITION.DUPLICATE_COMPETITION_NAME"
)

// 赛事管理器
type CompetitionManager interface {
	// 添加一个新赛事
	Add(name string, tickets []*domain.Ticket) error

	// 删除一个赛事
	Delete(id string)

	// 完成一个赛事
	Finish(id string)

	// 获取一个赛事
	Get(id string) *domain.Competition

	// 根据赛事名获取一个赛事
	GetByName(name string) *domain.Competition

	// 检索赛事
	Retrieve(lastID *string, limit int, name string) []*domain.Competition

	// 更新一个赛事的门票集
	UpdateTickets(id string, tickets []*domain.Ticket)
}

// MongoDB赛事管理器
type MongoDBCompetitionManager struct {
	// 赛事集合
	competitionCollection *mgo.Collection
}

// 创建一个MongoDB赛事管理器
func NewMongoDBCompetitionManager(mxsDB *mgo.Database) *MongoDBCompetitionManager {
	return &MongoDBCompetitionManager{
		competitionCollection: mxsDB.C("Competitions"),
	}
}

func (mgr *MongoDBCompetitionManager) Add(name string, tickets []*domain.Ticket) error {
	if mgr.GetByName(name) != nil {
		return failure.New(FAIL_CD_DUPLICATE_COMPETITION_NAME)
	}

	for _, ticket := range tickets {
		ticket.ID = bson.NewObjectId()
	}

	if err := mgr.competitionCollection.Insert(&domain.Competition{
		Name: name,
		IsFinished: false,
		Tickets: tickets,
	}); err != nil {
		panic(err)
	}

	return nil
}

func (mgr *MongoDBCompetitionManager) Delete(id string) {
	if err := mgr.competitionCollection.RemoveId(bson.ObjectIdHex(id)); err != nil {
		panic(err)
	}
}

func (mgr *MongoDBCompetitionManager) Finish(id string) {
	if err := mgr.competitionCollection.UpdateId(bson.ObjectIdHex(id), bson.M{
		"$set": bson.M{
			"isFinished": true,
		},
	}); err != nil {
		panic(err)
	}
}

func (mgr *MongoDBCompetitionManager) Get(id string) *domain.Competition {
	competitions := make([]*domain.Competition, 0)
	if err := mgr.competitionCollection.FindId(bson.ObjectIdHex(id)).All(&competitions); err != nil {
		panic(err)
	}
	if len(competitions) == 0 {
		return nil
	}
	return competitions[0]
}

func (mgr *MongoDBCompetitionManager) GetByName(name string) *domain.Competition {
	competitions := make([]*domain.Competition, 0)
	if err := mgr.competitionCollection.Find(bson.M{
		"name": name,
	}).All(&competitions); err != nil {
		panic(err)
	}
	if len(competitions) == 0 {
		return nil
	}
	return competitions[0]
}

func (mgr *MongoDBCompetitionManager) Retrieve(lastID *string, limit int, name string) []*domain.Competition {
	competitions := make([]*domain.Competition, 0)
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
	if err := mgr.competitionCollection.Find(query).Sort("_id").Limit(limit).All(&competitions); err != nil {
		panic(err)
	}
	return competitions
}

func (mgr *MongoDBCompetitionManager) UpdateTickets(id string, tickets []*domain.Ticket) {
	if err := mgr.competitionCollection.UpdateId(bson.ObjectIdHex(id), bson.M{
		"$set": bson.M{
			"tickets": tickets,
		},
	}); err != nil {
		panic(err)
	}
}
