package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

const (
	ModelRelationCollection    = "c_relation_model"
	ResourceRelationCollection = "c_relation_resource"
)

type RelationDAO interface {
	CreateModelRelation(ctx context.Context, mg ModelRelation) (int64, error)
	CreateResourceRelation(ctx context.Context, mg ResourceRelation) (int64, error)
}

func NewRelationDAO(client *mongo.Client) RelationDAO {
	return &relationDAO{
		db: mongox.NewMongo(client),
	}
}

type relationDAO struct {
	db *mongox.Mongo
}

func (dao *relationDAO) CreateModelRelation(ctx context.Context, mr ModelRelation) (int64, error) {
	now := time.Now()
	mr.Ctime, mr.Utime = now.UnixMilli(), now.UnixMilli()
	mr.Id = dao.db.GetIdGenerator(ModelRelationCollection)
	col := dao.db.Collection(ModelRelationCollection)

	mr.RelationName = fmt.Sprintf("%s_%s_%s",
		mr.SourceModelIdentifies, mr.RelationTypeIdentifies, mr.TargetModelIdentifies)

	_, err := col.InsertMany(ctx, []interface{}{mr})

	if err != nil {
		return 0, err
	}

	return mr.Id, nil
}

func (dao *relationDAO) CreateResourceRelation(ctx context.Context, mr ResourceRelation) (int64, error) {
	now := time.Now()
	mr.Ctime, mr.Utime = now.UnixMilli(), now.UnixMilli()
	mr.Id = dao.db.GetIdGenerator(ResourceRelationCollection)
	col := dao.db.Collection(ResourceRelationCollection)

	mr.RelationName = fmt.Sprintf("%s_%s_%s",
		mr.SourceModelIdentifies, mr.RelationTypeIdentifies, mr.TargetModelIdentifies)

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

type ResourceRelation struct {
	Id                     int64  `bson:"id"`
	SourceModelIdentifies  string `bson:"source_model_identifies"`
	TargetModelIdentifies  string `bson:"target_model_identifies"`
	SourceResourceID       int64  `bson:"source_resource_id"`
	TargetResourceID       int64  `bson:"target_resource_id"`
	RelationTypeIdentifies string `bson:"relation_type_identifies"`
	RelationName           string `bson:"relation_name"`
	Ctime                  int64  `bson:"ctime"`
	Utime                  int64  `bson:"utime"`
}
