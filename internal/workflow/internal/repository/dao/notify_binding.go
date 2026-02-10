package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const NotifyBindingCollection = "c_workflow_notify_binding"

type NotifyType string

type NotifyBindingDAO interface {
	// Create 创建
	Create(ctx context.Context, n NotifyBinding) (int64, error)
	// Update 更新
	Update(ctx context.Context, n NotifyBinding) (int64, error)
	// Delete 删除
	Delete(ctx context.Context, id int64) (int64, error)
	// Find 查询流程下的所有绑定
	Find(ctx context.Context, workflowId int64) ([]NotifyBinding, error)
	// FindBinding 获取生效的配置 (含默认兜底逻辑)
	FindBinding(ctx context.Context, workflowId int64, notifyType NotifyType, channel string) (NotifyBinding, error)
}

type notifyBindingDAO struct {
	db *mongox.Mongo
}

func NewNotifyBindingDAO(db *mongox.Mongo) NotifyBindingDAO {
	return &notifyBindingDAO{
		db: db,
	}
}

type NotifyBinding struct {
	Id         int64      `bson:"id"`          // 记录ID
	WorkflowId int64      `bson:"workflow_id"` // 流程ID
	NotifyType NotifyType `bson:"notify_type"` // 通知类型: 抄送、审批等
	Channel    string     `bson:"channel"`     // 通知渠道
	TemplateId int64      `bson:"template_id"` // 通知模版ID
	Ctime      int64      `bson:"ctime"`       // 创建时间
	Utime      int64      `bson:"utime"`       // 修改时间
}

func (dao *notifyBindingDAO) Create(ctx context.Context, n NotifyBinding) (int64, error) {
	n.Id = dao.db.GetIdGenerator(NotifyBindingCollection)
	col := dao.db.Collection(NotifyBindingCollection)
	now := time.Now().UnixMilli()
	n.Ctime, n.Utime = now, now

	_, err := col.InsertOne(ctx, n)
	if err != nil {
		return 0, fmt.Errorf("插入数据错误: %w", err)
	}

	return n.Id, nil
}

func (dao *notifyBindingDAO) Update(ctx context.Context, n NotifyBinding) (int64, error) {
	col := dao.db.Collection(NotifyBindingCollection)
	updateDoc := bson.M{
		"$set": bson.M{
			"notify_type": n.NotifyType,
			"channel":     n.Channel,
			"template_id": n.TemplateId,
			"utime":       time.Now().UnixMilli(),
		},
	}
	filter := bson.M{"id": n.Id}

	result, err := col.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return 0, fmt.Errorf("修改文档错误: %w", err)
	}

	return result.ModifiedCount, nil
}

func (dao *notifyBindingDAO) Delete(ctx context.Context, id int64) (int64, error) {
	col := dao.db.Collection(NotifyBindingCollection)
	filter := bson.M{"id": id}

	result, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("删除文档错误: %w", err)
	}

	return result.DeletedCount, nil
}

func (dao *notifyBindingDAO) Find(ctx context.Context, workflowId int64) ([]NotifyBinding, error) {
	col := dao.db.Collection(NotifyBindingCollection)
	filter := bson.M{"workflow_id": workflowId}

	cursor, err := col.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("查询错误: %w", err)
	}
	defer cursor.Close(ctx)

	var result []NotifyBinding
	if err = cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("解码错误: %w", err)
	}

	return result, nil
}

// FindBinding 查找绑定关系 (支持默认配置)
// 优先使用指定 WorkflowId 的配置, 如果没有则使用默认配置 (WorkflowId=0)
func (dao *notifyBindingDAO) FindBinding(ctx context.Context, workflowId int64, notifyType NotifyType,
	channel string) (NotifyBinding, error) {
	col := dao.db.Collection(NotifyBindingCollection)
	filter := bson.M{
		"workflow_id": bson.M{"$in": []int64{workflowId, 0}},
		"notify_type": notifyType,
		"channel":     channel,
	}

	// 按 workflow_id 倒序排列，确保优先匹配指定 workflowId 的记录 (非 0)
	opts := options.FindOne().SetSort(bson.M{"workflow_id": -1})

	var result NotifyBinding
	if err := col.FindOne(ctx, filter, opts).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return NotifyBinding{}, nil // Return empty if not found, or use a specific error
		}
		return NotifyBinding{}, fmt.Errorf("解码错误: %w", err)
	}

	return result, nil
}
