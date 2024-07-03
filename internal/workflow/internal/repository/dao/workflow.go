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
	WorkFlowCollection = "c_workflow"
)

type WorkflowDAO interface {
	Create(ctx context.Context, w Workflow) (int64, error)
	List(ctx context.Context, offset, limit int64) ([]Workflow, error)
	Count(ctx context.Context) (int64, error)
}

func NewWorkflowDAO(db *mongox.Mongo) WorkflowDAO {
	return &workflowDAO{
		db: db,
	}
}

type workflowDAO struct {
	db *mongox.Mongo
}

func (dao *workflowDAO) Create(ctx context.Context, w Workflow) (int64, error) {
	w.Id = dao.db.GetIdGenerator(WorkFlowCollection)
	col := dao.db.Collection(WorkFlowCollection)
	now := time.Now()
	w.Ctime, w.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, w)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return w.Id, nil
}

func (dao *workflowDAO) List(ctx context.Context, offset, limit int64) ([]Workflow, error) {
	col := dao.db.Collection(WorkFlowCollection)
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

	var result []Workflow
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *workflowDAO) Count(ctx context.Context) (int64, error) {
	col := dao.db.Collection(WorkFlowCollection)
	filer := bson.M{}

	count, err := col.CountDocuments(ctx, filer)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

type Workflow struct {
	Id         int64                  `bson:"id"`
	TemplateId int64                  `bson:"template_id"`
	Name       string                 `bson:"name"`
	Icon       string                 `bson:"icon"`
	Owner      string                 `bson:"owner"`
	Desc       string                 `bson:"desc"`
	FlowData   map[string]interface{} `bson:"flow_data"`
	Ctime      int64                  `bson:"ctime"`
	Utime      int64                  `bson:"utime"`
}
