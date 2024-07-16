package dao

import (
	"context"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"gorm.io/gorm"
	"time"
)

type ProcessEngineDAO interface {
	CountTodo(ctx context.Context, userId, processName string) (int64, error)
	CountStartUser(ctx context.Context, userId, processName string) (int64, error)
	ListStartUser(ctx context.Context, userId, processName string, offset, limit int) (
		[]Instance, error)
}

type processEngineDAO struct {
	db *gorm.DB
}

func NewProcessEngineDAO(db *gorm.DB) ProcessEngineDAO {
	return &processEngineDAO{
		db: db,
	}
}

func (g *processEngineDAO) CountTodo(ctx context.Context, userId, processName string) (int64, error) {
	var res int64
	db := g.db.WithContext(ctx).Model(&model.Task{}).Table("proc_task")
	// 根据 userId 是否为空添加条件
	if userId != "" {
		db = db.Where("user_id = ?", userId)
	}
	if processName != "" {
		db = db.Where("process_name = ?", processName)
	}

	db = db.Where("is_finished = ?", 0)
	err := db.Count(&res).Error
	return res, err
}

func (g *processEngineDAO) CountStartUser(ctx context.Context, userId, processName string) (int64, error) {
	var res int64
	db := g.db.WithContext(ctx).Model(&model.Instance{}).Table("proc_inst")
	// 根据 userId 是否为空添加条件
	if userId != "" {
		db = db.Where("starter = ?", userId)
	}
	if processName != "" {
		db = db.Where("name = ?", processName)
	}

	err := db.Count(&res).Error
	return res, err
}

func (g *processEngineDAO) ListStartUser(ctx context.Context, userId, processName string, offset, limit int) (
	[]Instance, error) {
	var res []Instance
	// TODO 当前默认审批人是只有一个，不然 JOIN proc_task 会存在问题，后期需要单独抽出函数处理
	db := g.db.WithContext(ctx).Table("proc_inst as a").Select("a.id, a.proc_id, a.proc_version, " +
		"a.business_id, a.starter, a.current_node_id, c.node_name as current_node_name, a.create_time, " +
		"a.status, b.name, c.id as task_id, c.user_id").
		Joins("JOIN proc_def b ON a.proc_id = b.id").
		Joins("JOIN proc_task c ON a.id = c.proc_inst_id AND a.current_node_id = c.node_id " +
			"AND a.starter = c.starter").
		Limit(limit).
		Offset(offset)

	if userId != "" {
		db = db.Where("c.starter = ?", userId)
	}
	if processName != "" {
		db = db.Where("name = ?", processName)
	}

	err := db.Scan(&res).Error
	return res, err
}

type Instance struct {
	TaskID          int        `gorm:"column:task_id;"`          //任务ID
	ProcInstID      int        `gorm:"column:id;"`               //流程实例ID
	ProcID          int        `gorm:"column:proc_id"`           //流程ID
	ProcName        string     `gorm:"column:name"`              //流程名称
	ProcVersion     int        `gorm:"column:proc_version"`      //流程版本号
	BusinessID      string     `gorm:"column:business_id"`       //业务ID
	Starter         string     `gorm:"column:starter"`           //流程发起人用户ID
	CurrentNodeID   string     `gorm:"column:current_node_id"`   //当前进行节点ID
	CurrentNodeName string     `gorm:"column:current_node_name"` //当前进行节点名称
	CreateTime      *time.Time `gorm:"column:create_time"`       //创建时间
	ApprovedBy      string     `gorm:"column:user_id"`           //创建时间
	Status          int        `gorm:"column:status"`            //0:未完成(审批中) 1:已完成(通过) 2:撤销
}
