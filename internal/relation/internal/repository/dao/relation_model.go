package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/errs"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RelationModelDAO interface {
	// CreateModelRelation 创建模型关联关系
	CreateModelRelation(ctx context.Context, mr ModelRelation) (int64, error)

	// DeleteModelRelation 删除模型关联关系
	DeleteModelRelation(ctx context.Context, id int64) (int64, error)

	// BatchCreate 批量创建模型关联关系
	BatchCreate(ctx context.Context, relations []ModelRelation) error

	// GetByRelationNames 根据唯一标识获取数据
	GetByRelationNames(ctx context.Context, names []string) ([]ModelRelation, error)

	// ListRelationByModelUid 根据模型 UID 获取。支持分页
	ListRelationByModelUid(ctx context.Context, offset, limit int64, modelUid string) ([]ModelRelation, error)

	// CountByModelUid 根据模型 UID 获取数量
	CountByModelUid(ctx context.Context, modelUid string) (int64, error)

	// FindModelDiagramBySrcUids 查询模型拓扑图
	FindModelDiagramBySrcUids(ctx context.Context, srcUids []string) ([]ModelRelation, error)

	// CountByRelationTypeUid 根据关联类型 UID 获取数量
	CountByRelationTypeUid(ctx context.Context, uid string) (int64, error)

	// GetByID 根据 ID 获取数据
	GetByID(ctx context.Context, id int64) (ModelRelation, error)

	// UpdateModelRelation 更新模型关联关系
	UpdateModelRelation(ctx context.Context, mr ModelRelation) (int64, error)
}

func NewRelationModelDAO(db *mongox.Mongo) RelationModelDAO {
	return &modelDAO{
		db: db,
	}
}

type modelDAO struct {
	db *mongox.Mongo
}

func (dao *modelDAO) BatchCreate(ctx context.Context, relations []ModelRelation) error {
	if len(relations) == 0 {
		return nil
	}

	col := dao.db.Collection(ModelRelationCollection)
	now := time.Now().UnixMilli()

	// 批量获取起始 ID（一次数据库调用）
	startID, err := dao.db.GetBatchIdGenerator(ModelRelationCollection, len(relations))
	if err != nil {
		return fmt.Errorf("获取批量 ID 错误: %w", err)
	}

	// 为每个属性设置 ID 和时间戳
	docs := make([]interface{}, len(relations))
	for i := range relations {
		relations[i].Id = startID + int64(i)
		relations[i].Ctime, relations[i].Utime = now, now
		docs[i] = relations[i]
	}

	_, err = col.InsertMany(ctx, docs)
	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return fmt.Errorf("批量插入关联类型: %w", errs.ErrUniqueDuplicate)
		}
		return fmt.Errorf("批量插入数据错误: %w", err)
	}

	return nil
}

func (dao *modelDAO) GetByRelationNames(ctx context.Context, names []string) ([]ModelRelation, error) {
	col := dao.db.Collection(ModelRelationCollection)
	filter := bson.M{}
	filter["relation_name"] = bson.M{"$in": names}
	opts := &options.FindOptions{}

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

func (dao *modelDAO) CreateModelRelation(ctx context.Context, mr ModelRelation) (int64, error) {
	now := time.Now()
	mr.Ctime, mr.Utime = now.UnixMilli(), now.UnixMilli()

	// 直接插入数据，并自增ID
	_, err := dao.db.InsertOneWithAutoID(ctx, ModelRelationCollection, &mr)

	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return 0, fmt.Errorf("模型关联关系插入: %w", errs.ErrUniqueDuplicate)
		}
		return 0, err
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

func (dao *modelDAO) CountByRelationTypeUid(ctx context.Context, uid string) (int64, error) {
	col := dao.db.Collection(ModelRelationCollection)
	filter := bson.M{"relation_type_uid": uid}
	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("关联引用统计错误: %w", err)
	}
	return count, nil
}

func (dao *modelDAO) GetByID(ctx context.Context, id int64) (ModelRelation, error) {
	col := dao.db.Collection(ModelRelationCollection)
	filter := bson.M{"id": id}
	var res ModelRelation
	if err := col.FindOne(ctx, filter).Decode(&res); err != nil {
		return ModelRelation{}, fmt.Errorf("查询错误: %w", err)
	}
	return res, nil
}

func (dao *modelDAO) UpdateModelRelation(ctx context.Context, mr ModelRelation) (int64, error) {
	col := dao.db.Collection(ModelRelationCollection)
	filter := bson.M{"id": mr.Id}
	update := bson.M{
		"$set": bson.M{
			"source_model_uid":  mr.SourceModelUid,
			"target_model_uid":  mr.TargetModelUid,
			"relation_type_uid": mr.RelationTypeUid,
			"relation_name":     mr.RelationName,
			"mapping":           mr.Mapping,
			"utime":             time.Now().UnixMilli(),
		},
	}
	res, err := col.UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("更新文档错误: %w", err)
	}
	return res.ModifiedCount, nil
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

func (a *ModelRelation) SetID(id int64) {
	a.Id = id
}

func (a *ModelRelation) GetID() int64 {
	return a.Id
}
