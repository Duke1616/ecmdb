package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type RelationModelDAO interface {
	CreateModelRelation(ctx context.Context, mg ModelRelation) (int64, error)
	ListModelRelation(ctx context.Context, offset, limit int64) ([]*ModelRelation, error)
	ListRelationByModelUid(ctx context.Context, offset, limit int64, modelUid string) ([]*ModelRelation, error)
	CountByModelUid(ctx context.Context, modelUid string) (int64, error)
	Count(ctx context.Context) (int64, error)

	FindModelRelationBySourceUID(ctx context.Context, sourceUid string) ([]*ModelRelation, error)
}

func NewRelationModelDAO(client *mongo.Client) RelationModelDAO {
	return &modelDAO{
		db: mongox.NewMongo(client),
	}
}

type modelDAO struct {
	db *mongox.Mongo
}

func (dao *modelDAO) CreateModelRelation(ctx context.Context, mr ModelRelation) (int64, error) {
	now := time.Now()
	mr.Ctime, mr.Utime = now.UnixMilli(), now.UnixMilli()
	mr.Id = dao.db.GetIdGenerator(ModelRelationCollection)
	col := dao.db.Collection(ModelRelationCollection)

	mr.RelationName = fmt.Sprintf("%s_%s_%s",
		mr.SourceModelUID, mr.RelationTypeUID, mr.TargetModelUID)

	_, err := col.InsertMany(ctx, []interface{}{mr})

	if err != nil {
		return 0, err
	}

	return mr.Id, nil
}

func (dao *modelDAO) ListModelRelation(ctx context.Context, offset, limit int64) ([]*ModelRelation, error) {
	col := dao.db.Collection(ModelRelationCollection)

	filer := bson.M{}
	opt := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	resp, err := col.Find(ctx, filer, opt)
	var set []*ModelRelation
	for resp.Next(ctx) {
		ins := &ModelRelation{}
		if err = resp.Decode(ins); err != nil {
			return nil, err
		}
		set = append(set, ins)
	}

	return set, nil
}

func (dao *modelDAO) Count(ctx context.Context) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (dao *modelDAO) ListRelationByModelUid(ctx context.Context, offset, limit int64, modelUid string) ([]*ModelRelation, error) {
	col := dao.db.Collection(ModelRelationCollection)

	filer := bson.M{"relation_name": bson.M{"$regex": primitive.Regex{Pattern: modelUid, Options: "i"}}}
	opt := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	resp, err := col.Find(ctx, filer, opt)
	var set []*ModelRelation
	for resp.Next(ctx) {
		ins := &ModelRelation{}
		if err = resp.Decode(ins); err != nil {
			return nil, err
		}
		set = append(set, ins)
	}

	return set, nil
}

func (dao *modelDAO) FindModelRelationBySourceUID(ctx context.Context, sourceUid string) ([]*ModelRelation, error) {
	col := dao.db.Collection(ModelRelationCollection)

	filer := bson.M{"source_model_uid": sourceUid}
	opt := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	resp, err := col.Find(ctx, filer, opt)
	var set []*ModelRelation
	for resp.Next(ctx) {
		ins := &ModelRelation{}
		if err = resp.Decode(ins); err != nil {
			return nil, err
		}
		set = append(set, ins)
	}

	return set, nil
}

func (dao *modelDAO) CountByModelUid(ctx context.Context, modelUid string) (int64, error) {
	col := dao.db.Collection(ModelRelationCollection)
	filer := bson.M{"relation_name": bson.M{"$regex": primitive.Regex{Pattern: modelUid, Options: "i"}}}

	count, err := col.CountDocuments(ctx, filer)
	if err != nil {
		return 0, err
	}

	return count, nil
}

type ModelRelation struct {
	Id              int64  `bson:"id"`
	SourceModelUID  string `bson:"source_model_uid"`
	TargetModelUID  string `bson:"target_model_uid"`
	RelationTypeUID string `bson:"relation_type_uid"`
	RelationName    string `bson:"relation_name"` // 唯一标识、以防重复创建
	Mapping         string `bson:"mapping"`
	Ctime           int64  `bson:"ctime"`
	Utime           int64  `bson:"utime"`
}
