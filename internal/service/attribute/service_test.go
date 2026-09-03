package service

import (
	"context"
	"testing"

	"github.com/Duke1616/ecmdb/internal/domain"
	attributemocks "github.com/Duke1616/ecmdb/internal/service/attribute/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_CreateAttribute(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (*attributemocks.MockAttributeRepository, *attributemocks.MockAttributeGroupRepository)
		req     domain.Attribute
		wantID  int64
		wantErr string
	}{
		{
			name: "missing required fields",
			mock: func(ctrl *gomock.Controller) (*attributemocks.MockAttributeRepository, *attributemocks.MockAttributeGroupRepository) {
				return attributemocks.NewMockAttributeRepository(ctrl), attributemocks.NewMockAttributeGroupRepository(ctrl)
			},
			req: domain.Attribute{
				FieldUid:  "password",
				FieldName: "密码",
				FieldType: "string",
			},
			wantErr: "group_id 不能为空",
		},
		{
			name: "group not found",
			mock: func(ctrl *gomock.Controller) (*attributemocks.MockAttributeRepository, *attributemocks.MockAttributeGroupRepository) {
				repo := attributemocks.NewMockAttributeRepository(ctrl)
				groupRepo := attributemocks.NewMockAttributeGroupRepository(ctrl)
				groupRepo.EXPECT().ListAttributeGroupByIds(gomock.Any(), []int64{11}).
					Return([]domain.AttributeGroup{}, nil)
				return repo, groupRepo
			},
			req: domain.Attribute{
				GroupId:   11,
				ModelUid:  "host",
				FieldUid:  "password",
				FieldName: "密码",
				FieldType: "string",
			},
			wantErr: "属性分组不存在",
		},
		{
			name: "group model mismatch",
			mock: func(ctrl *gomock.Controller) (*attributemocks.MockAttributeRepository, *attributemocks.MockAttributeGroupRepository) {
				repo := attributemocks.NewMockAttributeRepository(ctrl)
				groupRepo := attributemocks.NewMockAttributeGroupRepository(ctrl)
				groupRepo.EXPECT().ListAttributeGroupByIds(gomock.Any(), []int64{11}).
					Return([]domain.AttributeGroup{
						{ID: 11, ModelUid: "network"},
					}, nil)
				return repo, groupRepo
			},
			req: domain.Attribute{
				GroupId:   11,
				ModelUid:  "host",
				FieldUid:  "password",
				FieldName: "密码",
				FieldType: "string",
			},
			wantErr: "属性分组不属于当前模型",
		},
		{
			name: "create with validated payload",
			mock: func(ctrl *gomock.Controller) (*attributemocks.MockAttributeRepository, *attributemocks.MockAttributeGroupRepository) {
				repo := attributemocks.NewMockAttributeRepository(ctrl)
				groupRepo := attributemocks.NewMockAttributeGroupRepository(ctrl)
				groupRepo.EXPECT().ListAttributeGroupByIds(gomock.Any(), []int64{11}).
					Return([]domain.AttributeGroup{
						{ID: 11, ModelUid: "host"},
					}, nil)
				repo.EXPECT().GetMaxSortKeyByGroupID(gomock.Any(), int64(11)).
					Return(int64(2000), nil)
				repo.EXPECT().CreateAttribute(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, attr domain.Attribute) (int64, error) {
						assert.Equal(t, int64(3000), attr.SortKey)
						assert.Equal(t, "password", attr.FieldUid)
						return 99, nil
					})
				return repo, groupRepo
			},
			req: domain.Attribute{
				GroupId:   11,
				ModelUid:  "host",
				FieldUid:  "password",
				FieldName: "密码",
				FieldType: "string",
			},
			wantID: 99,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, groupRepo := tc.mock(ctrl)
			secureProducer := attributemocks.NewMockFieldSecureAttrChangeEventProducer(ctrl)
			deleteProducer := attributemocks.NewMockIFieldDeleteEventProducer(ctrl)

			svc := NewService(repo, groupRepo, secureProducer, deleteProducer)
			id, err := svc.CreateAttribute(context.Background(), tc.req)

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantID, id)
			}
		})
	}
}

func TestService_BatchCreateAttribute(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (*attributemocks.MockAttributeRepository, *attributemocks.MockAttributeGroupRepository)
		attrs   []domain.Attribute
		wantErr string
	}{
		{
			name: "validate group ownership",
			mock: func(ctrl *gomock.Controller) (*attributemocks.MockAttributeRepository, *attributemocks.MockAttributeGroupRepository) {
				repo := attributemocks.NewMockAttributeRepository(ctrl)
				groupRepo := attributemocks.NewMockAttributeGroupRepository(ctrl)
				groupRepo.EXPECT().ListAttributeGroupByIds(gomock.Any(), []int64{11}).
					Return([]domain.AttributeGroup{
						{ID: 11, ModelUid: "host"},
					}, nil)
				return repo, groupRepo
			},
			attrs: []domain.Attribute{
				{
					GroupId:   11,
					ModelUid:  "network",
					FieldUid:  "ip",
					FieldName: "IP",
					FieldType: "string",
				},
			},
			wantErr: "属性分组不属于当前模型",
		},
		{
			name: "batch create success",
			mock: func(ctrl *gomock.Controller) (*attributemocks.MockAttributeRepository, *attributemocks.MockAttributeGroupRepository) {
				repo := attributemocks.NewMockAttributeRepository(ctrl)
				groupRepo := attributemocks.NewMockAttributeGroupRepository(ctrl)
				groupRepo.EXPECT().ListAttributeGroupByIds(gomock.Any(), gomock.Any()).
					Return([]domain.AttributeGroup{
						{ID: 11, ModelUid: "host"},
						{ID: 12, ModelUid: "host"},
					}, nil)
				repo.EXPECT().BatchCreateAttribute(gomock.Any(), gomock.Len(2)).
					Return(nil)
				return repo, groupRepo
			},
			attrs: []domain.Attribute{
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
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo, groupRepo := tc.mock(ctrl)
			secureProducer := attributemocks.NewMockFieldSecureAttrChangeEventProducer(ctrl)
			deleteProducer := attributemocks.NewMockIFieldDeleteEventProducer(ctrl)

			svc := NewService(repo, groupRepo, secureProducer, deleteProducer)
			err := svc.BatchCreateAttribute(context.Background(), tc.attrs)

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
