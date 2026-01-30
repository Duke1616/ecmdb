package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Bunny3th/easy-workflow/workflow/database"
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"gorm.io/gorm"
)

type ProcessEngineDAO interface {
	CountTodo(ctx context.Context, userId, processName string) (int64, error)
	CountStartUser(ctx context.Context, userId, processName string) (int64, error)
	ListHistory(ctx context.Context, userId, processName string, offset, limit int)
	ListStartUser(ctx context.Context, userId, processName string, offset, limit int) (
		[]Instance, error)

	GetTasksByCurrentNodeId(ctx context.Context, processInstId int, currentNodeId string) ([]model.Task, error)

	ListTaskRecord(ctx context.Context, processInstId, offset, limit int) ([]model.Task, error)
	CountTaskRecord(ctx context.Context, processInstId int) (int64, error)
	SearchStartByProcessInstIds(ctx context.Context, processInstIds []int) ([]Instance, error)
	UpdateIsFinishedByPreNodeId(ctx context.Context, nodeId string, status int, comment string) error
	CountReject(ctx context.Context, taskId int) (int64, error)

	ListTasksByProcInstId(ctx context.Context, processInstIds []int, starter string) ([]model.Task, error)

	GetAutomationTask(ctx context.Context, currentNodeId string, processInstId int) (model.Task, error)
	GetTasksByInstUsers(ctx context.Context, processInstId int, userIds []string) ([]model.Task, error)
	GetOrderIdByVariable(ctx context.Context, processInstId int) (string, error)
	// GetProxyNodeID 获取代理转发的节点ID
	GetProxyNodeID(ctx context.Context, prevNodeID string) (model.Task, error)
}

type processEngineDAO struct {
	db *gorm.DB
}

func (g *processEngineDAO) GetProxyNodeID(ctx context.Context, prevNodeID string) (model.Task, error) {
	var node model.Task
	err := g.db.WithContext(ctx).Table("proc_task").First(&node,
		"node_id = ? AND status = ? AND is_finished = ?", prevNodeID, 0, 0).Error
	return node, err
}

func (g *processEngineDAO) GetTasksByCurrentNodeId(ctx context.Context, processInstId int, currentNodeId string) ([]model.Task, error) {
	var res []model.Task
	err := g.db.WithContext(ctx).Model(&model.Task{}).Table("proc_task").
		Where("proc_inst_id = ? AND status = ? AND is_finished = ? AND node_id = ?",
			processInstId, 0, 0, currentNodeId).
		Find(&res).Error

	return res, err
}

func (g *processEngineDAO) GetOrderIdByVariable(ctx context.Context, processInstId int) (string, error) {
	var res database.ProcInstVariable
	err := g.db.WithContext(ctx).Model(&database.ProcInstVariable{}).Table("proc_inst_variable").
		Where("proc_inst_id = ? AND `key` = ?", processInstId, `order_id`).
		First(&res).Error

	return res.Value, err
}

func (g *processEngineDAO) GetTasksByInstUsers(ctx context.Context, processInstId int,
	userIds []string) ([]model.Task, error) {
	var res []model.Task
	err := g.db.WithContext(ctx).Model(&model.Task{}).Table("proc_task").
		Where("proc_inst_id = ? AND status = ? AND is_finished = ? AND user_id IN ?",
			processInstId, 0, 0, userIds).
		Find(&res).Error

	return res, err
}

func (g *processEngineDAO) GetAutomationTask(ctx context.Context, currentNodeId string, processInstId int) (
	model.Task, error) {
	var res model.Task
	err := g.db.WithContext(ctx).Model(&model.Task{}).Table("proc_task").
		Where("node_id = ? AND proc_inst_id = ? AND is_finished = ? AND status = ?",
			currentNodeId, processInstId, 0, 0).
		First(&res).Error
	return res, err
}

func (g *processEngineDAO) ListTasksByProcInstId(ctx context.Context, processInstIds []int, starter string) (
	[]model.Task, error) {
	var res []model.Task
	err := g.db.WithContext(ctx).Model(&model.Task{}).Table("proc_task").
		Where("starter = ? AND is_finished = 0 AND proc_inst_id IN ?", starter, processInstIds).
		Find(&res).Error

	return res, err
}

func (g *processEngineDAO) CountReject(ctx context.Context, taskId int) (int64, error) {
	var res int64
	err := g.db.WithContext(ctx).Model(&database.ProcTask{}).
		Where("id = ? AND status = ?", taskId, 2).
		Select("COUNT(id)").Count(&res).Error
	return res, err
}

func (g *processEngineDAO) UpdateIsFinishedByPreNodeId(ctx context.Context, nodeId string, status int, comment string) error {
	proTask := database.ProcTask{Status: status, IsFinished: 1, Comment: comment,
		FinishedTime: database.LTime.Now()}

	return g.db.WithContext(ctx).
		Where("prev_node_id = ? AND is_finished = ? AND status = ?", nodeId, 0, 0).
		Updates(proTask).Error
}

func NewProcessEngineDAO(db *gorm.DB) ProcessEngineDAO {
	return &processEngineDAO{
		db: db,
	}
}

