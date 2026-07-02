package service

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/repository"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	// Create 创建模型
	Create(ctx context.Context, req domain.Model) (int64, error)

	// CreateModelWithDefaults 创建模型并初始化默认属性（原子化编排）
	CreateModelWithDefaults(ctx context.Context, req domain.Model) (int64, error)

	// List 获取模型列表、带有分页
	List(ctx context.Context, offset, limit int64) ([]domain.Model, int64, error)

	// ListAll 获取所有模型
	ListAll(ctx context.Context) ([]domain.Model, error)

	// GetByUids 根据唯一标识检索模型列表
	GetByUids(ctx context.Context, uids []string) ([]domain.Model, error)

	// GetByUid 根据唯一标识获取模型详情
	GetByUid(ctx context.Context, uid string) (domain.Model, error)

	// DeleteById 根据ID删除指定模型
	DeleteById(ctx context.Context, id int64) (int64, error)

	// DeleteByModelUid 根据唯一标识删除指定模型
	DeleteByModelUid(ctx context.Context, modelUid string) (int64, error)

	// FindModelById 根据ID 检索模型
	FindModelById(ctx context.Context, id int64) (domain.Model, error)

	// ListModelByGroupIds  获取指定组下的所有模型
	ListModelByGroupIds(ctx context.Context, mgids []int64) ([]domain.Model, error)
}

// IDefaultAttributeCreator 创建模型时初始化默认属性的能力接口
// NOTE: 接口反转设计——model 模块仅依赖自定义的窄接口，由 attribute 模块的 Service 提供实现
type IDefaultAttributeCreator interface {
	CreateDefaultAttribute(ctx context.Context, modelUid string) (int64, error)
}

// IDeleteModelDependencyChecker 多维度依赖探测接口，各子模块注册以在删除模型前实施级联阻断校验
type IDeleteModelDependencyChecker interface {
	// CheckBeforeDelete 深度校验模型是否可删除，若已被使用则返回 error
	CheckBeforeDelete(ctx context.Context, modelUid string) error
}

type service struct {
	repo        repository.ModelRepository
	checkers    []IDeleteModelDependencyChecker
	attrCreator IDefaultAttributeCreator
}

func (s *service) GetByUid(ctx context.Context, uid string) (domain.Model, error) {
	return s.repo.GetByUid(ctx, uid)
}

func (s *service) GetByUids(ctx context.Context, uids []string) ([]domain.Model, error) {
	return s.repo.GetByUids(ctx, uids)
}

func NewModelService(repo repository.ModelRepository, checkers []IDeleteModelDependencyChecker, attrCreator IDefaultAttributeCreator) Service {
	return &service{
		repo:        repo,
		checkers:    checkers,
		attrCreator: attrCreator,
	}
}

func (s *service) ListAll(ctx context.Context) ([]domain.Model, error) {
	return s.repo.ListAll(ctx)
}

func (s *service) Create(ctx context.Context, req domain.Model) (int64, error) {
	return s.repo.Create(ctx, req)
}

// CreateModelWithDefaults 创建模型并初始化默认属性
// NOTE: 创建默认属性失败时补偿回滚模型，确保不会产生「无属性的孤儿模型」
func (s *service) CreateModelWithDefaults(ctx context.Context, req domain.Model) (int64, error) {
	id, err := s.repo.Create(ctx, req)
	if err != nil {
		return 0, err
	}

	if s.attrCreator != nil {
		if _, err = s.attrCreator.CreateDefaultAttribute(ctx, req.UID); err != nil {
			// 补偿：创建默认属性失败时回滚模型
			_, _ = s.repo.DeleteByUid(ctx, req.UID)
			return 0, fmt.Errorf("创建默认属性失败，模型已回滚: %w", err)
		}
	}

	return id, nil
}

func (s *service) FindModelById(ctx context.Context, id int64) (domain.Model, error) {
	return s.repo.FindById(ctx, id)
}

func (s *service) List(ctx context.Context, offset, limit int64) ([]domain.Model, int64, error) {
	var (
		total  int64
		models []domain.Model
		eg     errgroup.Group
	)
	eg.Go(func() error {
		var err error
		models, err = s.repo.List(ctx, offset, limit)
		return err
	})
	eg.Go(func() error {
		var err error
		total, err = s.repo.Total(ctx)
		return err
	})
	if err := eg.Wait(); err != nil {
		return models, total, err
	}
	return models, total, nil
}

func (s *service) ListModelByGroupIds(ctx context.Context, mgids []int64) ([]domain.Model, error) {
	return s.repo.ListByGroupIds(ctx, mgids)
}

func (s *service) DeleteById(ctx context.Context, id int64) (int64, error) {
	m, err := s.repo.FindById(ctx, id)
	if err != nil {
		return 0, err
	}
	// 委托领域对象校验是否允许删除
	if err = m.EnsureDeletable(); err != nil {
		return 0, err
	}

	// 级联删除安全守卫：遍历所有模块探测器进行安全校验
	for _, checker := range s.checkers {
		if err = checker.CheckBeforeDelete(ctx, m.UID); err != nil {
			return 0, err
		}
	}

	return s.repo.DeleteById(ctx, id)
}

func (s *service) DeleteByModelUid(ctx context.Context, modelUid string) (int64, error) {
	models, err := s.repo.GetByUids(ctx, []string{modelUid})
	if err != nil {
		return 0, err
	}
	// 委托领域对象校验是否允许删除
	if len(models) > 0 {
		if err = models[0].EnsureDeletable(); err != nil {
			return 0, err
		}
	}

	// 级联删除安全守卫：遍历所有模块探测器进行安全校验
	for _, checker := range s.checkers {
		if err = checker.CheckBeforeDelete(ctx, modelUid); err != nil {
			return 0, err
		}
	}

	return s.repo.DeleteByUid(ctx, modelUid)
}
