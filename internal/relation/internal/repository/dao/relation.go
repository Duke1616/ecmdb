package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type RelationDAO interface {
	CreateModelRelation(ctx context.Context, mg ModelRelation) (int64, error)
}

func NewRelationDAO(client *mongo.Client) RelationDAO {
	return &relationDAO{
		db: client.Database("cmdb"),
	}
}

type relationDAO struct {
	db *mongo.Database
}

func (dao *relationDAO) CreateModelRelation(ctx context.Context, mr ModelRelation) (int64, error) {
	now := time.Now()
	mr.Ctime, mr.Utime = now.UnixMilli(), now.UnixMilli()
	mr.Id = mongox.GetDataID(dao.db, "c_relation_model")
	mr.RelationName = fmt.Sprintf("%s_%s_%s",
		mr.SourceModelIdentifies, mr.RelationTypeIdentifies, mr.TargetModelIdentifies)

	col := dao.db.Collection("c_relation_model")
	_, err := col.InsertMany(ctx, []interface{}{mr})

	if err != nil {
		return 0, err
	}

	return mr.Id, nil
}

type ModelRelation struct {
	Id                     int64  `bson:"id"`
	SourceModelIdentifies  string `bson:"source_model_identifies"`
	TargetModelIdentifies  string `bson:"target_model_identifies"`
	RelationTypeIdentifies string `bson:"relation_type_identifies"`
	RelationName           string `bson:"relation_name"`
	Mapping                string `bson:"mapping"`
	Ctime                  int64  `bson:"ctime"`
	Utime                  int64  `bson:"utime"`
}
