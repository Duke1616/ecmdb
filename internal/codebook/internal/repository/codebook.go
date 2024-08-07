package repository

import (
	"context"
	"github.com/Duke1616/ecmdb/internal/codebook/internal/domain"
	"github.com/Duke1616/ecmdb/internal/codebook/internal/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"github.com/google/uuid"
)

type CodebookRepository interface {
	CreateCodebook(ctx context.Context, req domain.Codebook) (int64, error)
	DetailCodebook(ctx context.Context, id int64) (domain.Codebook, error)
	ListCodebook(ctx context.Context, offset, limit int64) ([]domain.Codebook, error)
	Total(ctx context.Context) (int64, error)
	UpdateCodebook(ctx context.Context, req domain.Codebook) (int64, error)
	DeleteCodebook(ctx context.Context, id int64) (int64, error)
	FindBySecret(ctx context.Context, identifier string, secret string) (domain.Codebook, error)
	FindByUid(ctx context.Context, identifier string) (domain.Codebook, error)
}

func NewCodebookRepository(dao dao.CodebookDAO) CodebookRepository {
	return &codebookRepository{
		dao: dao,
	}
}

type codebookRepository struct {
	dao dao.CodebookDAO
}

func (repo *codebookRepository) FindByUid(ctx context.Context, identifier string) (domain.Codebook, error) {
	c, err := repo.dao.FindByUid(ctx, identifier)
	return repo.toDomain(c), err
}

func (repo *codebookRepository) CreateCodebook(ctx context.Context, req domain.Codebook) (int64, error) {
	return repo.dao.CreateCodebook(ctx, repo.toEntity(req))
}

func (repo *codebookRepository) DetailCodebook(ctx context.Context, id int64) (domain.Codebook, error) {
	t, err := repo.dao.DetailCodebook(ctx, id)
	return repo.toDomain(t), err
}

func (repo *codebookRepository) ListCodebook(ctx context.Context, offset, limit int64) ([]domain.Codebook, error) {
	ts, err := repo.dao.ListCodebook(ctx, offset, limit)
	return slice.Map(ts, func(idx int, src dao.Codebook) domain.Codebook {
		return repo.toDomain(src)
	}), err
}

func (repo *codebookRepository) UpdateCodebook(ctx context.Context, req domain.Codebook) (int64, error) {
	return repo.dao.UpdateCodebook(ctx, repo.toEntity(req))
}

func (repo *codebookRepository) DeleteCodebook(ctx context.Context, id int64) (int64, error) {
	return repo.dao.DeleteCodebook(ctx, id)
}

func (repo *codebookRepository) FindBySecret(ctx context.Context, identifier string, secret string) (domain.Codebook, error) {
	c, err := repo.dao.FindBySecret(ctx, identifier, secret)
	return repo.toDomain(c), err
}

func (repo *codebookRepository) Total(ctx context.Context) (int64, error) {
	return repo.dao.Count(ctx)
}

func (repo *codebookRepository) toEntity(req domain.Codebook) dao.Codebook {
	return dao.Codebook{
		Id:         req.Id,
		Name:       req.Name,
		Code:       req.Code,
		Language:   req.Language,
		Secret:     uuid.NewString(),
		Identifier: req.Identifier,
	}
}

func (repo *codebookRepository) toDomain(req dao.Codebook) domain.Codebook {
	return domain.Codebook{
		Id:         req.Id,
		Name:       req.Name,
		Code:       req.Code,
		Language:   req.Language,
		Secret:     req.Secret,
		Identifier: req.Identifier,
	}
}
