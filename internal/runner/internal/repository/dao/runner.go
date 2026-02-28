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

const (
	RunnerCollection = "c_runner"
)

// RunnerDAO 执行器数据访问接口
// 提供与底层数据库（如 MongoDB）交互的基础 CRUD 及聚合统计方法
type RunnerDAO interface {
	// Create 在数据库中插入一条新的执行器文档记录
	Create(ctx context.Context, r Runner) (int64, error)
	// Update 局部更新指定的执行器文档，内部维护修改时间等信息
	Update(ctx context.Context, req Runner) (int64, error)
	// Delete 根据自增或逻辑 ID 删除对应的执行器存盘文档
	Delete(ctx context.Context, id int64) (int64, error)
	// FindById 根据逻辑 ID 从数据库加载执行器文档实体
	FindById(ctx context.Context, id int64) (Runner, error)
	// List 返回根据时间戳等规则排序的执行器分页文档列表
	List(ctx context.Context, offset, limit int64, keyword, kind string) ([]Runner, error)
	// Count 统计数据库中当前有效的执行器总文档数
	Count(ctx context.Context, keyword, kind string) (int64, error)
	// FindByCodebookUidAndTag 通过参数 UID 和内部 tags 匹配查找对应的执行器文档
	FindByCodebookUidAndTag(ctx context.Context, codebookUid string, tag string) (Runner, error)
	// ListByCodebookUid 查出关联到对应独立脚本 UID 的所有挂载执行器文档节点
	ListByCodebookUid(ctx context.Context, offset, limit int64, codebookUid, keyword, kind string) ([]Runner, error)
	// CountByCodebookUid 统计通过脚本 UID 获取的具有承载特性的数据量
	CountByCodebookUid(ctx context.Context, codebookUid, keyword, kind string) (int64, error)
	// ListExcludeCodebookUid 用以返回过滤掉指定脚本 UID 后剩余可用的那些备选执行器列表
	ListExcludeCodebookUid(ctx context.Context, offset, limit int64, codebookUid, keyword, kind string) ([]Runner, error)
	// CountExcludeCodebookUid 统计未关联某特征 UID 剩余执行器的池大小
	CountExcludeCodebookUid(ctx context.Context, codebookUid, keyword, kind string) (int64, error)
	// ListByCodebookUids 通过脚本 UID 列表，用 $in 批量获取其对应的执行器文档
	ListByCodebookUids(ctx context.Context, codebookUids []string) ([]Runner, error)
	// ListByIds 使用给定的指定 ID 列表过滤拉取所有的关联文档
	ListByIds(ctx context.Context, ids []int64) ([]Runner, error)
	// AggregateTags 利用 Mongo pipeline 等方式聚合每个脚本绑定的节点和队列标记
	AggregateTags(ctx context.Context) ([]RunnerPipeline, error)
}

func NewRunnerDAO(db *mongox.Mongo) RunnerDAO {
	return &runnerDAO{
		db: db,
	}
}

type runnerDAO struct {
	db *mongox.Mongo
}

func (dao *runnerDAO) ListByIds(ctx context.Context, ids []int64) ([]Runner, error) {
	col := dao.db.Collection(RunnerCollection)
	filter := bson.M{"id": bson.M{"$in": ids}}

	cursor, err := col.Find(ctx, filter)
	var result []Runner
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *runnerDAO) ListByCodebookUids(ctx context.Context, codebookUids []string) ([]Runner, error) {
	col := dao.db.Collection(RunnerCollection)
	filter := bson.M{"codebook_uid": bson.M{"$in": codebookUids}}

	cursor, err := col.Find(ctx, filter)
	var result []Runner
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *runnerDAO) ListByCodebookUid(ctx context.Context, offset, limit int64, codebookUid, keyword, kind string) ([]Runner, error) {
	col := dao.db.Collection(RunnerCollection)
	filter := bson.M{"codebook_uid": codebookUid}
	if keyword != "" {
		filter["name"] = bson.M{"$regex": keyword, "$options": "i"}
	}
	if kind != "" {
		filter["kind"] = kind
	}
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	cursor, err := col.Find(ctx, filter, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("按UID查询错误: %w", err)
	}

	var result []Runner
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *runnerDAO) CountByCodebookUid(ctx context.Context, codebookUid, keyword, kind string) (int64, error) {
	col := dao.db.Collection(RunnerCollection)
	filter := bson.M{"codebook_uid": codebookUid}
	if keyword != "" {
		filter["name"] = bson.M{"$regex": keyword, "$options": "i"}
	}
	if kind != "" {
		filter["kind"] = kind
	}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("按UID计数错误: %w", err)
	}

	return count, nil
}

