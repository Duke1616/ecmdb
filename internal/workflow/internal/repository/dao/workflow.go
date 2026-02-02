package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	WorkFlowCollection = "c_workflow"
)

type WorkflowDAO interface {
	// Create 创建流程定义
	Create(ctx context.Context, w Workflow) (int64, error)
	// List 分页查询流程定义列表
	List(ctx context.Context, offset, limit int64) ([]Workflow, error)
	// Count 统计流程定义总数
	Count(ctx context.Context) (int64, error)
	// Update 更新流程定义
	Update(ctx context.Context, c Workflow) (int64, error)
	// UpdateProcessId 绑定流程引擎ID
	UpdateProcessId(ctx context.Context, id int64, processId int) error
	// Delete 删除流程定义
	Delete(ctx context.Context, id int64) (int64, error)
	// Find 根据ID查询流程定义
	Find(ctx context.Context, id int64) (Workflow, error)
	// FindByKeyword 根据关键字搜索流程
	FindByKeyword(ctx context.Context, keyword string, offset, limit int64) ([]Workflow, error)
	// CountByKeyword 根据关键字统计流程总数
	CountByKeyword(ctx context.Context, keyword string) (int64, error)
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

func (dao *workflowDAO) Find(ctx context.Context, id int64) (Workflow, error) {
	col := dao.db.Collection(WorkFlowCollection)
	var w Workflow
	filter := bson.M{"id": id}

	if err := col.FindOne(ctx, filter).Decode(&w); err != nil {
		return Workflow{}, fmt.Errorf("解码错误，%w", err)
	}

	return w, nil
}

func (dao *workflowDAO) Delete(ctx context.Context, id int64) (int64, error) {
	col := dao.db.Collection(WorkFlowCollection)
	filter := bson.M{"id": id}

	result, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *workflowDAO) Update(ctx context.Context, c Workflow) (int64, error) {
	col := dao.db.Collection(WorkFlowCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"name":          c.Name,
			"desc":          c.Desc,
			"owner":         c.Owner,
			"is_notify":     c.IsNotify,
			"notify_method": c.NotifyMethod,
			"flow_data":     c.FlowData,
			"utime":         time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": c.Id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *workflowDAO) UpdateProcessId(ctx context.Context, id int64, processId int) error {
	col := dao.db.Collection(WorkFlowCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"process_id": processId,
			"utime":      time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": id}
	_, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("修改文档操作: %w", err)
	}

	return nil
}

type Workflow struct {
	Id           int64     `bson:"id"`
	TemplateId   int64     `bson:"template_id"`
	Name         string    `bson:"name"`
	Icon         string    `bson:"icon"`
	Owner        string    `bson:"owner"`
	Desc         string    `bson:"desc"`
	ProcessId    int       `bson:"process_id"`
	FlowData     LogicFlow `bson:"flow_data"`
	IsNotify     bool      `bson:"is_notify"`
	NotifyMethod uint8     `bson:"notify_method"`
	Ctime        int64     `bson:"ctime"`
	Utime        int64     `bson:"utime"`
}

type LogicFlow struct {
	Edges []map[string]interface{} `bson:"edges"`
	Nodes []map[string]interface{} `bson:"nodes"`
}

func (dao *workflowDAO) FindByKeyword(ctx context.Context, keyword string, offset, limit int64) ([]Workflow, error) {
	col := dao.db.Collection(WorkFlowCollection)

	// 默认为空查询
	filter := bson.M{}

	// 如果关键字不为空，则添加过滤条件
	if keyword != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": keyword, "$options": "i"}},
			{"desc": bson.M{"$regex": keyword, "$options": "i"}},
		}
	}

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

func (dao *workflowDAO) CountByKeyword(ctx context.Context, keyword string) (int64, error) {
	col := dao.db.Collection(WorkFlowCollection)

	// 默认为空查询
	filter := bson.M{}

	// 如果关键字不为空，则添加过滤条件
	if keyword != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": keyword, "$options": "i"}},
			{"desc": bson.M{"$regex": keyword, "$options": "i"}},
		}
	}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}
