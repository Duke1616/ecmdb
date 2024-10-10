package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type RelationModelDAO interface {
	CreateModelRelation(ctx context.Context, mr ModelRelation) (int64, error)

	// ListRelationByModelUid 查询模型关联关系
	ListRelationByModelUid(ctx context.Context, offset, limit int64, modelUid string) ([]ModelRelation, error)

	CountByModelUid(ctx context.Context, modelUid string) (int64, error)

	// FindModelDiagramBySrcUids 查询模型拓扑图
	FindModelDiagramBySrcUids(ctx context.Context, srcUids []string) ([]ModelRelation, error)

	DeleteModelRelation(ctx context.Context, id int64) (int64, error)
}

func NewRelationModelDAO(db *mongox.Mongo) RelationModelDAO {
	return &modelDAO{
		db: db,
	}
}

type modelDAO struct {
	db *mongox.Mongo
}

func (dao *modelDAO) CreateModelRelation(ctx context.Context, mr ModelRelation) (int64, error) {
	mr.Id = dao.db.GetIdGenerator(ModelRelationCollection)
	col := dao.db.Collection(ModelRelationCollection)
	now := time.Now()
	mr.Ctime, mr.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, mr)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return mr.Id, nil
}

func (dao *modelDAO) FindModelDiagramBySrcUids(ctx context.Context, srcUids []string) ([]ModelRelation, error) {
	col := dao.db.Collection(ModelRelationCollection)
	filter := bson.M{"source_model_uid": bson.M{"$in": srcUids}}
	opts := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	cursor, err := col.Find(ctx, filter, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []ModelRelation
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *modelDAO) ListRelationByModelUid(ctx context.Context, offset, limit int64, modelUid string) ([]ModelRelation, error) {
	col := dao.db.Collection(ModelRelationCollection)
	filter := bson.M{
		"$or": bson.A{
			bson.M{"source_model_uid": modelUid},
			bson.M{"target_model_uid": modelUid},
		},
	}
	// 这种情况会出现意外、比如 host-1 host-2 会查询错误
	//filter := bson.M{"relation_name": bson.M{"$regex": primitive.Regex{Pattern: modelUid, Options: "i"}}}
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	cursor, err := col.Find(ctx, filter, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []ModelRelation
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *modelDAO) CountByModelUid(ctx context.Context, modelUid string) (int64, error) {
	col := dao.db.Collection(ModelRelationCollection)
	filter := bson.M{
		"$or": bson.A{
			bson.M{"source_model_uid": modelUid},
			bson.M{"target_model_uid": modelUid},
		},
	}
	//filter := bson.M{"relation_name": bson.M{"$regex": primitive.Regex{Pattern: modelUid, Options: "i"}}}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *modelDAO) DeleteModelRelation(ctx context.Context, id int64) (int64, error) {
	col := dao.db.Collection(ModelRelationCollection)
	filter := bson.M{"id": id}

	result, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

type ModelRelation struct {
	Id              int64  `bson:"id"`
	SourceModelUid  string `bson:"source_model_uid"`
	TargetModelUid  string `bson:"target_model_uid"`
	RelationTypeUid string `bson:"relation_type_uid"`
	RelationName    string `bson:"relation_name"` // 唯一标识、以防重复创建
	Mapping         string `bson:"mapping"`
	Ctime           int64  `bson:"ctime"`
	Utime           int64  `bson:"utime"`
}
