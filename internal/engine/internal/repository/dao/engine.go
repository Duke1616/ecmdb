package dao

import (
	"context"
	"encoding/json"
	"errors"
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
	// ForceUpdateIsFinishedByPreNodeId 强制清理指定节点下的所有任务（包括已完成的）
	// 用于并行网关驳回时清理兄弟分支，避免已完成任务逃过清理导致网关误判
	ForceUpdateIsFinishedByPreNodeId(ctx context.Context, nodeId string, status int, comment string) error
	// ForceUpdateIsFinishedByNodeId 强制清理指定节点ID的所有任务（包括已完成的）
	ForceUpdateIsFinishedByNodeId(ctx context.Context, nodeId string, status int, comment string) error
	CountReject(ctx context.Context, taskId int) (int64, error)

	ListTasksByProcInstId(ctx context.Context, processInstIds []int, starter string) ([]model.Task, error)

	GetAutomationTask(ctx context.Context, currentNodeId string, processInstId int) (model.Task, error)
	GetTasksByInstUsers(ctx context.Context, processInstId int, userIds []string) ([]model.Task, error)
	GetOrderIdByVariable(ctx context.Context, processInstId int) (string, error)
	// GetProxyNodeID 获取代理转发的节点ID
	GetProxyNodeID(ctx context.Context, prevNodeID string) (model.Task, error)
	// GetProxyNodeByProcessInstId 通过流程实例ID获取 proxy 节点
	GetProxyNodeByProcessInstId(ctx context.Context, processInstId int) (model.Task, error)
	// DeleteProxyNodeByNodeId 删除指定 proxy 节点任务记录
	DeleteProxyNodeByNodeId(ctx context.Context, nodeId string) error
	// UpdateTaskPrevNodeID 修改任务的上级节点ID
	UpdateTaskPrevNodeID(ctx context.Context, taskId int, prevNodeId string) error

	// GetInstanceByID 获取流程实例详情 (用于获取版本号)
	GetInstanceByID(ctx context.Context, processInstId int) (Instance, error)
	// GetProcessDefineByVersion 获取指定版本的流程定义 (包含历史版本)
	GetProcessDefineByVersion(ctx context.Context, processID, version int) (model.Process, error)
	// GetLatestProcessVersion 获取流程的最新版本号
	GetLatestProcessVersion(ctx context.Context, processID int) (int, error)
}

type processEngineDAO struct {
	db *gorm.DB
}

func (g *processEngineDAO) UpdateTaskPrevNodeID(ctx context.Context, taskId int, prevNodeId string) error {
	return g.db.WithContext(ctx).Table("proc_task").Where("id = ?", taskId).Update("prev_node_id", prevNodeId).Error
}

func (g *processEngineDAO) GetProxyNodeID(ctx context.Context, prevNodeID string) (model.Task, error) {
	var node model.Task
	// NOTE: 查找从指定网关节点（prevNodeID）出发的 proxy 节点
	// proxy 节点的特征：user_id = 'sys_auto'
	err := g.db.WithContext(ctx).Table("proc_task").First(&node,
		"prev_node_id = ? AND user_id = ?", prevNodeID, "sys_auto").Error
	return node, err
}

// GetProxyNodeByProcessInstId 通过流程实例ID获取 proxy 节点
func (g *processEngineDAO) GetProxyNodeByProcessInstId(ctx context.Context, processInstId int) (model.Task, error) {
	var node model.Task
	// NOTE: 查找当前流程实例中的 proxy 节点
	// proxy 节点的特征：user_id = 'sys_auto'
	err := g.db.WithContext(ctx).Table("proc_task").First(&node,
		"proc_inst_id = ? AND user_id = ?", processInstId, "sys_auto").Error
	return node, err
}

