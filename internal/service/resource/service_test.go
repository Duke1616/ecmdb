package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/Duke1616/ecmdb/internal/domain"
	attributemocks "github.com/Duke1616/ecmdb/internal/service/attribute/mocks"
	resourcemocks "github.com/Duke1616/ecmdb/internal/service/resource/mocks"
	attribute "github.com/Duke1616/ecmdb/internal/service/attribute"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const AesKey = "1234567890123456"

func crypto() cryptox.Crypto {
	return cryptox.NewCryptoManager("V1").
		Register("V1", cryptox.MustNewAESCrypto(AesKey))
}

func Test_BatchUpdate_Resources(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (attribute.Service, *resourcemocks.MockResourceRepository)
		input   []domain.Resource
		wantErr error
	}{
		{
			name: "批量修改资源成功",
			mock: func(ctrl *gomock.Controller) (attribute.Service, *resourcemocks.MockResourceRepository) {
				attrSvc := attributemocks.NewMockService(ctrl)
				attrSvc.EXPECT().
					SearchAttributeFieldsBySecure(gomock.Any(), []string{"host"}).
					Return(map[string][]string{
						"host": {"password", "backend"},
					}, nil)

				repo := resourcemocks.NewMockResourceRepository(ctrl)
				repo.EXPECT().BatchUpdateResources(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, resources []domain.Resource) (int64, error) {
						if len(resources) != 1 {
							return 0, fmt.Errorf("期望1个资源，得到%d个", len(resources))
						}

						expected := map[string]string{
							"password": "123456",
							"backend":  "mysql",
						}

						if err := verifyEncryptedResource(t, resources[0], expected); err != nil {
							return 0, err
						}
						return 1, nil
					})

				return attrSvc, repo
			},
			input: []domain.Resource{
				{
					ID:       1,
					Name:     "Instance01",
					ModelUID: "host",
					Data: map[string]interface{}{
						"password": "123456",
						"backend":  "mysql",
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "attrSvc 查询失败",
			mock: func(ctrl *gomock.Controller) (attribute.Service, *resourcemocks.MockResourceRepository) {
				attrSvc := attributemocks.NewMockService(ctrl)
				attrSvc.EXPECT().
					SearchAttributeFieldsBySecure(gomock.Any(), []string{"host"}).
					Return(nil, fmt.Errorf("attr 查询错误"))

				repo := resourcemocks.NewMockResourceRepository(ctrl)
				return attrSvc, repo
			},
			input: []domain.Resource{
				{
					ID:       1,
					Name:     "Instance02",
					ModelUID: "host",
					Data:     map[string]interface{}{"password": "xxx"},
				},
			},
			wantErr: fmt.Errorf("attr 查询错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			attrSvc, repo := tc.mock(ctrl)
			c := crypto()
			svc := NewService(repo, attrSvc, c)

			_, err := svc.BatchUpdateResources(context.Background(), tc.input)

			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func verifyEncryptedResource(t *testing.T, resource domain.Resource, expected map[string]string) error {
	for field, plain := range expected {
		encrypted, ok := resource.Data[field].(string)
		if !ok {
			return fmt.Errorf("%s 字段不是字符串类型", field)
		}
		decrypted, err := crypto().Decrypt(encrypted)
		if err != nil {
			return fmt.Errorf("%s 解密失败: %v", field, err)
		}
		assert.Equal(t, plain, decrypted, "字段 %s 解密后不匹配", field)
	}
	return nil
}

func Test_AdminSearchStructure(t *testing.T) {
	testCases := []struct {
		name      string
		mock      func(ctrl *gomock.Controller) (attribute.Service, *resourcemocks.MockResourceRepository)
		text      string
		wantTotal int
		wantOrgs  int
		wantPers  int
		wantErr   error
	}{
		{
			name: "大盘结构聚合成功并正确区分组织与个人空间",
			mock: func(ctrl *gomock.Controller) (attribute.Service, *resourcemocks.MockResourceRepository) {
				repo := resourcemocks.NewMockResourceRepository(ctrl)
				repo.EXPECT().AdminSearchStructure(gomock.Any(), "prod").
					Return([]domain.AdminSearchModelCount{
						{TenantID: 1, ModelUID: "host", Total: 10},
						{TenantID: 1, ModelUID: "switch", Total: 5},
						{TenantID: 200, ModelUID: "host", Total: 3},
					}, nil)

				attrSvc := attributemocks.NewMockService(ctrl)
				return attrSvc, repo
			},
			text:      "prod",
			wantTotal: 18,
			wantOrgs:  1,
			wantPers:  1,
			wantErr:   nil,
		},
		{
			name: "大盘结构查询结果为空",
			mock: func(ctrl *gomock.Controller) (attribute.Service, *resourcemocks.MockResourceRepository) {
				repo := resourcemocks.NewMockResourceRepository(ctrl)
				repo.EXPECT().AdminSearchStructure(gomock.Any(), "empty").
					Return([]domain.AdminSearchModelCount{}, nil)

				attrSvc := attributemocks.NewMockService(ctrl)
				return attrSvc, repo
			},
			text:      "empty",
			wantTotal: 0,
			wantOrgs:  0,
			wantPers:  0,
			wantErr:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			attrSvc, repo := tc.mock(ctrl)
			svc := NewService(repo, attrSvc, crypto())

			res, err := svc.AdminSearchStructure(context.Background(), tc.text)
			if tc.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantTotal, res.Total)
				assert.Len(t, res.Organizations, tc.wantOrgs)
				assert.Len(t, res.Personals, tc.wantPers)
			}
		})
	}
}

func Test_SearchPagedResources(t *testing.T) {
	testCases := []struct {
		name      string
		mock      func(ctrl *gomock.Controller) (attribute.Service, *resourcemocks.MockResourceRepository)
		tenantID  int64
		modelUID  string
		text      string
		wantTotal int64
		wantCount int
		wantErr   error
	}{
		{
			name: "指定租户物理分页检索成功并精准脱敏",
			mock: func(ctrl *gomock.Controller) (attribute.Service, *resourcemocks.MockResourceRepository) {
				repo := resourcemocks.NewMockResourceRepository(ctrl)
				repo.EXPECT().SearchPagedResources(gomock.Any(), int64(1), "host", "web", int64(0), int64(10)).
					Return([]domain.Resource{
						{
							ID:       101,
							Name:     "web-prod-01",
							ModelUID: "host",
							Data:     mongox.MapStr{"ip": "192.168.1.10", "password": "plain-password"},
						},
					}, int64(1), nil)

				attrSvc := attributemocks.NewMockService(ctrl)
				attrSvc.EXPECT().SearchAttributeFieldsBySecure(gomock.Any(), []string{"host"}).
					Return(map[string][]string{
						"host": {"password"},
					}, nil)

				return attrSvc, repo
			},
			tenantID:  1,
			modelUID:  "host",
			text:      "web",
			wantTotal: 1,
			wantCount: 1,
			wantErr:   nil,
		},
		{
			name: "单租户默认上下文物理分页检索成功并精准脱敏",
			mock: func(ctrl *gomock.Controller) (attribute.Service, *resourcemocks.MockResourceRepository) {
				repo := resourcemocks.NewMockResourceRepository(ctrl)
				repo.EXPECT().SearchPagedResources(gomock.Any(), int64(0), "host", "web", int64(0), int64(10)).
					Return([]domain.Resource{
						{
							ID:       102,
							Name:     "web-01",
							ModelUID: "host",
							Data:     mongox.MapStr{"ip": "10.0.0.1", "password": "plain-password"},
						},
					}, int64(1), nil)

				attrSvc := attributemocks.NewMockService(ctrl)
				attrSvc.EXPECT().SearchAttributeFieldsBySecure(gomock.Any(), []string{"host"}).
					Return(map[string][]string{
						"host": {"password"},
					}, nil)

				return attrSvc, repo
			},
			tenantID:  0,
			modelUID:  "host",
			text:      "web",
			wantTotal: 1,
			wantCount: 1,
			wantErr:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			attrSvc, repo := tc.mock(ctrl)
			svc := NewService(repo, attrSvc, crypto())

			res, total, err := svc.SearchPagedResources(context.Background(), tc.tenantID, tc.modelUID, tc.text, 0, 10)
			if tc.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantTotal, total)
				assert.Len(t, res, tc.wantCount)
				assert.Equal(t, "", res[0].Data["password"], "敏感字段必须脱敏置空")
			}
		})
	}
}
