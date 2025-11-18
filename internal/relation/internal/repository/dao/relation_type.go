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

type RelationTypeDAO interface {
	// Create 创建关联类型
	Create(ctx context.Context, r RelationType) (int64, error)

	// BatchCreate 批量创建
	BatchCreate(ctx context.Context, rts []RelationType) error

	// List 关联列表
	List(ctx context.Context, offset, limit int64) ([]RelationType, error)

	// Count 数量
	Count(ctx context.Context) (int64, error)
}

func NewRelationTypeDAO(db *mongox.Mongo) RelationTypeDAO {
	return &relationDAO{
		db: db,
	}
}

type relationDAO struct {
	db *mongox.Mongo
}

func (dao *relationDAO) BatchCreate(ctx context.Context, rts []RelationType) error {
	if len(rts) == 0 {
		return nil
	}

	col := dao.db.Collection(RelationTypeCollection)
	now := time.Now().UnixMilli()

	// 批量获取起始 ID（一次数据库调用）
	startID, err := dao.db.GetBatchIdGenerator(RelationTypeCollection, len(rts))
	if err != nil {
		return fmt.Errorf("获取批量 ID 错误: %w", err)
	}

	// 为每个属性设置 ID 和时间戳
	docs := make([]interface{}, len(rts))
	for i := range rts {
		rts[i].Id = startID + int64(i)
		rts[i].Ctime, rts[i].Utime = now, now
		docs[i] = rts[i]
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

func (dao *relationDAO) Create(ctx context.Context, r RelationType) (int64, error) {
	now := time.Now()
	r.Ctime, r.Utime = now.UnixMilli(), now.UnixMilli()

	// 直接插入数据，并自增ID
	_, err := dao.db.InsertOneWithAutoID(ctx, RelationTypeCollection, &r)

	if err != nil {
		if mongox.IsUniqueConstraintError(err) {
			return 0, fmt.Errorf("关联类型插入: %w", errs.ErrUniqueDuplicate)
		}
		return 0, err
	}

	return r.Id, nil
}

func (dao *relationDAO) List(ctx context.Context, offset, limit int64) ([]RelationType, error) {
	col := dao.db.Collection(RelationTypeCollection)
	filter := bson.M{}
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

	var result []RelationType
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *relationDAO) Count(ctx context.Context) (int64, error) {
	col := dao.db.Collection(RelationTypeCollection)
	filer := bson.M{}

	count, err := col.CountDocuments(ctx, filer)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

type RelationType struct {
	Id             int64  `bson:"id"`
	Name           string `bson:"name"`
	UID            string `bson:"uid"`
	SourceDescribe string `bson:"source_describe"`
	TargetDescribe string `bson:"target_describe"`
	Ctime          int64  `bson:"ctime"`
	Utime          int64  `bson:"utime"`
}

func (a *RelationType) SetID(id int64) {
	a.Id = id
}

func (a *RelationType) GetID() int64 {
	return a.Id
}
