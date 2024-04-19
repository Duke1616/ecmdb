package repository

import (
	"context"
	"fmt"
	"github.com/Duke1616/ecmdb/internal/relation/internal/domain"
	"github.com/Duke1616/ecmdb/internal/relation/internal/repository/dao"
	"golang.org/x/sync/errgroup"
)

type RelationResourceRepository interface {
	CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error)
	ListResourceRelation(ctx context.Context, offset, limit int64) ([]domain.ResourceRelation, error)
	ListResourceIds(ctx context.Context, modelUid string, relationType string) ([]int64, error)

	ListSrcResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, error)
	ListDstResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, error)

	ListSrcAggregated(ctx context.Context, modelUid string, id int64) ([]domain.ResourceAggregatedData, error)
	ListDstAggregated(ctx context.Context, modelUid string, id int64) ([]domain.ResourceAggregatedData, error)

	ListSrcRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error)
	ListDstRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error)
}

func NewRelationResourceRepository(dao dao.RelationResourceDAO) RelationResourceRepository {
	return &resourceRepository{
		dao: dao,
	}
}

type resourceRepository struct {
	dao dao.RelationResourceDAO
}

func (r *resourceRepository) CreateResourceRelation(ctx context.Context, req domain.ResourceRelation) (int64, error) {
	return r.dao.CreateResourceRelation(ctx, dao.ResourceRelation{
		RelationName:     req.RelationName,
		SourceResourceID: req.SourceResourceID,
		TargetResourceID: req.TargetResourceID,
	})
}

func (r *resourceRepository) ListResourceRelation(ctx context.Context, offset, limit int64) ([]domain.ResourceRelation, error) {
	resourceRelations, err := r.dao.ListResourceRelation(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	res := make([]domain.ResourceRelation, 0, len(resourceRelations))

	for _, value := range resourceRelations {
		res = append(res, r.toResourceDomain(value))
	}

	return res, nil
}

func (r *resourceRepository) TotalByModelIdentifies(ctx context.Context, modelUid string) (int64, error) {
	return r.dao.CountByModelUid(ctx, modelUid)
}

func (r *resourceRepository) ListResourceIds(ctx context.Context, modelUid string, relationType string) ([]int64, error) {
	var (
		eg     errgroup.Group
		srcids []int64
		dstids []int64
	)
	eg.Go(func() error {
		var err error
		srcids, err = r.dao.ListSrcResourceIds(ctx, modelUid, relationType)
		return err
	})

	eg.Go(func() error {
		var err error
		dstids, err = r.dao.ListDstResourceIds(ctx, modelUid, relationType)
		return err
	})

	//total = int64(len(rd.SRC) + len(rd.DST))
	//return rd, total, eg.Wait()
	fmt.Print(dstids)
	return srcids, nil
}

func (r *resourceRepository) ListSrcResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, error) {
	resourceRelations, err := r.dao.ListSrcResources(ctx, modelUid, id)
	if err != nil {
		return nil, err
	}

	res := make([]domain.ResourceRelation, 0, len(resourceRelations))

	for _, value := range resourceRelations {
		res = append(res, r.toResourceDomain(value))
	}

	return res, nil
}

func (r *resourceRepository) ListDstResources(ctx context.Context, modelUid string, id int64) ([]domain.ResourceRelation, error) {
	resourceRelations, err := r.dao.ListDstResources(ctx, modelUid, id)
	if err != nil {
		return nil, err
	}

	res := make([]domain.ResourceRelation, 0, len(resourceRelations))

	for _, value := range resourceRelations {
		res = append(res, r.toResourceDomain(value))
	}

	return res, nil
}

func (r *resourceRepository) ListSrcAggregated(ctx context.Context, modelUid string, id int64) ([]domain.ResourceAggregatedData, error) {
	rrs, err := r.dao.ListSrcAggregated(ctx, modelUid, id)
	if err != nil {
		return nil, err
	}

	var rads []domain.ResourceAggregatedData
	for _, val := range rrs {

		var rr []domain.ResourceRelation
		for _, data := range val.Data {
			rr = append(rr, domain.ResourceRelation{
				ID:               data.Id,
				SourceModelUID:   data.SourceModelUID,
				TargetModelUID:   data.TargetModelUID,
				SourceResourceID: data.SourceResourceID,
				TargetResourceID: data.TargetResourceID,
				RelationTypeUID:  data.RelationTypeUID,
				RelationName:     data.RelationName,
			})
		}

		a := domain.ResourceAggregatedData{
			RelationName: val.RelationName,
			ModelUid:     val.ModelUid,
			Count:        val.Count,
			Data:         rr,
		}

		rads = append(rads, a)

	}

	return rads, nil
}

func (r *resourceRepository) ListDstAggregated(ctx context.Context, modelUid string, id int64) ([]domain.ResourceAggregatedData, error) {
	rrs, err := r.dao.ListDstAggregated(ctx, modelUid, id)
	if err != nil {
		return nil, err
	}

	var rads []domain.ResourceAggregatedData
	for _, val := range rrs {

		var rr []domain.ResourceRelation
		for _, data := range val.Data {
			rr = append(rr, domain.ResourceRelation{
				ID:               data.Id,
				SourceModelUID:   data.SourceModelUID,
				TargetModelUID:   data.TargetModelUID,
				SourceResourceID: data.SourceResourceID,
				TargetResourceID: data.TargetResourceID,
				RelationTypeUID:  data.RelationTypeUID,
				RelationName:     data.RelationName,
			})
		}

		a := domain.ResourceAggregatedData{
			RelationName: val.RelationName,
			ModelUid:     val.ModelUid,
			Count:        val.Count,
			Data:         rr,
		}

		rads = append(rads, a)

	}

	return rads, nil
}

func (r *resourceRepository) ListSrcRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error) {
	return r.dao.ListSrcRelated(ctx, modelUid, relationName, id)
}

func (r *resourceRepository) ListDstRelated(ctx context.Context, modelUid, relationName string, id int64) ([]int64, error) {
	return r.dao.ListDstRelated(ctx, modelUid, relationName, id)
}

func (r *resourceRepository) toResourceDomain(resourceDao dao.ResourceRelation) domain.ResourceRelation {
	return domain.ResourceRelation{
		ID:               resourceDao.Id,
		SourceModelUID:   resourceDao.SourceModelUID,
		TargetModelUID:   resourceDao.TargetModelUID,
		SourceResourceID: resourceDao.SourceResourceID,
		TargetResourceID: resourceDao.TargetResourceID,
		RelationTypeUID:  resourceDao.RelationTypeUID,
		RelationName:     resourceDao.RelationName,
	}
}
