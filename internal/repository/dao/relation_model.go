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

const (
	ModelRelationCollection = "c_relation_model"
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

func NewRelationModelDAO(db *mongox.DB) RelationModelDAO {
	return &modelRelationDAO{
		db:   db,
		coll: mongox.NewCollection[ModelRelation](db, ModelRelationCollection),
	}
}

type modelRelationDAO struct {
	db   *mongox.DB
	coll *mongox.Collection[ModelRelation]
}

func (dao *modelRelationDAO) BatchCreate(ctx context.Context, relations []ModelRelation) error {
	if len(relations) == 0 {
		return nil
	}

	now := time.Now().UnixMilli()

	// 批量获取起始 ID
	startID, err := dao.db.GetBatchIdGenerator(ModelRelationCollection, len(relations))
	if err != nil {
		return fmt.Errorf("获取批量 ID 错误: %w", err)
	}

	docs := make([]*ModelRelation, len(relations))
	for i := range relations {
		relations[i].Id = startID + int64(i)
		relations[i].Ctime, relations[i].Utime = now, now
		docs[i] = &relations[i]
	}

	_, err = dao.coll.InsertMany(ctx, docs)
	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return fmt.Errorf("批量插入关联关系: %w", errs.ErrUniqueDuplicate)
		}
		return fmt.Errorf("批量插入数据错误: %w", err)
	}

	return nil
}

func (dao *modelRelationDAO) GetByRelationNames(ctx context.Context, names []string) ([]ModelRelation, error) {
	filter := bson.M{
		"relation_name": bson.M{"$in": names},
	}
	opts := &options.FindOptions{}

	return dao.coll.Find(ctx, filter, opts)
}

func (dao *modelRelationDAO) CreateModelRelation(ctx context.Context, mr ModelRelation) (int64, error) {
	now := time.Now()
	mr.Ctime, mr.Utime = now.UnixMilli(), now.UnixMilli()

	// 直接插入数据，借助 AutoIDPlugin 插件自动注入自增 ID。
	_, err := dao.coll.InsertOne(ctx, &mr)
	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return 0, fmt.Errorf("模型关联关系插入: %w", errs.ErrUniqueDuplicate)
		}
		return 0, err
	}

	return mr.Id, nil
}

func (dao *modelRelationDAO) FindModelDiagramBySrcUids(ctx context.Context, srcUids []string) ([]ModelRelation, error) {
	filter := bson.M{
		"source_model_uid": bson.M{"$in": srcUids},
	}
	opts := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	return dao.coll.Find(ctx, filter, opts)
}

func (dao *modelRelationDAO) ListRelationByModelUid(ctx context.Context, offset, limit int64, modelUid string) ([]ModelRelation, error) {
	filter := bson.M{
		"$or": bson.A{
			bson.M{"source_model_uid": modelUid},
			bson.M{"target_model_uid": modelUid},
		},
	}
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	return dao.coll.Find(ctx, filter, opts)
}

func (dao *modelRelationDAO) CountByModelUid(ctx context.Context, modelUid string) (int64, error) {
	filter := bson.M{
		"$or": bson.A{
			bson.M{"source_model_uid": modelUid},
			bson.M{"target_model_uid": modelUid},
		},
	}

	count, err := dao.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *modelRelationDAO) DeleteModelRelation(ctx context.Context, id int64) (int64, error) {
	filter := bson.M{"id": id}

	result, err := dao.coll.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *modelRelationDAO) CountByRelationTypeUid(ctx context.Context, uid string) (int64, error) {
	filter := bson.M{"relation_type_uid": uid}
	count, err := dao.coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("关联引用统计错误: %w", err)
	}
	return count, nil
}

func (dao *modelRelationDAO) GetByID(ctx context.Context, id int64) (ModelRelation, error) {
	filter := bson.M{"id": id}
	res, err := dao.coll.FindOne(ctx, filter)
	if err != nil {
		return ModelRelation{}, fmt.Errorf("查询错误: %w", err)
	}
	return *res, nil
}

func (dao *modelRelationDAO) UpdateModelRelation(ctx context.Context, mr ModelRelation) (int64, error) {
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
	res, err := dao.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("更新文档错误: %w", err)
	}
	return res.ModifiedCount, nil
}

type ModelRelation struct {
	TenantID        int64  `bson:"tenant_id"`
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
