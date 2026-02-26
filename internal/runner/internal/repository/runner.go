package repository

import (
	"context"

	"github.com/Duke1616/ecmdb/internal/runner/internal/domain"
	"github.com/Duke1616/ecmdb/internal/runner/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

// RunnerRepository 执行器仓储接口
// 提供面向业务领域层 (`domain.Runner`) 的持久化操作防腐层抽象与缓存屏蔽等
type RunnerRepository interface {
	// Create 将领域模型的执行器实例调度持久化落盘
	Create(ctx context.Context, req domain.Runner) (int64, error)
	// Update 将修改后的领域模型状态映射到持久化层中更新存储
	Update(ctx context.Context, req domain.Runner) (int64, error)
	// Delete 根据抽象 ID 移除指定的执行器对象数据
	Delete(ctx context.Context, id int64) (int64, error)
	// FindById 根据 ID 从底层层拉取持久化数据并转换为面向业务逻辑的领域模型
	FindById(ctx context.Context, id int64) (domain.Runner, error)
	// List 提供支持分页的基础执行器节点领域模型列表查询能力
	List(ctx context.Context, offset, limit int64, keyword, runMode string) ([]domain.Runner, error)
	// Count 返回数据存储下所映射的可用执行器节点库容大小
	Count(ctx context.Context, keyword, runMode string) (int64, error)
	// ListByCodebookUid 通过单一特定脚本的特征 UID 拉取有关联的全部执行器领域实体（分页）
	ListByCodebookUid(ctx context.Context, offset, limit int64, codebookUid, keyword, runMode string) ([]domain.Runner, error)
	// CountByCodebookUid 统计匹配特定特征 UID 环境下的总数据量
	CountByCodebookUid(ctx context.Context, codebookUid, keyword, runMode string) (int64, error)
	// ListExcludeCodebookUid 反向检索不带有特点脚本特性 UID 的可备选执行器实体（分页）
	ListExcludeCodebookUid(ctx context.Context, offset, limit int64, codebookUid, keyword, runMode string) ([]domain.Runner, error)
	// CountExcludeCodebookUid 统计未关联特定特征 UID 的执行器数量
	CountExcludeCodebookUid(ctx context.Context, codebookUid, keyword, runMode string) (int64, error)
	// ListByCodebookUids 依循脚本级 UID 的条件批量拉取关联的执行器领域模型列表
	ListByCodebookUids(ctx context.Context, codebookUids []string) ([]domain.Runner, error)
	// ListByIds 依循一组既定的独立节点 ID 批量载入业务层用执行器模型集
	ListByIds(ctx context.Context, ids []int64) ([]domain.Runner, error)
	// FindByCodebookUidAndTag 使用脚本的特定 UID 及对应的调度 Tag 拉取特定的执行器
	FindByCodebookUidAndTag(ctx context.Context, codebookUid string, tag string) (domain.Runner, error)
	// AggregateTags 从底层封装聚合出的按脚本提取的分配与触发特征模型映射链
	AggregateTags(ctx context.Context) ([]domain.RunnerTags, error)
}

func NewRunnerRepository(dao dao.RunnerDAO) RunnerRepository {
	return &runnerRepository{
		dao: dao,
	}
}

type runnerRepository struct {
	dao dao.RunnerDAO
}

func (repo *runnerRepository) ListByIds(ctx context.Context, ids []int64) ([]domain.Runner, error) {
	rs, err := repo.dao.ListByIds(ctx, ids)
	return slice.Map(rs, func(idx int, src dao.Runner) domain.Runner {
		return repo.toDomain(src)
	}), err
}

func (repo *runnerRepository) ListByCodebookUids(ctx context.Context, codebookUids []string) ([]domain.Runner, error) {
	rs, err := repo.dao.ListByCodebookUids(ctx, codebookUids)
	return slice.Map(rs, func(idx int, src dao.Runner) domain.Runner {
		return repo.toDomain(src)
	}), err
}

func (repo *runnerRepository) ListByCodebookUid(ctx context.Context, offset, limit int64, codebookUid, keyword, runMode string) ([]domain.Runner, error) {
	rs, err := repo.dao.ListByCodebookUid(ctx, offset, limit, codebookUid, keyword, runMode)
	return slice.Map(rs, func(idx int, src dao.Runner) domain.Runner {
		return repo.toDomain(src)
	}), err
}

func (repo *runnerRepository) CountByCodebookUid(ctx context.Context, codebookUid, keyword, runMode string) (int64, error) {
	return repo.dao.CountByCodebookUid(ctx, codebookUid, keyword, runMode)
}

func (repo *runnerRepository) ListExcludeCodebookUid(ctx context.Context, offset, limit int64, codebookUid, keyword, runMode string) ([]domain.Runner, error) {
	rs, err := repo.dao.ListExcludeCodebookUid(ctx, offset, limit, codebookUid, keyword, runMode)
	return slice.Map(rs, func(idx int, src dao.Runner) domain.Runner {
		return repo.toDomain(src)
	}), err
}

func (repo *runnerRepository) CountExcludeCodebookUid(ctx context.Context, codebookUid, keyword, runMode string) (int64, error) {
	return repo.dao.CountExcludeCodebookUid(ctx, codebookUid, keyword, runMode)
}

func (repo *runnerRepository) FindById(ctx context.Context, id int64) (domain.Runner, error) {
	runner, err := repo.dao.FindById(ctx, id)
	return repo.toDomain(runner), err
}

func (repo *runnerRepository) Delete(ctx context.Context, id int64) (int64, error) {
	return repo.dao.Delete(ctx, id)
}

func (repo *runnerRepository) Update(ctx context.Context, req domain.Runner) (int64, error) {
	return repo.dao.Update(ctx, repo.toEntity(req))
}

func (repo *runnerRepository) AggregateTags(ctx context.Context) ([]domain.RunnerTags, error) {
	pipeline, err := repo.dao.AggregateTags(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.RunnerTags, len(pipeline))
	for i, src := range pipeline {
		tagSet := make(map[string]string)
		for _, runner := range src.RunnerTags {
			for _, tag := range runner.Tags {
				tagSet[tag] = runner.Topic
			}
		}

		result[i] = domain.RunnerTags{
			CodebookUid:      src.CodebookUid,
			TagsMappingTopic: tagSet,
		}
	}

	return result, nil
}

func (repo *runnerRepository) FindByCodebookUidAndTag(ctx context.Context, codebookUid string, tag string) (domain.Runner, error) {
	runner, err := repo.dao.FindByCodebookUidAndTag(ctx, codebookUid, tag)
	return repo.toDomain(runner), err
}

func (repo *runnerRepository) Create(ctx context.Context, req domain.Runner) (int64, error) {
	return repo.dao.Create(ctx, repo.toEntity(req))
}

func (repo *runnerRepository) List(ctx context.Context, offset, limit int64, keyword, runMode string) ([]domain.Runner, error) {
	ws, err := repo.dao.List(ctx, offset, limit, keyword, runMode)
	return slice.Map(ws, func(idx int, src dao.Runner) domain.Runner {
		return repo.toDomain(src)
	}), err
}

func (repo *runnerRepository) Count(ctx context.Context, keyword, runMode string) (int64, error) {
	return repo.dao.Count(ctx, keyword, runMode)
}

func (repo *runnerRepository) toEntity(req domain.Runner) dao.Runner {
	r := dao.Runner{
		Id:             req.Id,
		CodebookSecret: req.CodebookSecret,
		CodebookUid:    req.CodebookUid,
		RunMode:        req.RunMode.ToString(),
		Name:           req.Name,
		Tags:           req.Tags,
		Variables: slice.Map(req.Variables, func(idx int, src domain.Variables) dao.Variables {
			return dao.Variables{
				Key:    src.Key,
				Value:  src.Value,
				Secret: src.Secret,
			}
		}),
		Desc:   req.Desc,
		Action: req.Action.ToUint8(),
	}

	// NOTE: Worker 和 Execute 字段通过 inline 平铺存储，根据 domain 指针是否存在来决定是否赋值
	if req.Worker != nil {
		r.Worker.WorkerName = req.Worker.WorkerName
		r.Worker.Topic = req.Worker.Topic
	}
	if req.Execute != nil {
		r.Execute.ServiceName = req.Execute.ServiceName
		r.Execute.Handler = req.Execute.Handler
	}

	return r
}

func (repo *runnerRepository) toDomain(req dao.Runner) domain.Runner {
	r := domain.Runner{
		Id:             req.Id,
		Name:           req.Name,
		CodebookSecret: req.CodebookSecret,
		CodebookUid:    req.CodebookUid,
		RunMode:        domain.RunMode(req.RunMode),
		Tags:           req.Tags,
		Variables: slice.Map(req.Variables, func(idx int, src dao.Variables) domain.Variables {
			return domain.Variables{
				Key:    src.Key,
				Value:  src.Value,
				Secret: src.Secret,
			}
		}),
		Desc:   req.Desc,
		Action: domain.Action(req.Action),
	}

	// 兼容历史不存在字段的情况，默认使用 Worker 模式
	if req.RunMode == "" {
		r.RunMode = domain.RunModeWorker
	}

	if req.Worker.WorkerName != "" {
		r.Worker = &domain.Worker{
			WorkerName: req.Worker.WorkerName,
			Topic:      req.Worker.Topic,
		}
	}
	if req.Execute.ServiceName != "" {
		r.Execute = &domain.Execute{
			ServiceName: req.Execute.ServiceName,
			Handler:     req.Execute.Handler,
		}
	}

	return r
}
