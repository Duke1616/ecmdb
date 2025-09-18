package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/Duke1616/ecmdb/internal/attribute"
	attributemocks "github.com/Duke1616/ecmdb/internal/attribute/mocks"
	"github.com/Duke1616/ecmdb/internal/resource/internal/domain"
	resourcemocks "github.com/Duke1616/ecmdb/internal/resource/mocks"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const AesKey = "1234567890"

func crypto() cryptox.Crypto[string] {
	return cryptox.NewCryptoManager[string]("V1").
		RegisterAesAlgorithm("V1", "1234567890")
}

func Test_BatchUpdate_Resources(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (Service, attribute.Service)
		input   []domain.Resource
		wantErr error
	}{
		{
			name: "批量修改资源成功",
			mock: func(ctrl *gomock.Controller) (Service, attribute.Service) {
				attrSvc := attributemocks.NewMockService(ctrl)
				attrSvc.EXPECT().
					SearchAttributeFieldsBySecure(gomock.Any(), []string{"host"}).
					Return(map[string][]string{
						"host": {"password", "backend"},
					}, nil)

				svc := resourcemocks.NewMockService(ctrl)
				svc.EXPECT().BatchUpdateResources(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, resources []domain.Resource) (int64, error) {
						if len(resources) != 1 {
							return 0, fmt.Errorf("期望1个资源，得到%d个", len(resources))
						}

						expected := map[string]string{
							"password": "123456",
							"backend":  "mysql",
						}
						fmt.Println()
						if err := verifyEncryptedResource(t, resources[0], expected); err != nil {
							return 0, err
						}
						return 1, nil
					})

				return svc, attrSvc
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
			mock: func(ctrl *gomock.Controller) (Service, attribute.Service) {
				attrSvc := attributemocks.NewMockService(ctrl)
				attrSvc.EXPECT().
					SearchAttributeFieldsBySecure(gomock.Any(), []string{"host"}).
					Return(nil, fmt.Errorf("attr 查询错误"))

				svc := resourcemocks.NewMockService(ctrl)
				return svc, attrSvc
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

			svc, attrSvc := tc.mock(ctrl)
			crypto := cryptox.NewCryptoManager[string]("V1").
				RegisterAesAlgorithm("V1", "1234567890")
			encryptedSvc := NewEncryptedResourceService(svc, attrSvc, crypto)

			_, err := encryptedSvc.BatchUpdateResources(context.Background(), tc.input)

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
