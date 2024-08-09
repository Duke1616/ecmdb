package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const RoleCollection = "c_role"

type RoleDAO interface {
	CreateRole(ctx context.Context, r Role) (int64, error)
	ListRole(ctx context.Context, offset, limit int64) ([]Role, error)
	UpdateRole(ctx context.Context, r Role) (int64, error)
	Count(ctx context.Context) (int64, error)
}

type roleDAO struct {
	db *mongox.Mongo
}

func (dao *roleDAO) UpdateRole(ctx context.Context, r Role) (int64, error) {
	col := dao.db.Collection(RoleCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"name":   r.Name,
			"desc":   r.Desc,
			"status": r.Status,
			"code":   r.Code,
			"utime":  time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": r.Id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *roleDAO) ListRole(ctx context.Context, offset, limit int64) ([]Role, error) {
	col := dao.db.Collection(RoleCollection)
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

	var result []Role
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *roleDAO) Count(ctx context.Context) (int64, error) {
	col := dao.db.Collection(RoleCollection)
	filer := bson.M{}

	count, err := col.CountDocuments(ctx, filer)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *roleDAO) CreateRole(ctx context.Context, r Role) (int64, error) {
	r.Id = dao.db.GetIdGenerator(RoleCollection)
	col := dao.db.Collection(RoleCollection)
	now := time.Now()
	r.Ctime, r.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, r)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return r.Id, nil
}

func NewRoleDAO(db *mongox.Mongo) RoleDAO {
	return &roleDAO{
		db: db,
	}
}

type Role struct {
	Id     int64  `bson:"id"`
	Name   string `bson:"name"`
	Code   string `bson:"code"`
	Desc   string `bson:"desc"`
	Status bool   `bson:"status"`
	Ctime  int64  `bson:"ctime"`
	Utime  int64  `bson:"utime"`
}
