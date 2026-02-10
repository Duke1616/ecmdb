package repository

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/workflow/internal/domain"
	"github.com/Duke1616/ecmdb/internal/workflow/internal/repository/dao"
)

type NotifyBindingRepository interface {
	Create(ctx context.Context, n domain.NotifyBinding) (int64, error)
	// Update 更新
	Update(ctx context.Context, n domain.NotifyBinding) (int64, error)
	// Delete 删除
	Delete(ctx context.Context, id int64) (int64, error)
	// List 查询流程下的所有绑定
	List(ctx context.Context, workflowId int64) ([]domain.NotifyBinding, error)
	// GetEffective 获取生效的配置 (含默认兜底逻辑)
	GetEffective(ctx context.Context, workflowId int64, notifyType domain.NotifyType, channel string) (domain.NotifyBinding, error)
}

type notifyBindingRepository struct {
	dao dao.NotifyBindingDAO
}

func NewNotifyBindingRepository(dao dao.NotifyBindingDAO) NotifyBindingRepository {
	return &notifyBindingRepository{
		dao: dao,
	}
}

func (repo *notifyBindingRepository) Create(ctx context.Context, n domain.NotifyBinding) (int64, error) {
	return repo.dao.Create(ctx, repo.toEntity(n))
}

func (repo *notifyBindingRepository) Update(ctx context.Context, n domain.NotifyBinding) (int64, error) {
	return repo.dao.Update(ctx, repo.toEntity(n))
}

func (repo *notifyBindingRepository) Delete(ctx context.Context, id int64) (int64, error) {
	return repo.dao.Delete(ctx, id)
}

func (repo *notifyBindingRepository) List(ctx context.Context, workflowId int64) ([]domain.NotifyBinding, error) {
	entities, err := repo.dao.Find(ctx, workflowId)
	if err != nil {
		return nil, err
	}
	return repo.toDomains(entities), nil
}

func (repo *notifyBindingRepository) GetEffective(ctx context.Context, workflowId int64, notifyType domain.NotifyType, channel string) (domain.NotifyBinding, error) {
	entity, err := repo.dao.FindBinding(ctx, workflowId, dao.NotifyType(notifyType), channel)
	if err != nil {
		return domain.NotifyBinding{}, err
	}
	return repo.toDomain(entity), nil
}

func (repo *notifyBindingRepository) toEntity(n domain.NotifyBinding) dao.NotifyBinding {
	return dao.NotifyBinding{
		Id:         n.Id,
		WorkflowId: n.WorkflowId,
		NotifyType: dao.NotifyType(n.NotifyType),
		Channel:    n.Channel,
		TemplateId: n.TemplateId,
	}
}

func (repo *notifyBindingRepository) toDomain(n dao.NotifyBinding) domain.NotifyBinding {
	return domain.NotifyBinding{
		Id:         n.Id,
		WorkflowId: n.WorkflowId,
		NotifyType: domain.NotifyType(n.NotifyType),
		Channel:    n.Channel,
		TemplateId: n.TemplateId,
		Ctime:      n.Ctime,
		Utime:      n.Utime,
	}
}

func (repo *notifyBindingRepository) toDomains(ns []dao.NotifyBinding) []domain.NotifyBinding {
	result := make([]domain.NotifyBinding, len(ns))
	for i, n := range ns {
		result[i] = repo.toDomain(n)
	}
	return result
}
