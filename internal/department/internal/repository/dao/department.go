package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	DepartmentCollection = "c_department"
)

type DepartmentDAO interface {
	CreateDepartment(ctx context.Context, req Department) (int64, error)
	UpdateDepartment(ctx context.Context, req Department) (int64, error)
	DeleteDepartment(ctx context.Context, id int64) (int64, error)
	ListDepartment(ctx context.Context) ([]Department, error)
	FindByid(ctx context.Context, id int64) (Department, error)
	ListDepartmentByIds(ctx context.Context, ids []int64) ([]Department, error)
}

func NewDepartmentDAO(db *mongox.Mongo) DepartmentDAO {
	return &departmentDAO{
		db: db,
	}
}

type departmentDAO struct {
	db *mongox.Mongo
}

func (dao *departmentDAO) FindByid(ctx context.Context, id int64) (Department, error) {
	col := dao.db.Collection(DepartmentCollection)
	var department Department
	filter := bson.M{"id": id}

	if err := col.FindOne(ctx, filter).Decode(&department); err != nil {
		return Department{}, fmt.Errorf("解码错误，%w", err)
	}

	return department, nil
}

func (dao *departmentDAO) ListDepartmentByIds(ctx context.Context, ids []int64) ([]Department, error) {
	col := dao.db.Collection(DepartmentCollection)
	filter := bson.M{"id": bson.M{"$in": ids}}

	cursor, err := col.Find(ctx, filter)
	var result []Department
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *departmentDAO) DeleteDepartment(ctx context.Context, id int64) (int64, error) {
	col := dao.db.Collection(DepartmentCollection)
	filter := bson.M{"id": id}

	result, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *departmentDAO) UpdateDepartment(ctx context.Context, req Department) (int64, error) {
	col := dao.db.Collection(DepartmentCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"name":        req.Name,
			"pid":         req.Pid,
			"sort":        req.Sort,
			"enabled":     req.Enabled,
			"leaders":     req.Leaders,
			"main_leader": req.MainLeader,
			"utime":       time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": req.Id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *departmentDAO) ListDepartment(ctx context.Context) ([]Department, error) {
	col := dao.db.Collection(DepartmentCollection)
	filter := bson.M{}

	opts := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}
	cursor, err := col.Find(ctx, filter, opts)
	var result []Department
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *departmentDAO) CreateDepartment(ctx context.Context, req Department) (int64, error) {
	req.Id = dao.db.GetIdGenerator(DepartmentCollection)
	col := dao.db.Collection(DepartmentCollection)
	now := time.Now()
	req.Ctime, req.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return req.Id, nil
}

type Department struct {
	Id         int64   `bson:"id"`
	Pid        int64   `bson:"pid"`
	Name       string  `bson:"name"`
	Sort       int64   `bson:"sort"`
	Enabled    bool    `bson:"enabled"`
	Leaders    []int64 `bson:"leaders"`
	MainLeader int64   `bson:"main_leader"`
	Ctime      int64   `bson:"ctime"`
	Utime      int64   `bson:"utime"`
}
