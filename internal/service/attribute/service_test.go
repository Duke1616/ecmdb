package service

import (
	"context"
	"testing"

	"github.com/Duke1616/ecmdb/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_CreateAttribute(t *testing.T) {
	t.Parallel()

	t.Run("missing required fields", func(t *testing.T) {
		t.Parallel()

		repo := &stubAttributeRepository{}
		groupRepo := &stubAttributeGroupRepository{}
		svc := NewService(repo, groupRepo, noopSecureProducer{}, noopDeleteProducer{})

		_, err := svc.CreateAttribute(context.Background(), domain.Attribute{
			FieldUid:  "password",
			FieldName: "密码",
			FieldType: "string",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "group_id 不能为空")
		assert.Contains(t, err.Error(), "model_uid 不能为空")
		assert.False(t, repo.createCalled)
	})

	t.Run("group not found", func(t *testing.T) {
		t.Parallel()

		repo := &stubAttributeRepository{}
		groupRepo := &stubAttributeGroupRepository{}
		svc := NewService(repo, groupRepo, noopSecureProducer{}, noopDeleteProducer{})

		_, err := svc.CreateAttribute(context.Background(), domain.Attribute{
			GroupId:   11,
			ModelUid:  "host",
			FieldUid:  "password",
			FieldName: "密码",
			FieldType: "string",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "属性分组不存在")
		assert.False(t, repo.createCalled)
	})

	t.Run("group model mismatch", func(t *testing.T) {
		t.Parallel()

		repo := &stubAttributeRepository{}
		groupRepo := &stubAttributeGroupRepository{
			groupsByID: map[int64]domain.AttributeGroup{
				11: {ID: 11, ModelUid: "network"},
			},
		}
		svc := NewService(repo, groupRepo, noopSecureProducer{}, noopDeleteProducer{})

		_, err := svc.CreateAttribute(context.Background(), domain.Attribute{
			GroupId:   11,
			ModelUid:  "host",
			FieldUid:  "password",
			FieldName: "密码",
			FieldType: "string",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "属性分组不属于当前模型")
		assert.False(t, repo.createCalled)
	})

	t.Run("create with validated payload", func(t *testing.T) {
		t.Parallel()

		repo := &stubAttributeRepository{
			maxSortKey: 2000,
			createID:   99,
		}
		groupRepo := &stubAttributeGroupRepository{
			groupsByID: map[int64]domain.AttributeGroup{
				11: {ID: 11, ModelUid: "host"},
			},
		}
		svc := NewService(repo, groupRepo, noopSecureProducer{}, noopDeleteProducer{})

		id, err := svc.CreateAttribute(context.Background(), domain.Attribute{
			GroupId:   11,
			ModelUid:  "host",
			FieldUid:  "password",
			FieldName: "密码",
			FieldType: "string",
		})

		require.NoError(t, err)
		assert.Equal(t, int64(99), id)
		assert.True(t, repo.createCalled)
		assert.Equal(t, int64(3000), repo.created.SortKey)
		assert.Equal(t, int64(11), repo.maxSortKeyGroupID)
	})
}

func TestService_BatchCreateAttribute(t *testing.T) {
	t.Parallel()

	t.Run("validate group ownership", func(t *testing.T) {
		t.Parallel()

		repo := &stubAttributeRepository{}
		groupRepo := &stubAttributeGroupRepository{
			groupsByID: map[int64]domain.AttributeGroup{
				11: {ID: 11, ModelUid: "host"},
			},
		}
		svc := NewService(repo, groupRepo, noopSecureProducer{}, noopDeleteProducer{})

		err := svc.BatchCreateAttribute(context.Background(), []domain.Attribute{
			{
				GroupId:   11,
				ModelUid:  "network",
				FieldUid:  "ip",
				FieldName: "IP",
				FieldType: "string",
			},
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "属性分组不属于当前模型")
		assert.False(t, repo.batchCreateCalled)
	})

	t.Run("batch create success", func(t *testing.T) {
		t.Parallel()

		repo := &stubAttributeRepository{}
		groupRepo := &stubAttributeGroupRepository{
			groupsByID: map[int64]domain.AttributeGroup{
				11: {ID: 11, ModelUid: "host"},
				12: {ID: 12, ModelUid: "host"},
			},
		}
		svc := NewService(repo, groupRepo, noopSecureProducer{}, noopDeleteProducer{})

		err := svc.BatchCreateAttribute(context.Background(), []domain.Attribute{
			{
				GroupId:   11,
				ModelUid:  "host",
				FieldUid:  "ip",
				FieldName: "IP",
				FieldType: "string",
			},
			{
				GroupId:   12,
				ModelUid:  "host",
				FieldUid:  "password",
				FieldName: "密码",
				FieldType: "string",
			},
		})

		require.NoError(t, err)
		assert.True(t, repo.batchCreateCalled)
		assert.Len(t, repo.batchCreated, 2)
	})
}

type stubAttributeRepository struct {
	maxSortKey        int64
	maxSortKeyGroupID int64
	createID          int64
	createCalled      bool
	created           domain.Attribute
	batchCreateCalled bool
	batchCreated      []domain.Attribute
}

func (s *stubAttributeRepository) CreateAttribute(_ context.Context, req domain.Attribute) (int64, error) {
	s.createCalled = true
	s.created = req
	return s.createID, nil
}

func (s *stubAttributeRepository) BatchCreateAttribute(_ context.Context, attrs []domain.Attribute) error {
	s.batchCreateCalled = true
	s.batchCreated = attrs
	return nil
}

func (s *stubAttributeRepository) SearchAttributeFieldsByModelUid(context.Context, string) ([]string, error) {
	return nil, nil
}

func (s *stubAttributeRepository) SearchAttributeFieldsBySecure(context.Context, []string) (map[string][]string, error) {
	return nil, nil
}

func (s *stubAttributeRepository) ListAttributes(context.Context, string) ([]domain.Attribute, error) {
	return nil, nil
}

func (s *stubAttributeRepository) Total(context.Context, string) (int64, error) {
	return 0, nil
}

func (s *stubAttributeRepository) DeleteAttribute(context.Context, int64) (int64, error) {
	return 0, nil
}

func (s *stubAttributeRepository) CustomAttributeFieldColumns(context.Context, string, []string) (int64, error) {
	return 0, nil
}

func (s *stubAttributeRepository) CustomAttributeFieldColumnsReverse(context.Context, string, []string) (int64, error) {
	return 0, nil
}

func (s *stubAttributeRepository) ListAttributePipeline(context.Context, string) ([]domain.AttributePipeline, error) {
	return nil, nil
}

func (s *stubAttributeRepository) UpdateAttribute(context.Context, domain.Attribute) (int64, error) {
	return 0, nil
}

func (s *stubAttributeRepository) DetailAttribute(context.Context, int64) (domain.Attribute, error) {
	return domain.Attribute{}, nil
}

func (s *stubAttributeRepository) DeleteByGroupId(context.Context, int64) (int64, error) {
	return 0, nil
}

func (s *stubAttributeRepository) ListByGroupID(context.Context, int64) ([]domain.Attribute, error) {
	return nil, nil
}

func (s *stubAttributeRepository) GetMaxSortKeyByGroupID(_ context.Context, groupId int64) (int64, error) {
	s.maxSortKeyGroupID = groupId
	return s.maxSortKey, nil
}

func (s *stubAttributeRepository) UpdateSort(context.Context, int64, int64, int64) error {
	return nil
}

func (s *stubAttributeRepository) BatchUpdateSortKey(context.Context, []domain.AttributeSortItem) error {
	return nil
}

type stubAttributeGroupRepository struct {
	groupsByID map[int64]domain.AttributeGroup
}

func (s *stubAttributeGroupRepository) CreateAttributeGroup(context.Context, domain.AttributeGroup) (int64, error) {
	return 0, nil
}

func (s *stubAttributeGroupRepository) BatchCreateAttributeGroup(context.Context, []domain.AttributeGroup) ([]domain.AttributeGroup, error) {
	return nil, nil
}

func (s *stubAttributeGroupRepository) ListAttributeGroup(context.Context, string) ([]domain.AttributeGroup, error) {
	return nil, nil
}

func (s *stubAttributeGroupRepository) ListAttributeGroupByIds(_ context.Context, ids []int64) ([]domain.AttributeGroup, error) {
	res := make([]domain.AttributeGroup, 0, len(ids))
	for _, id := range ids {
		group, ok := s.groupsByID[id]
		if ok {
			res = append(res, group)
		}
	}
	return res, nil
}

func (s *stubAttributeGroupRepository) DeleteAttributeGroup(context.Context, int64) (int64, error) {
	return 0, nil
}

func (s *stubAttributeGroupRepository) RenameAttributeGroup(context.Context, int64, string) (int64, error) {
	return 0, nil
}

func (s *stubAttributeGroupRepository) GetMaxSortKeyByModuleUid(context.Context, string) (int64, error) {
	return 0, nil
}

func (s *stubAttributeGroupRepository) UpdateSort(context.Context, int64, int64) error {
	return nil
}

func (s *stubAttributeGroupRepository) BatchUpdateSort(context.Context, []domain.AttributeGroupSortItem) error {
	return nil
}

type noopSecureProducer struct{}

func (noopSecureProducer) Produce(context.Context, domain.FieldSecureAttrChange) error {
	return nil
}

type noopDeleteProducer struct{}

func (noopDeleteProducer) Produce(context.Context, domain.FieldDelete) error {
	return nil
}