// DeleteProxyNodeByNodeId 删除指定 proxy 节点任务记录
func (g *processEngineDAO) DeleteProxyNodeByNodeId(ctx context.Context, nodeId string) error {
	// NOTE: 精确删除指定的 proxy 节点，防止误删同一流程实例下其他分支的 proxy 节点
	return g.db.WithContext(ctx).Table("proc_task").
		Where("node_id = ? AND user_id = ?", nodeId, "sys_auto").
		Delete(&model.Task{}).Error
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

// ForceUpdateIsFinishedByPreNodeId 强制清理指定节点下的所有任务（包括已完成的）
// 移除了 is_finished 和 status 的限制条件，用于并行网关驳回时清理所有兄弟分支
func (g *processEngineDAO) ForceUpdateIsFinishedByPreNodeId(ctx context.Context, nodeId string, status int, comment string) error {
	proTask := database.ProcTask{
		Status:       status,
		IsFinished:   1,
		Comment:      comment,
		FinishedTime: database.LTime.Now(),
	}

	// 不限制 is_finished 和 status，强制更新所有状态的任务
	return g.db.WithContext(ctx).
		Where("prev_node_id = ?", nodeId).
		Updates(proTask).Error
}

// ForceUpdateIsFinishedByNodeId 强制清理指定节点ID的所有任务（包括已完成的）
func (g *processEngineDAO) ForceUpdateIsFinishedByNodeId(ctx context.Context, nodeId string, status int, comment string) error {
	proTask := database.ProcTask{
		Status:       status,
		IsFinished:   1,
		Comment:      comment,
		FinishedTime: database.LTime.Now(),
	}

	// 根据 node_id 清理，不限制 is_finished 和 status
	return g.db.WithContext(ctx).
		Where("node_id = ?", nodeId).
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
		Select("task_id,proc_id, proc_inst_id,business_id,starter,node_id,node_name,"+
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
		Select("task_id,proc_id, proc_inst_id,business_id,starter,node_id,node_name,"+
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

func (g *processEngineDAO) GetInstanceByID(ctx context.Context, processInstId int) (Instance, error) {
	var res Instance
	err := g.db.WithContext(ctx).Table("proc_inst").Where("id = ?", processInstId).First(&res).Error

	// 如果在活跃表中没找到，尝试去历史表中查找
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = g.db.WithContext(ctx).Table("hist_proc_inst").Where("proc_inst_id = ?", processInstId).First(&res).Error
	}

	return res, err
}

func (g *processEngineDAO) GetProcessDefineByVersion(ctx context.Context, processID, version int) (model.Process, error) {
	var resource string

	// 使用 UNION ALL 查询主表和历史表
	// 优先级：大部分冷数据在 hist 表，热数据在 proc_def 表
	// SELECT resource FROM (SELECT resource, version FROM proc_def WHERE id=? UNION ALL SELECT resource, version FROM hist_proc_def WHERE proc_id=?) as t WHERE version=? LIMIT 1
	subQuery1 := g.db.Table("proc_def").Select("resource, version").Where("id = ?", processID)
	subQuery2 := g.db.Table("hist_proc_def").Select("resource, version").Where("proc_id = ?", processID)

	err := g.db.Table("(?) as t", g.db.Raw("? UNION ALL ?", subQuery1, subQuery2)).
		Select("resource").
		Where("version = ?", version).
		Limit(1).
		Scan(&resource).Error

	if err != nil {
		return model.Process{}, err
	}

	if resource == "" {
		return model.Process{}, fmt.Errorf("definition for process_id=%d version=%d not found", processID, version)
	}

	// Parse JSON
	var process model.Process
	if err = json.Unmarshal([]byte(resource), &process); err != nil {
		return model.Process{}, err
	}
	return process, nil
}

func (g *processEngineDAO) GetLatestProcessVersion(ctx context.Context, processID int) (int, error) {
	var version int
	err := g.db.WithContext(ctx).Table("proc_def").
		Select("version").
		Where("id = ?", processID).
		Scan(&version).Error
	return version, err
}
