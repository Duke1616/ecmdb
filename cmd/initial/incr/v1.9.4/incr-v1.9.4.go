package v194

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/cmd/initial/backup"
	"github.com/Duke1616/ecmdb/cmd/initial/incr"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gotomicro/ego/core/elog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	WorkFlowCollection = "c_workflow"
	SnapshotCollection = "c_workflow_snapshot"
)

type Workflow struct {
	Id        int64     `bson:"id"`
	Name      string    `bson:"name"`
	ProcessId int       `bson:"process_id"`
	FlowData  LogicFlow `bson:"flow_data"`
}

type LogicFlow struct {
	Edges []map[string]interface{} `bson:"edges"`
	Nodes []map[string]interface{} `bson:"nodes"`
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

type incrV194 struct {
	App    *ioc.App
	logger elog.Component
}

func NewIncrV194(app *ioc.App) incr.InitialIncr {
	return &incrV194{
		App:    app,
		logger: *elog.DefaultLogger,
	}
}

func (i *incrV194) Version() string {
	return "v1.9.4"
}

func (i *incrV194) Commit(ctx context.Context) error {
	i.logger.Info("开始执行 Commit", elog.String("版本", i.Version()))

	limit := int64(200)
	offset := int64(0)

	for {
		// 1. 分页获取 Workflow
		wfs, err := i.fetchWorkflows(ctx, offset, limit)
		if err != nil {
			return err
		}
		if len(wfs) == 0 {
			break
		}

		// 2. 提取 ProcessIDs 并获取对应的版本号
		pids := slice.FilterMap(wfs, func(idx int, src Workflow) (int, bool) {
			return src.ProcessId, src.ProcessId != 0
		})

		if len(pids) == 0 {
			offset += limit
			continue
		}

		pidToVerMap, err := i.getProcessVersions(ctx, pids)
		if err != nil {
			return err
		}

		// 3. 构建快照候选列表
		candidates := slice.FilterMap(wfs, func(idx int, src Workflow) (Snapshot, bool) {
			ver, ok := pidToVerMap[src.ProcessId]
			if !ok {
				return Snapshot{}, false
			}
			return Snapshot{
				WorkflowID:     int(src.Id),
				ProcessID:      src.ProcessId,
				ProcessVersion: ver,
				Name:           src.Name,
				FlowData:       src.FlowData,
			}, true
		})

		// 4. 过滤掉已存在的快照并批量插入
		if err = i.createMissingSnapshots(ctx, candidates); err != nil {
			return err
		}

		offset += limit
	}

	i.logger.Info("Commit 执行完成", elog.String("版本", i.Version()))
	return nil
}

func (i *incrV194) fetchWorkflows(ctx context.Context, offset, limit int64) ([]Workflow, error) {
	col := i.App.DB.Collection(WorkFlowCollection)
	cursor, err := col.Find(ctx, bson.M{}, &options.FindOptions{
		Limit: &limit,
		Skip:  &offset,
	})
	if err != nil {
		return nil, fmt.Errorf("查询 Workflow 失败: %w", err)
	}
	defer cursor.Close(ctx)

	var wfs []Workflow
	if err = cursor.All(ctx, &wfs); err != nil {
		return nil, fmt.Errorf("解码 Workflow 失败: %w", err)
	}
	return wfs, nil
}

func (i *incrV194) getProcessVersions(ctx context.Context, pids []int) (map[int]int, error) {
	type ProcDef struct {
		Id      int `gorm:"column:id"`
		Version int `gorm:"column:version"`
	}
	var procDefs []ProcDef

	if err := i.App.GormDB.WithContext(ctx).Table("proc_def").
		Select("id, version").
		Where("id IN ?", pids).
		Find(&procDefs).Error; err != nil {
		return nil, fmt.Errorf("批量查询 proc_def 失败: %w", err)
	}

	return slice.ToMapV(procDefs, func(element ProcDef) (int, int) {
		return element.Id, element.Version
	}), nil
}

func (i *incrV194) createMissingSnapshots(ctx context.Context, candidates []Snapshot) error {
	if len(candidates) == 0 {
		return nil
	}

	col := i.App.DB.Collection(SnapshotCollection)

	// 构建查询条件来查找已存在的快照
	orConditions := slice.Map(candidates, func(idx int, src Snapshot) bson.M {
		return bson.M{
			"process_id":      src.ProcessID,
			"process_version": src.ProcessVersion,
		}
	})

	cursor, err := col.Find(ctx, bson.M{"$or": orConditions})
	if err != nil {
		return fmt.Errorf("批量查询 Snapshot 失败: %w", err)
	}
	defer cursor.Close(ctx)

	type SnapshotLite struct {
		ProcessID      int `bson:"process_id"`
		ProcessVersion int `bson:"process_version"`
	}
	var existing []SnapshotLite
	if err = cursor.All(ctx, &existing); err != nil {
		return fmt.Errorf("解码 SnapshotLite 失败: %w", err)
	}

	existingMap := make(map[string]bool, len(existing))
	for _, s := range existing {
		existingMap[fmt.Sprintf("%d-%d", s.ProcessID, s.ProcessVersion)] = true
	}

	// 过滤出真正需要插入的
	toInsert := slice.FilterMap(candidates, func(idx int, src Snapshot) (interface{}, bool) {
		key := fmt.Sprintf("%d-%d", src.ProcessID, src.ProcessVersion)
		if existingMap[key] {
			return nil, false
		}
		// 补全 ID 和 时间
		src.ID = i.App.DB.GetIdGenerator(SnapshotCollection)
		src.Ctime = time.Now().UnixMilli()
		return src, true
	})

	if len(toInsert) > 0 {
		if _, err = col.InsertMany(ctx, toInsert); err != nil {
			return fmt.Errorf("批量插入 Snapshot 失败: %w", err)
		}
		i.logger.Info("成功批量补全快照", elog.Int("count", len(toInsert)))
	}

	return nil
}

func (i *incrV194) Rollback(ctx context.Context) error {
	i.logger.Info("开始执行 Rollback", elog.String("版本", i.Version()))
	i.logger.Info("Rollback 执行完成", elog.String("版本", i.Version()))
	return nil
}

func (i *incrV194) Before(ctx context.Context) error {
	i.logger.Info("开始执行 Before，备份数据", elog.String("版本", i.Version()))

	backupManager := backup.NewBackupManager(i.App)
	opts := backup.Options{
		Version:     i.Version(),
		Description: "vv1.9.4 版本更新前备份",
		Tags: map[string]string{
			"type":   "version_upgrade",
			"module": "workflow",
		},
	}

	if _, err := backupManager.BackupMongoCollection(ctx, SnapshotCollection, opts); err != nil {
		return err
	}

	i.logger.Info("Before 执行完成，数据备份完成")
	return nil
}

func (i *incrV194) After(ctx context.Context) error {
	i.logger.Info("开始执行 After，更新版本信息", elog.String("版本", i.Version()))
	if err := i.App.VerSvc.CreateOrUpdateVersion(ctx, i.Version()); err != nil {
		i.logger.Error("更新版本信息失败", elog.FieldErr(err))
		return err
	}
	i.logger.Info("After 执行完成，版本信息已更新")
	return nil
}