func (g *processEngineDAO) ListTaskRecord(ctx context.Context, processInstId, offset, limit int) ([]model.Task, error) {
	var res []model.Task
	procInstDb := g.db.WithContext(ctx).Model(&model.Task{}).Table("proc_task").
		Select("id, proc_id, proc_inst_id, business_id,starter,node_id,node_name,"+
			"prev_node_id,is_cosigned,batch_code,user_id,status,is_finished,comment,proc_inst_create_time,"+
			"create_time,finished_time").
		Where("proc_inst_id = ?", processInstId)
	procHistInstDb := g.db.WithContext(ctx).Model(&model.Task{}).Table("hist_proc_task").
		Select("id,proc_id, proc_inst_id,business_id,starter,node_id,node_name,"+
			"prev_node_id,is_cosigned,batch_code,user_id,status,is_finished,comment,proc_inst_create_time,"+
			"create_time,finished_time").
		Where("proc_inst_id = ?", processInstId)

	query := g.db.Raw("? UNION ALL ?", procInstDb, procHistInstDb)
	db := g.db.Table("(?) as a", query).Select("a.id,a.proc_id,b.name,a.proc_inst_id," +
		"a.business_id,a.starter,a.node_id,a.node_name,a.prev_node_id,a.is_cosigned,a.batch_code,a.user_id,a.status," +
		"a.is_finished,a.comment,a.proc_inst_create_time,a.create_time,a.finished_time").
		Joins("JOIN proc_def b ON a.proc_id = b.id").
		Offset(offset).
		Limit(limit)

	err := db.Scan(&res).Error
	return res, err
}

func (g *processEngineDAO) CountTaskRecord(ctx context.Context, processInstId int) (int64, error) {
	var res int64
	procInstDb := g.db.WithContext(ctx).Model(&model.Task{}).Table("proc_task").
		Select("id, proc_id, proc_inst_id, business_id,starter,node_id,node_name,"+
			"prev_node_id,is_cosigned,batch_code,user_id,status,is_finished,comment,proc_inst_create_time,"+
			"create_time,finished_time").
		Where("proc_inst_id = ?", processInstId)
	procHistInstDb := g.db.WithContext(ctx).Model(&model.Task{}).Table("hist_proc_task").
		Select("id,proc_id, proc_inst_id,business_id,starter,node_id,node_name,"+
			"prev_node_id,is_cosigned,batch_code,user_id,status,is_finished,comment,proc_inst_create_time,"+
			"create_time,finished_time").
		Where("proc_inst_id = ?", processInstId)

	query := g.db.Raw("? UNION ALL ?", procInstDb, procHistInstDb)
	db := g.db.Table("(?) as a", query).Select("a.id,a.proc_id,b.name,a.proc_inst_id," +
		"a.business_id,a.starter,a.node_id,a.node_name,a.prev_node_id,a.is_cosigned,a.batch_code,a.user_id,a.status," +
		"a.is_finished,a.comment,a.proc_inst_create_time,a.create_time,a.finished_time").
		Joins("JOIN proc_def b ON a.proc_id = b.id")

	err := db.Count(&res).Error
	return res, err
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

func (g *processEngineDAO) SearchStartByProcessInstIds(ctx context.Context, processInstIds []int) ([]Instance, error) {
	//TODO implement me
	panic("implement me")
}

func (g *processEngineDAO) ListHistory(ctx context.Context, userId, processName string, offset, limit int) {
	//TODO implement me
	panic("implement me")
}

func (g *processEngineDAO) ListStartUser(ctx context.Context, userId, processName string, offset, limit int) (
	[]Instance, error) {
	var res []Instance
	// TODO 当前默认审批人是只有一个，不然 JOIN proc_task 会存在问题，后期需要单独抽出函数处理
	db := g.db.WithContext(ctx).Table("proc_inst as a").Select("a.id, a.proc_id, a.proc_version, " +
		"a.business_id, a.starter, a.current_node_id, a.create_time, " +
		"a.status, b.name, c.id as task_id, c.user_id, c.node_name as current_node_name").
		Joins("JOIN proc_def b ON a.proc_id = b.id").
		Joins("JOIN proc_task c ON a.id = c.proc_inst_id AND a.current_node_id = c.node_id").
		Order("a.id").
		Limit(limit).
		Offset(offset)

	if userId != "" {
		db = db.Where("c.starter = ?", userId)
	}
	if processName != "" {
		db = db.Where("name = ?", processName)
	}

	err := db.Scan(&res).Error

	fmt.Println(res)
	return res, err
}

func (g *processEngineDAO) CountStartUser(ctx context.Context, userId, processName string) (int64, error) {
	var res int64
	db := g.db.WithContext(ctx).Model(&model.Instance{}).Table("proc_inst as a").
		Joins("JOIN proc_def b ON a.proc_id = b.id").
		Joins("JOIN proc_task c ON a.id = c.proc_inst_id AND a.current_node_id = c.node_id " +
			"AND a.starter = c.starter").
		Order("a.id")

	// 根据 userId 是否为空添加条件
	if userId != "" {
		db = db.Where("c.starter = ?", userId)
	}
	if processName != "" {
		db = db.Where("name = ?", processName)
	}

	err := db.Count(&res).Error
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
