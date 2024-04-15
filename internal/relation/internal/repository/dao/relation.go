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

const (
	ModelRelationCollection    = "c_relation_model"
	ResourceRelationCollection = "c_relation_resource"
)

type RelationDAO interface {
	CreateModelRelation(ctx context.Context, mg ModelRelation) (int64, error)
	CreateResourceRelation(ctx context.Context, mg ResourceRelation) (int64, error)
	ListModelRelation(ctx context.Context, offset, limit int64) ([]*ModelRelation, error)
	ListResourceRelation(ctx context.Context, offset, limit int64) ([]*ResourceRelation, error)
	Count(ctx context.Context) (int64, error)

	ListRelationByModelUid(ctx context.Context, offset, limit int64, modelUid string) ([]*ModelRelation, error)
	CountByModelUid(ctx context.Context, modelUid string) (int64, error)
	ListResourceIds(ctx context.Context, modelUid string, relationType string) ([]int64, error)
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
		mr.SourceModelUID, mr.RelationTypeUID, mr.TargetModelUID)

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
		mr.SourceModelUID, mr.RelationTypeUID, mr.TargetModelUID)

	_, err := col.InsertMany(ctx, []interface{}{mr})

	if err != nil {
		return 0, err
	}

	return mr.Id, nil
}

func (dao *relationDAO) ListModelRelation(ctx context.Context, offset, limit int64) ([]*ModelRelation, error) {
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

func (dao *relationDAO) ListResourceRelation(ctx context.Context, offset, limit int64) ([]*ResourceRelation, error) {
	col := dao.db.Collection(ResourceRelationCollection)

	filer := bson.M{}
	opt := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	resp, err := col.Find(ctx, filer, opt)
	var set []*ResourceRelation
	for resp.Next(ctx) {
		ins := &ResourceRelation{}
		if err = resp.Decode(ins); err != nil {
			return nil, err
		}
		set = append(set, ins)
	}

	return set, nil
}

func (dao *relationDAO) Count(ctx context.Context) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (dao *relationDAO) ListRelationByModelUid(ctx context.Context, offset, limit int64, modelUid string) ([]*ModelRelation, error) {
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

func (dao *relationDAO) CountByModelUid(ctx context.Context, modelUid string) (int64, error) {
	col := dao.db.Collection(ModelRelationCollection)
	filer := bson.M{"relation_name": bson.M{"$regex": primitive.Regex{Pattern: modelUid, Options: "i"}}}

	count, err := col.CountDocuments(ctx, filer)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (dao *relationDAO) ListResourceIds(ctx context.Context, modelUid string, relationType string) ([]int64, error) {
	col := dao.db.Collection(ResourceRelationCollection)
	filer := bson.M{
		"$and": []bson.M{
			{"relation_name": bson.M{"$regex": primitive.Regex{Pattern: modelUid, Options: "i"}}},
			{"relation_type_uid": relationType},
		},
	}
	opt := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	resp, err := col.Find(ctx, filer, opt)

	var set []int64
	for resp.Next(ctx) {
		var result struct {
			Id int64 `bson:"id"`
		}

		if err = resp.Decode(&result); err != nil {
			return nil, err
		}
		set = append(set, result.Id)
	}

	return set, nil
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

type ResourceRelation struct {
	Id               int64  `bson:"id"`
	SourceModelUID   string `bson:"source_model_uid"`
	TargetModelUID   string `bson:"target_model_uid"`
	SourceResourceID int64  `bson:"source_resource_id"`
	TargetResourceID int64  `bson:"target_resource_id"`
	RelationTypeUID  string `bson:"relation_type_uid"`
	RelationName     string `bson:"relation_name"` // 唯一标识、以防重复创建
	Ctime            int64  `bson:"ctime"`
	Utime            int64  `bson:"utime"`
}
