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
	CodebookCollection = "c_codebook"
)

type CodebookDAO interface {
	CreateCodebook(ctx context.Context, c Codebook) (int64, error)
	DetailCodebook(ctx context.Context, id int64) (Codebook, error)
	ListCodebook(ctx context.Context, offset, limit int64) ([]Codebook, error)
	Count(ctx context.Context) (int64, error)
	UpdateCodebook(ctx context.Context, c Codebook) (int64, error)
	DeleteCodebook(ctx context.Context, id int64) (int64, error)
}

func NewCodebookDAO(db *mongox.Mongo) CodebookDAO {
	return &codebookDAO{
		db: db,
	}
}

type codebookDAO struct {
	db *mongox.Mongo
}

func (dao *codebookDAO) CreateCodebook(ctx context.Context, c Codebook) (int64, error) {
	c.Id = dao.db.GetIdGenerator(CodebookCollection)
	col := dao.db.Collection(CodebookCollection)
	now := time.Now()
	c.Ctime, c.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, c)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return c.Id, nil
}

func (dao *codebookDAO) DetailCodebook(ctx context.Context, id int64) (Codebook, error) {
	col := dao.db.Collection(CodebookCollection)
	filter := bson.M{"id": id}

	var t Codebook
	if err := col.FindOne(ctx, filter).Decode(&t); err != nil {
		return Codebook{}, fmt.Errorf("解码错误，%w", err)
	}

	return t, nil
}

func (dao *codebookDAO) ListCodebook(ctx context.Context, offset, limit int64) ([]Codebook, error) {
	col := dao.db.Collection(CodebookCollection)
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

	var result []Codebook
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *codebookDAO) Count(ctx context.Context) (int64, error) {
	col := dao.db.Collection(CodebookCollection)
	filer := bson.M{}

	count, err := col.CountDocuments(ctx, filer)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *codebookDAO) UpdateCodebook(ctx context.Context, c Codebook) (int64, error) {
	col := dao.db.Collection(CodebookCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"name":  c.Name,
			"code":  c.Code,
			"utime": time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": c.Id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *codebookDAO) DeleteCodebook(ctx context.Context, id int64) (int64, error) {
	col := dao.db.Collection(CodebookCollection)
	filter := bson.M{"id": id}

	result, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

type Codebook struct {
	Id       int64  `bson:"id"`
	Name     string `bson:"name"`
	Code     string `bson:"code"`
	Language string `bson:"language"`
	Ctime    int64  `bson:"ctime"`
	Utime    int64  `bson:"utime"`
}
