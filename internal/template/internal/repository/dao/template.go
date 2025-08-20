package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrDataNotFound 通用的数据没找到
var ErrDataNotFound = mongo.ErrNoDocuments

type TemplateDAO interface {
	CreateTemplate(ctx context.Context, t Template) (int64, error)
	FindByHash(ctx context.Context, hash string) (Template, error)
	FindByExternalTemplateId(ctx context.Context, externalTemplateId string) (Template, error)
	DetailTemplate(ctx context.Context, id int64) (Template, error)
	DetailTemplateByExternalTemplateId(ctx context.Context, externalId string) (Template, error)
	DeleteTemplate(ctx context.Context, id int64) (int64, error)
	UpdateTemplate(ctx context.Context, t Template) (int64, error)
	ListTemplate(ctx context.Context, offset, limit int64) ([]Template, error)
	Pipeline(ctx context.Context) ([]TemplatePipeline, error)
	Count(ctx context.Context) (int64, error)

	FindByTemplateIds(ctx context.Context, ids []int64) ([]Template, error)

	GetByWorkflowId(ctx context.Context, workflowId int64) ([]Template, error)
}

func NewTemplateDAO(db *mongox.Mongo) TemplateDAO {
	return &templateDAO{
		db: db,
	}
}

type templateDAO struct {
	db *mongox.Mongo
}

func (dao *templateDAO) FindByTemplateIds(ctx context.Context, ids []int64) ([]Template, error) {
	col := dao.db.Collection(TemplateCollection)
	filter := bson.M{"id": bson.M{"$in": ids}}
	cursor, err := col.Find(ctx, filter)
	var result []Template
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *templateDAO) GetByWorkflowId(ctx context.Context, workflowId int64) ([]Template, error) {
	col := dao.db.Collection(TemplateCollection)
	filter := bson.M{}
	filter["workflow_id"] = workflowId
	opts := &options.FindOptions{
		Sort: bson.D{{Key: "ctime", Value: -1}},
	}

	cursor, err := col.Find(ctx, filter, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []Template
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *templateDAO) DetailTemplateByExternalTemplateId(ctx context.Context, externalId string) (Template, error) {
	col := dao.db.Collection(TemplateCollection)
	filter := bson.M{"external_template_id": externalId}

	var t Template
	if err := col.FindOne(ctx, filter).Decode(&t); err != nil {
		return Template{}, fmt.Errorf("解码错误，%w", err)
	}

	return t, nil
}

func (dao *templateDAO) CreateTemplate(ctx context.Context, t Template) (int64, error) {
	t.Id = dao.db.GetIdGenerator(TemplateCollection)
	col := dao.db.Collection(TemplateCollection)
	now := time.Now()
	t.Ctime, t.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, t)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return t.Id, nil
}

func (dao *templateDAO) FindByHash(ctx context.Context, hash string) (Template, error) {
	col := dao.db.Collection(TemplateCollection)
	var t Template
	filter := bson.M{"unique_hash": hash}

	if err := col.FindOne(ctx, filter).Decode(&t); err != nil {
		return Template{}, fmt.Errorf("解码错误，%w", err)
	}

	return t, nil
}

func (dao *templateDAO) FindByExternalTemplateId(ctx context.Context, externalTemplateId string) (Template, error) {
	col := dao.db.Collection(TemplateCollection)
	var t Template
	filter := bson.M{"external_template_id": externalTemplateId}

	if err := col.FindOne(ctx, filter).Decode(&t); err != nil {
		return Template{}, fmt.Errorf("解码错误，%w", err)
	}

	return t, nil
}

func (dao *templateDAO) DetailTemplate(ctx context.Context, id int64) (Template, error) {
	col := dao.db.Collection(TemplateCollection)
	filter := bson.M{"id": id}

	var t Template
	if err := col.FindOne(ctx, filter).Decode(&t); err != nil {
		return Template{}, fmt.Errorf("解码错误，%w", err)
	}

	return t, nil
}

func (dao *templateDAO) DeleteTemplate(ctx context.Context, id int64) (int64, error) {
	col := dao.db.Collection(TemplateCollection)
	filter := bson.M{"id": id}

	result, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *templateDAO) UpdateTemplate(ctx context.Context, t Template) (int64, error) {
	col := dao.db.Collection(TemplateCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"name":        t.Name,
			"workflow_id": t.WorkflowId,
			"group_id":    t.GroupId,
			"icon":        t.Icon,
			"desc":        t.Desc,
			"rules":       t.Rules,
			"options":     t.Options,
			"utime":       time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": t.Id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *templateDAO) ListTemplate(ctx context.Context, offset, limit int64) ([]Template, error) {
	col := dao.db.Collection(TemplateCollection)
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

	var result []Template
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *templateDAO) Count(ctx context.Context) (int64, error) {
	col := dao.db.Collection(TemplateCollection)
	filer := bson.M{}

	count, err := col.CountDocuments(ctx, filer)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *templateDAO) Pipeline(ctx context.Context) ([]TemplatePipeline, error) {
	col := dao.db.Collection(TemplateCollection)
	filters := bson.M{"create_type": 1}
	pipeline := mongo.Pipeline{
		{{"$match", filters}},
		{{"$group", bson.D{
			{"_id", "$group_id"},
			{"total", bson.D{{"$sum", 1}}},
			// 使用 $push 累加器将选择的字段添加到 templates 数组中
			{"templates", bson.D{{"$push", bson.D{
				{"icon", "$icon"},
				{"name", "$name"},
				{"id", "$id"},
			}}}},
		}}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []TemplatePipeline
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}

	return result, nil
}
