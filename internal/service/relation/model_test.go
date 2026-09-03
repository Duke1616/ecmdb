package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/Duke1616/ecmdb/internal/errs"
)

func TestRelationMappingErrorIsBusinessError(t *testing.T) {
	err := relationMappingError("many_to_one")

	var errCode errs.ErrorCode
	if !errors.As(err, &errCode) {
		t.Fatalf("relationMappingError() should be errs.ErrorCode")
	}
	if errCode.GetCode() != errs.RelationMappingConstraint.Code {
		t.Fatalf("code = %d, want %d", errCode.GetCode(), errs.RelationMappingConstraint.Code)
	}
}

func TestResourceLabelUsesName(t *testing.T) {
	svc := &resourceService{
		resourceRepo: fakeResourceNameRepository{
			resource: domain.Resource{
				ID:       1222,
				ModelUID: "printer",
				Name:     "办公室打印机",
			},
		},
	}

	display := svc.resourceDisplay(context.Background(), "printer", 1222)
	want := "「办公室打印机」（模型：printer，ID：1222）"
	if display != want {
		t.Fatalf("display = %s, want %s", display, want)
	}
}

func TestResourceLabelFallbacksToID(t *testing.T) {
	svc := &resourceService{}

	display := svc.resourceDisplay(context.Background(), "printer", 1222)
	want := "（模型：printer，ID：1222）"
	if display != want {
		t.Fatalf("display = %s, want %s", display, want)
	}
}

type fakeResourceNameRepository struct {
	resource domain.Resource
	err      error
}

func (f fakeResourceNameRepository) FindResourceById(ctx context.Context, fields []string, id int64) (domain.Resource, error) {
	return f.resource, f.err
}
