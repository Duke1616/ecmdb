package dao

import (
	"context"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitIndexes(db *mongox.Mongo) error {
	col := db.Collection(TaskCollection)

	indexes := []mongo.IndexModel{
		// 1. 业务主键 ID (唯一索引)
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		// 2. 调度器核心索引：捞取已就绪的定时任务
		// 查询模式：{ status: WAITING, is_timing: true, scheduled_time: { $lte: now } }
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "is_timing", Value: 1},
				{Key: "scheduled_time", Value: 1},
			},
		},
		// 3. 流程实例 + 节点 ID (复合索引，加速工作流状态查询)
		// 查询模式：{ process_inst_id: xxx, current_node_id: xxx }
		{
			Keys: bson.D{
				{Key: "process_inst_id", Value: 1},
				{Key: "current_node_id", Value: 1},
			},
		},
		// 4. 流程轨迹扫描索引 (加速按实例列表展示)
		{
			Keys: bson.D{{Key: "process_inst_id", Value: 1}, {Key: "ctime", Value: -1}},
		},
		// 5. 成功任务同步索引 (用于外部系统同步已成功的任务)
		// 查询模式：{ status: SUCCESS, mark_passed: false, utime: { $gte: xxx } }
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "mark_passed", Value: 1},
				{Key: "utime", Value: 1},
			},
		},
		// 6. 状态分布统计/分页索引
		{
			Keys: bson.D{{Key: "status", Value: 1}, {Key: "ctime", Value: -1}},
		},
	}

	_, err := col.Indexes().CreateMany(context.Background(), indexes)

	return err
}
