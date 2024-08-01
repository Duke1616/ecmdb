package dao

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const TaskCollection = "c_task"

type TaskDAO interface {
	CreateTask(ctx context.Context, t Task) (int64, error)
	FindByProcessInstId(ctx context.Context, processInstId int, nodeId string) (Task, error)
	FindById(ctx context.Context, id int64) (Task, error)
	UpdateTask(ctx context.Context, t Task) (int64, error)
	UpdateTaskStatus(ctx context.Context, req Task) (int64, error)
	UpdateVariables(ctx context.Context, id int64, variables string) (int64, error)
	ListTask(ctx context.Context, offset, limit int64) ([]Task, error)
	ListTaskByStatus(ctx context.Context, offset, limit int64, status uint8) ([]Task, error)
	Count(ctx context.Context, status uint8) (int64, error)
	UpdateArgs(ctx context.Context, id int64, args map[string]interface{}) (int64, error)
	ListTasksByCtime(ctx context.Context, offset, limit int64, ctime int64) ([]Task, error)
	TotalByCtime(ctx context.Context, ctime int64) (int64, error)
}

type taskDAO struct {
	db *mongox.Mongo
}

func (dao *taskDAO) ListTasksByCtime(ctx context.Context, offset, limit int64, ctime int64) ([]Task, error) {
	col := dao.db.Collection(TaskCollection)
	filter := bson.M{}
	filter["ctime"] = bson.M{"$gte": ctime}

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

	var result []Task
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *taskDAO) TotalByCtime(ctx context.Context, ctime int64) (int64, error) {
	col := dao.db.Collection(TaskCollection)
	filter := bson.M{}
	filter["ctime"] = bson.M{"$lte": ctime}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *taskDAO) ListTask(ctx context.Context, offset, limit int64) ([]Task, error) {
	col := dao.db.Collection(TaskCollection)
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

	var result []Task
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *taskDAO) UpdateVariables(ctx context.Context, id int64, variables string) (int64, error) {
	col := dao.db.Collection(TaskCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"variables": variables,
			"utime":     time.Now().UnixMilli(),
		},
	}

	filter := bson.M{"id": id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *taskDAO) FindById(ctx context.Context, id int64) (Task, error) {
	col := dao.db.Collection(TaskCollection)
	var t Task
	filter := bson.M{}
	filter["id"] = id

	if err := col.FindOne(ctx, filter).Decode(&t); err != nil {
		return Task{}, fmt.Errorf("解码错误，%w", err)
	}

	return t, nil
}

func (dao *taskDAO) UpdateArgs(ctx context.Context, id int64, args map[string]interface{}) (int64, error) {
	col := dao.db.Collection(TaskCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"args":  args,
			"utime": time.Now().UnixMilli(),
		},
	}

	filter := bson.M{"id": id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *taskDAO) FindByProcessInstId(ctx context.Context, processInstId int, nodeId string) (Task, error) {
	col := dao.db.Collection(TaskCollection)
	var t Task
	filter := bson.M{}
	filter["process_inst_id"] = processInstId
	filter["current_node_id"] = nodeId

	if err := col.FindOne(ctx, filter).Decode(&t); err != nil {
		return Task{}, fmt.Errorf("解码错误，%w", err)
	}

	return t, nil
}

func NewTaskDAO(db *mongox.Mongo) TaskDAO {
	return &taskDAO{
		db: db,
	}
}

func (dao *taskDAO) UpdateTask(ctx context.Context, t Task) (int64, error) {
	col := dao.db.Collection(TaskCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"code":             t.Code,
			"order_id":         t.OrderId,
			"worker_name":      t.WorkerName,
			"codebook_uid":     t.CodebookUid,
			"workflow_id":      t.WorkflowId,
			"topic":            t.Topic,
			"language":         t.Language,
			"args":             t.Args,
			"variables":        t.Variables,
			"status":           t.Status,
			"result":           t.Result,
			"trigger_position": t.TriggerPosition,
			"utime":            time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": t.Id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *taskDAO) UpdateTaskStatus(ctx context.Context, t Task) (int64, error) {
	col := dao.db.Collection(TaskCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"result":           t.Result,
			"status":           t.Status,
			"utime":            time.Now().UnixMilli(),
			"trigger_position": t.TriggerPosition,
		},
	}
	filter := bson.M{"id": t.Id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *taskDAO) CreateTask(ctx context.Context, t Task) (int64, error) {
	now := time.Now()
	t.Ctime, t.Utime = now.UnixMilli(), now.UnixMilli()
	t.Id = dao.db.GetIdGenerator(TaskCollection)
	col := dao.db.Collection(TaskCollection)

	_, err := col.InsertOne(ctx, t)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return t.Id, nil
}

func (dao *taskDAO) ListTaskByStatus(ctx context.Context, offset, limit int64, status uint8) ([]Task, error) {
	col := dao.db.Collection(TaskCollection)
	filter := bson.M{}
	if status != 0 {
		filter["status"] = status
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

	var result []Task
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *taskDAO) Count(ctx context.Context, status uint8) (int64, error) {
	col := dao.db.Collection(TaskCollection)
	filter := bson.M{}
	if status != 0 {
		filter["status"] = status
	}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

type Task struct {
	Id              int64                  `bson:"id"`
	OrderId         int64                  `bson:"order_id"`
	ProcessInstId   int                    `bson:"process_inst_id"`
	CodebookUid     string                 `bson:"codebook_uid"`
	WorkerName      string                 `bson:"worker_name"`
	WorkflowId      int64                  `bson:"workflow_id"`
	Code            string                 `bson:"code"`
	Topic           string                 `bson:"topic"`
	Language        string                 `bson:"language"`
	Args            map[string]interface{} `bson:"args"`
	Variables       []Variables            `json:"variables"`
	Status          uint8                  `bson:"status"`
	Result          string                 `bson:"result"`
	TriggerPosition string                 `bson:"trigger_position"`
	CurrentNodeId   string                 `bson:"current_node_id"`
	Ctime           int64                  `bson:"ctime"`
	Utime           int64                  `bson:"utime"`
}

type Variables struct {
	Key    string `bson:"key"`
	Value  any    `bson:"value"`
	Secret bool   `bson:"secret"`
}
