package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/internal/task/domain"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const TaskCollection = "c_task"

type TaskDAO interface {
	// CreateTask 创建一个新的任务入库，内部会自动生成 Ctime 和 Utime
	CreateTask(ctx context.Context, t Task) (Task, error)

	// FindByProcessInstId 根据流程实例 ID 和节点 ID 查询对应的任务
	FindByProcessInstId(ctx context.Context, processInstId int, nodeId string) (Task, error)

	// FindById 根据主键 ID 获取唯一任务
	FindById(ctx context.Context, id int64) (Task, error)

	// UpdateTask 更新任务的全量核心信息，如：模版代码、主题、关联工单等
	UpdateTask(ctx context.Context, t Task) (int64, error)

	// UpdateTaskStatus 仅更新任务的执行状态、期望结果与实际执行成果等
	UpdateTaskStatus(ctx context.Context, req Task) (int64, error)

	// UpdateVariables 更新任务所绑定的环境变量信息
	UpdateVariables(ctx context.Context, id int64, variables []Variables) (int64, error)

	// ListTask 分页获取所有任务列表，按创建时间倒序
	ListTask(ctx context.Context, offset, limit int64) ([]Task, error)

	// ListTaskByStatus 分页获取特定状态下的任务列表，当 status 不为 0 时进行筛选匹配
	ListTaskByStatus(ctx context.Context, offset, limit int64, status uint8) ([]Task, error)

	// ListTaskByStatusAndMode 分页获取特定状态和运行模式的任务列表
	ListTaskByStatusAndMode(ctx context.Context, offset, limit int64, status uint8, mode string) ([]Task, error)

	// Count 统计指定状态下的任务总数，当 status 为 0 时获取全量文档数
	Count(ctx context.Context, status uint8) (int64, error)

	// CountByStatusAndMode 统计特定状态和运行模式的任务总数
	CountByStatusAndMode(ctx context.Context, status uint8, mode string) (int64, error)

	// UpdateArgs 更新任务的局部执行参数（UserArgs）
	UpdateArgs(ctx context.Context, id int64, args map[string]interface{}) (int64, error)

	// ListSuccessTasksByUtime 根据更新时间拉取已经执行成功且尚未被标记跳过的任务列表
	ListSuccessTasksByUtime(ctx context.Context, offset, limit int64, utime int64) ([]Task, error)

	// TotalByUtime 统计指定更新时间前已执行成功的任务总数
	TotalByUtime(ctx context.Context, utime int64) (int64, error)

	// FindTaskResult 根据流程实例获取对应节点的任务结果记录
	FindTaskResult(ctx context.Context, instanceId int, nodeId string) (Task, error)

	// ListReadyTasks 捞取已经准备好可以执行的 WAITING 任务（定时任务需满足执行时间）
	ListReadyTasks(ctx context.Context, limit int64) ([]Task, error)

	// ListTaskByInstanceId 根据工作流实例 ID 批量查阅此实例关联的所有任务分页列表
	ListTaskByInstanceId(ctx context.Context, offset, limit int64, instanceId int) ([]Task, error)

	// TotalByInstanceId 统计特定工作流实例名下的任务总数
	TotalByInstanceId(ctx context.Context, instanceId int) (int64, error)

	// MarkTaskAsAutoPassed 将对应任务的状态标识为自动通过（MarkPassed = true）
	MarkTaskAsAutoPassed(ctx context.Context, id int64) error

	// UpdateExternalId 绑定外部分布式平台的任务 ID
	UpdateExternalId(ctx context.Context, id int64, externalId string) error
}

type taskDAO struct {
	db *mongox.Mongo
}

func (dao *taskDAO) ListTaskByInstanceId(ctx context.Context, offset, limit int64, instanceId int) ([]Task, error) {
	col := dao.db.Collection(TaskCollection)
	filter := bson.M{}
	filter["process_inst_id"] = instanceId

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}
	defer cursor.Close(ctx)

	var result []Task
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *taskDAO) TotalByInstanceId(ctx context.Context, instanceId int) (int64, error) {
	col := dao.db.Collection(TaskCollection)
	filter := bson.M{}
	filter["process_inst_id"] = instanceId

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *taskDAO) FindTaskResult(ctx context.Context, instanceId int, nodeId string) (Task, error) {
	col := dao.db.Collection(TaskCollection)
	filter := bson.M{}
	filter["process_inst_id"] = instanceId
	filter["current_node_id"] = nodeId

	var result Task
	if err := col.FindOne(ctx, filter).Decode(&result); err != nil {
		return Task{}, fmt.Errorf("解码错误: %w", err)
	}

	return result, nil
}

func (dao *taskDAO) ListSuccessTasksByUtime(ctx context.Context, offset, limit int64, utime int64) ([]Task, error) {
	col := dao.db.Collection(TaskCollection)
	filter := bson.M{}
	filter["status"] = bson.M{"$eq": domain.SUCCESS}
	filter["utime"] = bson.M{"$gte": utime}
	filter["mark_passed"] = bson.M{"$eq": false}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}
	defer cursor.Close(ctx)

	var result []Task
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *taskDAO) TotalByUtime(ctx context.Context, utime int64) (int64, error) {
	col := dao.db.Collection(TaskCollection)
	filter := bson.M{}
	filter["status"] = bson.M{"$eq": domain.SUCCESS}
	filter["utime"] = bson.M{"$lte": utime}
	filter["mark_passed"] = bson.M{"$eq": false}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *taskDAO) MarkTaskAsAutoPassed(ctx context.Context, id int64) error {
	col := dao.db.Collection(TaskCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"mark_passed": true,
		},
	}
	filter := bson.M{"id": id}
	_, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("修改文档操作: %w", err)
	}

	return nil
}

