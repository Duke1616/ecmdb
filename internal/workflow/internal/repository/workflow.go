package repository

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/workflow/internal/domain"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type WorkflowRepository interface {
	// Create 创建流程定义
	Create(ctx context.Context, req domain.Workflow) (int64, error)
	// List 分页查询流程定义列表
	List(ctx context.Context, offset, limit int64) ([]domain.Workflow, error)
	// Total 统计流程定义总数
	Total(ctx context.Context) (int64, error)
	// Update 更新流程定义
	Update(ctx context.Context, req domain.Workflow) (int64, error)
	// UpdateProcessId 绑定流程引擎ID
	UpdateProcessId(ctx context.Context, id int64, processId int) error
	// Delete 删除流程定义
	Delete(ctx context.Context, id int64) (int64, error)
	// Find 根据ID查询流程定义
	Find(ctx context.Context, id int64) (domain.Workflow, error)
	// FindByKeyword 根据关键字搜索流程
	FindByKeyword(ctx context.Context, keyword string, offset, limit int64) ([]domain.Workflow, error)
	// CountByKeyword 根据关键字统计流程总数
	CountByKeyword(ctx context.Context, keyword string) (int64, error)
	// CreateSnapshot 创建流程快照
	CreateSnapshot(ctx context.Context, workflow domain.Workflow, processID, processVersion int) error
	// FindSnapshot 查找流程快照
	FindSnapshot(ctx context.Context, processID, processVersion int) (domain.Workflow, error)
}

func NewWorkflowRepository(dao dao.WorkflowDAO, snapshotDao dao.SnapshotDAO) WorkflowRepository {
	return &workflowRepository{
		dao:         dao,
		snapshotDao: snapshotDao,
	}
}

type workflowRepository struct {
	dao         dao.WorkflowDAO
	snapshotDao dao.SnapshotDAO
}

func (repo *workflowRepository) Create(ctx context.Context, req domain.Workflow) (int64, error) {
	return repo.dao.Create(ctx, repo.toEntity(req))
}

func (repo *workflowRepository) List(ctx context.Context, offset, limit int64) ([]domain.Workflow, error) {
	ws, err := repo.dao.List(ctx, offset, limit)
	return slice.Map(ws, func(idx int, src dao.Workflow) domain.Workflow {
		return repo.toDomain(src)
	}), err
}

func (repo *workflowRepository) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func (repo *workflowRepository) Find(ctx context.Context, id int64) (domain.Workflow, error) {
	w, err := repo.dao.Find(ctx, id)
	return repo.toDomain(w), err
}

func (repo *workflowRepository) Update(ctx context.Context, req domain.Workflow) (int64, error) {
	return repo.dao.Update(ctx, repo.toEntity(req))
}

func (repo *workflowRepository) Delete(ctx context.Context, id int64) (int64, error) {
	return repo.dao.Delete(ctx, id)
}

func (repo *workflowRepository) UpdateProcessId(ctx context.Context, id int64, processId int) error {
	return repo.dao.UpdateProcessId(ctx, id, processId)
}

func (repo *workflowRepository) toEntity(req domain.Workflow) dao.Workflow {
	return dao.Workflow{
		Id:           req.Id,
		TemplateId:   req.TemplateId,
		Name:         req.Name,
		Icon:         req.Icon,
		Owner:        req.Owner,
		Desc:         req.Desc,
		NotifyMethod: req.NotifyMethod.ToUint8(),
		IsNotify:     req.IsNotify,
		FlowData: dao.LogicFlow{
			Edges: req.FlowData.Edges,
			Nodes: req.FlowData.Nodes,
		},
	}
}

func (repo *workflowRepository) FindByKeyword(ctx context.Context, keyword string, offset, limit int64) ([]domain.Workflow, error) {
	ws, err := repo.dao.FindByKeyword(ctx, keyword, offset, limit)
	return slice.Map(ws, func(idx int, src dao.Workflow) domain.Workflow {
		return repo.toDomain(src)
	}), err
}

func (repo *workflowRepository) CountByKeyword(ctx context.Context, keyword string) (int64, error) {
	return repo.dao.CountByKeyword(ctx, keyword)
}

func (repo *workflowRepository) CreateSnapshot(ctx context.Context, workflow domain.Workflow, processID, processVersion int) error {
	return repo.snapshotDao.Create(ctx, dao.Snapshot{
		WorkflowID:     int(workflow.Id),
		ProcessID:      processID,
		ProcessVersion: processVersion,
		Name:           workflow.Name,
		FlowData: dao.LogicFlow{
			Edges: workflow.FlowData.Edges,
			Nodes: workflow.FlowData.Nodes,
		},
	})
}

func (repo *workflowRepository) FindSnapshot(ctx context.Context, processID, processVersion int) (domain.Workflow, error) {
	s, err := repo.snapshotDao.FindByProcess(ctx, processID, processVersion)
	if err != nil {
		return domain.Workflow{}, err
	}

	return domain.Workflow{
		Id:        int64(s.WorkflowID),
		ProcessId: s.ProcessID,
		Name:      s.Name,
		FlowData: domain.LogicFlow{
			Edges: s.FlowData.Edges,
			Nodes: s.FlowData.Nodes,
		},
	}, nil
}

func (repo *workflowRepository) toDomain(req dao.Workflow) domain.Workflow {
	return domain.Workflow{
		Id:           req.Id,
		TemplateId:   req.TemplateId,
		Name:         req.Name,
		Icon:         req.Icon,
		Owner:        req.Owner,
		Desc:         req.Desc,
		ProcessId:    req.ProcessId,
		NotifyMethod: domain.NotifyMethod(req.NotifyMethod),
		IsNotify:     req.IsNotify,
		FlowData: domain.LogicFlow{
			Edges: req.FlowData.Edges,
			Nodes: req.FlowData.Nodes,
		},
	}
}