func (dao *runnerDAO) ListExcludeCodebookUid(ctx context.Context, offset, limit int64, codebookUid, keyword, kind string) ([]Runner, error) {
	col := dao.db.Collection(RunnerCollection)
	filter := bson.M{"codebook_uid": bson.M{"$ne": codebookUid}}
	if keyword != "" {
		filter["name"] = bson.M{"$regex": keyword, "$options": "i"}
	}
	if kind != "" {
		filter["kind"] = kind
	}
	opts := &options.FindOptions{
		Sort:  bson.D{{Key: "ctime", Value: -1}},
		Limit: &limit,
		Skip:  &offset,
	}

	cursor, err := col.Find(ctx, filter, opts)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("排除UID查询错误: %w", err)
	}

	var result []Runner
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *runnerDAO) CountExcludeCodebookUid(ctx context.Context, codebookUid, keyword, kind string) (int64, error) {
	col := dao.db.Collection(RunnerCollection)
	filter := bson.M{"codebook_uid": bson.M{"$ne": codebookUid}}
	if keyword != "" {
		filter["name"] = bson.M{"$regex": keyword, "$options": "i"}
	}
	if kind != "" {
		filter["kind"] = kind
	}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("排除UID计数错误: %w", err)
	}

	return count, nil
}

func (dao *runnerDAO) FindById(ctx context.Context, id int64) (Runner, error) {
	col := dao.db.Collection(RunnerCollection)
	filter := bson.M{"id": id}

	var result Runner
	if err := col.FindOne(ctx, filter).Decode(&result); err != nil {
		return Runner{}, fmt.Errorf("解码错误，%w", err)
	}

	return result, nil
}

func (dao *runnerDAO) Delete(ctx context.Context, id int64) (int64, error) {
	col := dao.db.Collection(RunnerCollection)
	filter := bson.M{"id": id}

	result, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *runnerDAO) Update(ctx context.Context, req Runner) (int64, error) {
	col := dao.db.Collection(RunnerCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"name":            req.Name,
			"codebook_secret": req.CodebookSecret,
			"kind":            req.Kind,
			"target":          req.Target,
			"handler":         req.Handler,
			"tags":            req.Tags,
			"desc":            req.Desc,
			"variables":       req.Variables,
			"utime":           time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": req.Id}
	count, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档操作: %w", err)
	}

	return count.ModifiedCount, nil
}

func (dao *runnerDAO) FindByCodebookUidAndTag(ctx context.Context, codebookUid string, tag string) (Runner, error) {
	col := dao.db.Collection(RunnerCollection)
	filter := bson.M{}
	filter["codebook_uid"] = codebookUid
	filter["tags"] = bson.M{
		"$elemMatch": bson.M{"$eq": tag},
	}

	var result Runner
	if err := col.FindOne(ctx, filter).Decode(&result); err != nil {
		return Runner{}, fmt.Errorf("解码错误，%w", err)
	}

	return result, nil
}

func (dao *runnerDAO) Create(ctx context.Context, r Runner) (int64, error) {
	r.Id = dao.db.GetIdGenerator(RunnerCollection)
	col := dao.db.Collection(RunnerCollection)
	now := time.Now()
	r.Ctime, r.Utime = now.UnixMilli(), now.UnixMilli()

	_, err := col.InsertOne(ctx, r)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return r.Id, nil
}

func (dao *runnerDAO) List(ctx context.Context, offset, limit int64, keyword, kind string) ([]Runner, error) {
	col := dao.db.Collection(RunnerCollection)
	filter := bson.M{}
	if keyword != "" {
		filter["name"] = bson.M{"$regex": keyword, "$options": "i"}
	}
	if kind != "" {
		filter["kind"] = kind
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

	var result []Runner
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}
	return result, nil
}

func (dao *runnerDAO) Count(ctx context.Context, keyword, kind string) (int64, error) {
	col := dao.db.Collection(RunnerCollection)
	filter := bson.M{}
	if keyword != "" {
		filter["name"] = bson.M{"$regex": keyword, "$options": "i"}
	}
	if kind != "" {
		filter["kind"] = kind
	}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("文档计数错误: %w", err)
	}

	return count, nil
}

func (dao *runnerDAO) AggregateTags(ctx context.Context) ([]RunnerPipeline, error) {
	col := dao.db.Collection(RunnerCollection)
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$codebook_uid"},
			// 使用 $push 累加器将选择的字段添加到 runners 数组中
			{Key: "runner_tags", Value: bson.D{{Key: "$push", Value: bson.D{
				{Key: "tags", Value: "$tags"},
				{Key: "target", Value: "$target"},
				{Key: "handler", Value: "$handler"},
			}}}},
		}}},
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	defer cursor.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询错误, %w", err)
	}

	var result []RunnerPipeline
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("游标遍历错误: %w", err)
	}

	return result, nil
}

type Runner struct {
	Id             int64       `bson:"id"`
	Name           string      `bson:"name"`
	CodebookUid    string      `bson:"codebook_uid"`
	CodebookSecret string      `bson:"codebook_secret"`
	Kind           string      `bson:"kind"`
	Target         string      `bson:"target"`
	Handler        string      `bson:"handler"`
	Tags           []string    `bson:"tags"`
	Action         uint8       `bson:"action"`
	Desc           string      `bson:"desc"`
	Variables      []Variables `bson:"variables"`
	Ctime          int64       `bson:"ctime"`
	Utime          int64       `bson:"utime"`
}

type Variables struct {
	Key    string `bson:"key"`
	Value  string `bson:"value"`
	Secret bool   `bson:"secret"`
}

type RunnerPipeline struct {
	CodebookUid string       `bson:"_id"`
	RunnerTags  []RunnerTags `bson:"runner_tags"`
}

type RunnerTags struct {
	Target  string   `bson:"target"`
	Handler string   `bson:"handler"`
	Tags    []string `json:"tags"`
}