func (dao *taskDAO) UpdateExternalId(ctx context.Context, id int64, externalId string) error {
	col := dao.db.Collection(TaskCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"external_id": externalId,
			"utime":       time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": id}
	_, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("修改文档操作: %w", err)
	}

	return nil
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
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}
	defer cursor.Close(ctx)

	var result []Task
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *taskDAO) UpdateVariables(ctx context.Context, id int64, variables []Variables) (int64, error) {
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
			"codebook_uid":     t.CodebookUid,
			"codebook_name":    t.CodebookName,
			"workflow_id":      t.WorkflowId,
			"language":         t.Language,
			"args":             t.Args,
			"variables":        t.Variables,
			"status":           t.Status,
			"result":           t.Result,
			"run_mode":         t.RunMode,
			"worker_name":      t.Worker.WorkerName,
			"topic":            t.Topic,
			"service_name":     t.Execute.ServiceName,
			"handler":          t.Execute.Handler,
			"external_id":      t.ExternalId,
			"trigger_position": t.TriggerPosition,
			"is_timing":        t.IsTiming,
			"scheduled_time":   t.ScheduledTime,
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
			"want_result":      t.WantResult,
			"utime":            time.Now().UnixMilli(),
			"trigger_position": t.TriggerPosition,
		},
	}

	if t.StartTime > 0 {
		updateDoc["$set"].(bson.M)["start_time"] = t.StartTime
	}
	if t.EndTime > 0 {
		updateDoc["$set"].(bson.M)["end_time"] = t.EndTime
	}
	if t.RetryCount > 0 {
		// NOTE: 使用 $inc 而非 $set，保证并发安全的原子递增
		updateDoc["$inc"] = bson.M{"retry_count": t.RetryCount}
	} else if t.RetryCount < 0 {
		// NOTE: 约定传 -1 时重置为 0，用于人工手动重试场景
		updateDoc["$set"].(bson.M)["retry_count"] = 0
	}

	filter := bson.M{"id": t.Id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *taskDAO) CreateTask(ctx context.Context, t Task) (Task, error) {
	now := time.Now()
	t.Ctime, t.Utime = now.UnixMilli(), now.UnixMilli()
	t.MarkPassed = false
	t.Id = dao.db.GetIdGenerator(TaskCollection)
	col := dao.db.Collection(TaskCollection)

	_, err := col.InsertOne(ctx, t)
	if err != nil {
		return Task{}, fmt.Errorf("插入数据错误: %w", err)
	}

	return t, nil
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
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}
	defer cursor.Close(ctx)

	var result []Task
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *taskDAO) ListTaskByStatusAndMode(ctx context.Context, offset, limit int64, status uint8, mode string) ([]Task, error) {
	col := dao.db.Collection(TaskCollection)
	filter := bson.M{}
	if status != 0 {
		filter["status"] = status
	}
	if mode != "" {
		filter["run_mode"] = mode
	}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}
	defer cursor.Close(ctx)

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

func (dao *taskDAO) CountByStatusAndMode(ctx context.Context, status uint8, mode string) (int64, error) {
	col := dao.db.Collection(TaskCollection)
	filter := bson.M{}
	if status != 0 {
		filter["status"] = status
	}
	if mode != "" {
		filter["run_mode"] = mode
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
	CodebookName    string                 `bson:"codebook_name"`
	CodebookUid     string                 `bson:"codebook_uid"`
	WorkerName      string                 `bson:"worker_name"`
	WorkflowId      int64                  `bson:"workflow_id"`
	Code            string                 `bson:"code"`
	Topic           string                 `bson:"topic"`
	Language        string                 `bson:"language"`
	Args            map[string]interface{} `bson:"args"`
	Variables       []Variables            `bson:"variables"`
	RunMode         string                 `bson:"run_mode"`
	Worker          Worker                 `bson:",inline"`
	Execute         Execute                `bson:",inline"`
	ExternalId      string                 `bson:"external_id"`
	Status          uint8                  `bson:"status"`
	Result          string                 `bson:"result"`
	WantResult      string                 `bson:"want_result"`
	TriggerPosition string                 `bson:"trigger_position"`
	CurrentNodeId   string                 `bson:"current_node_id"`
	MarkPassed      bool                   `bson:"mark_passed"`
	Ctime           int64                  `bson:"ctime"`
	Utime           int64                  `bson:"utime"`
	IsTiming        bool                   `bson:"is_timing"`
	ScheduledTime   int64                  `bson:"scheduled_time"`
	StartTime       int64                  `bson:"start_time"`
	EndTime         int64                  `bson:"end_time"`
	RetryCount      int                    `bson:"retry_count"`
}

type Worker struct {
	WorkerName string `bson:"worker_name"`
	Topic      string `bson:"topic"`
}

type Execute struct {
	ServiceName string `bson:"service_name"`
	Handler     string `bson:"handler"`
}

type Variables struct {
	Key    string `bson:"key"`
	Value  string `bson:"value"`
	Secret bool   `bson:"secret"`
}

func (dao *taskDAO) ListReadyTasks(ctx context.Context, limit int64) ([]Task, error) {
	col := dao.db.Collection(TaskCollection)

	// 过滤条件：状态为 WAITING，且是定时任务，且计划执行时间已到
	now := time.Now().UnixMilli()
	filter := bson.M{
		"status":         domain.WAITING,
		"is_timing":      true,
		"scheduled_time": bson.M{"$lte": now},
	}

	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: 1}}, // 按创建时间正序，先入先出
		Limit: &limit,
	}

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}
	defer cursor.Close(ctx)

	var result []Task
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}
