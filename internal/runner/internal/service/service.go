package service

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/runner/internal/domain"
	"github.com/Duke1616/ecmdb/internal/runner/internal/repository"
	"golang.org/x/sync/errgroup"
)

// Service 执行器（Runner）服务接口
// 负责处理执行器的核心业务逻辑，包括注册、信息查询、状态流转以及与任务引擎间的数据聚合
type Service interface {
	// Create 注册一个新的执行器节点
	Create(ctx context.Context, req domain.Runner) (int64, error)
	// Update 更新现有执行器的属性与配置信息
	Update(ctx context.Context, req domain.Runner) (int64, error)
	// FindById 获取单个执行器的详细配置信息
	FindById(ctx context.Context, id int64) (domain.Runner, error)
	// Delete 根据 ID 删除指定的执行器
	Delete(ctx context.Context, id int64) (int64, error)
	// List 分页获取执行器节点列表及总数
	List(ctx context.Context, offset, limit int64, keyword, runMode string) ([]domain.Runner, int64, error)
	// FindByCodebookUidAndTag 根据绑定的脚本 UID 和特定策略标签匹配指定的执行器节点
	FindByCodebookUidAndTag(ctx context.Context, codebookUid string, tag string) (domain.Runner, error)
	// ListByCodebookUid 根据脚本 UID 获取其所有具备承载能力的执行器节点（支持分页）
	ListByCodebookUid(ctx context.Context, offset, limit int64, codebookUid, keyword, runMode string) ([]domain.Runner, int64, error)
	// ListExcludeCodebookUid 获取由于未绑定特定脚本 UID，剩余支持额外添加的执行器节点（支持分页）
	ListExcludeCodebookUid(ctx context.Context, offset, limit int64, codebookUid, keyword, runMode string) ([]domain.Runner, int64, error)
	// ListByCodebookUids 根据多个脚本 UID 批量拉取能承载相关任务的执行器节点
	ListByCodebookUids(ctx context.Context, codebookUids []string) ([]domain.Runner, error)
	// ListByIds 根据一组内建的 ID 列表批量拉取执行器对象
	ListByIds(ctx context.Context, ids []int64) ([]domain.Runner, error)
	// AggregateTags 提取并聚合所有剧本关联的 Runner 标签拓扑体系（用于前端资源联动等场景）
	AggregateTags(ctx context.Context) ([]domain.RunnerTags, error)
}

type service struct {
	repo repository.RunnerRepository
}

func (s *service) ListByIds(ctx context.Context, ids []int64) ([]domain.Runner, error) {
	return s.repo.ListByIds(ctx, ids)
}

func (s *service) ListByCodebookUids(ctx context.Context, codebookUids []string) ([]domain.Runner, error) {
	return s.repo.ListByCodebookUids(ctx, codebookUids)
}

func (s *service) ListByCodebookUid(ctx context.Context, offset, limit int64, codebookUid, keyword, runMode string) ([]domain.Runner, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Runner
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListByCodebookUid(ctx, offset, limit, codebookUid, keyword, runMode)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountByCodebookUid(ctx, codebookUid, keyword, runMode)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) ListExcludeCodebookUid(ctx context.Context, offset, limit int64, codebookUid, keyword, runMode string) ([]domain.Runner, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Runner
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.ListExcludeCodebookUid(ctx, offset, limit, codebookUid, keyword, runMode)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.CountExcludeCodebookUid(ctx, codebookUid, keyword, runMode)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}

func (s *service) FindById(ctx context.Context, id int64) (domain.Runner, error) {
	return s.repo.FindById(ctx, id)
}

func (s *service) Delete(ctx context.Context, id int64) (int64, error) {
	return s.repo.Delete(ctx, id)
}

func (s *service) Update(ctx context.Context, req domain.Runner) (int64, error) {
	return s.repo.Update(ctx, req)
}

func (s *service) AggregateTags(ctx context.Context) ([]domain.RunnerTags, error) {
	return s.repo.AggregateTags(ctx)
}

func (s *service) FindByCodebookUidAndTag(ctx context.Context, codebookUid string, tag string) (domain.Runner, error) {
	return s.repo.FindByCodebookUidAndTag(ctx, codebookUid, tag)
}

func NewService(repo repository.RunnerRepository) Service {
	return &service{
		repo: repo,
	}
}
func (s *service) Create(ctx context.Context, req domain.Runner) (int64, error) {
	return s.repo.Create(ctx, req)
}

func (s *service) List(ctx context.Context, offset, limit int64, keyword, runMode string) ([]domain.Runner, int64, error) {
	var (
		eg    errgroup.Group
		ts    []domain.Runner
		total int64
	)
	eg.Go(func() error {
		var err error
		ts, err = s.repo.List(ctx, offset, limit, keyword, runMode)
		return err
	})

	eg.Go(func() error {
		var err error
		total, err = s.repo.Count(ctx, keyword, runMode)
		return err
	})
	if err := eg.Wait(); err != nil {
		return ts, total, err
	}
	return ts, total, nil
}
