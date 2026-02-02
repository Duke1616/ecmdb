package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/pkg/mongox"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	SnapshotCollection = "c_workflow_snapshot"
)

type SnapshotDAO interface {
	// Create 创建快照记录
	Create(ctx context.Context, snapshot Snapshot) error
	// FindByProcess 根据流程引擎ID和版本号查找快照
	FindByProcess(ctx context.Context, processID, version int) (Snapshot, error)
}

type snapshotDAO struct {
	db *mongox.Mongo
}

func NewSnapshotDAO(db *mongox.Mongo) SnapshotDAO {
	return &snapshotDAO{
		db: db,
	}
}

type Snapshot struct {
	ID             int64     `bson:"id"`
	WorkflowID     int       `bson:"workflow_id"`
	ProcessID      int       `bson:"process_id"`
	ProcessVersion int       `bson:"process_version"`
	Name           string    `bson:"name"`
	FlowData       LogicFlow `bson:"flow_data"`
	Ctime          int64     `bson:"ctime"`
}

func (dao *snapshotDAO) Create(ctx context.Context, snapshot Snapshot) error {
	snapshot.ID = dao.db.GetIdGenerator(SnapshotCollection)
	snapshot.Ctime = time.Now().UnixMilli()

	col := dao.db.Collection(SnapshotCollection)
	_, err := col.InsertOne(ctx, snapshot)
	if err != nil {
		return fmt.Errorf("插入快照错误: %w", err)
	}

	return nil
}

func (dao *snapshotDAO) FindByProcess(ctx context.Context, processID, version int) (Snapshot, error) {
	col := dao.db.Collection(SnapshotCollection)
	var s Snapshot
	filter := bson.M{
		"process_id":      processID,
		"process_version": version,
	}

	if err := col.FindOne(ctx, filter).Decode(&s); err != nil {
		return Snapshot{}, fmt.Errorf("查询快照错误: %w", err)
	}

	return s, nil
}
