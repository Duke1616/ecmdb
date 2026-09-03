package service

import (
	"context"
	"testing"

	"github.com/Duke1616/ecmdb/internal/domain"
	attributemocks "github.com/Duke1616/ecmdb/internal/service/attribute/mocks"
	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/Duke1616/ecmdb/pkg/mongox"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type dummyCrypto struct{}

func (d *dummyCrypto) Encrypt(plainText string) (string, error) {
	return "ENC:V2:dummy_" + plainText, nil
}

func (d *dummyCrypto) Decrypt(encryptedText string) (string, error) {
	if encryptedText == "ENC:V2:dummy_secret" {
		return "secret", nil
	}
	return "dummy_plain", nil
}

func (d *dummyCrypto) DecryptCiphertext(encryptedText string) (string, error) {
	if encryptedText == "ENC:V2:dummy_secret" {
		return "secret", nil
	}
	return "", cryptox.ErrInvalidCiphertext
}

func TestResourceProtector_Lifecycle(t *testing.T) {
	testCases := []struct {
		name     string
		mockAttr func(ctrl *gomock.Controller) *attributemocks.MockService
		testFunc func(t *testing.T, protector IResourceProtector)
	}{
		{
			name: "成功加密敏感字段并跳过普通字段",
			mockAttr: func(ctrl *gomock.Controller) *attributemocks.MockService {
				svc := attributemocks.NewMockService(ctrl)
				svc.EXPECT().SearchAttributeFieldsBySecure(gomock.Any(), []string{"host"}).
					Return(map[string][]string{
						"host": {"password", "token"},
					}, nil)
				return svc
			},
			testFunc: func(t *testing.T, protector IResourceProtector) {
				res := domain.Resource{
					ID:       1,
					ModelUID: "host",
					Data: mongox.MapStr{
						"ip":       "192.168.1.1",
						"password": "my_password",
					},
				}
				encrypted, err := protector.Encrypt(context.Background(), res)
				assert.NoError(t, err)
				assert.Equal(t, "192.168.1.1", encrypted.Data["ip"])
				assert.Equal(t, "ENC:V2:dummy_my_password", encrypted.Data["password"])
			},
		},
		{
			name: "脱敏占位符跳过加密",
			mockAttr: func(ctrl *gomock.Controller) *attributemocks.MockService {
				svc := attributemocks.NewMockService(ctrl)
				svc.EXPECT().SearchAttributeFieldsBySecure(gomock.Any(), []string{"host"}).
					Return(map[string][]string{
						"host": {"password"},
					}, nil)
				return svc
			},
			testFunc: func(t *testing.T, protector IResourceProtector) {
				res := domain.Resource{
					ID:       1,
					ModelUID: "host",
					Data: mongox.MapStr{
						"password": cryptox.DefaultMask,
					},
				}
				encrypted, err := protector.Encrypt(context.Background(), res)
				assert.NoError(t, err)
				assert.Equal(t, cryptox.DefaultMask, encrypted.Data["password"])
			},
		},
		{
			name: "成功解密单个资产敏感字段",
			mockAttr: func(ctrl *gomock.Controller) *attributemocks.MockService {
				svc := attributemocks.NewMockService(ctrl)
				svc.EXPECT().SearchAttributeFieldsBySecure(gomock.Any(), []string{"host"}).
					Return(map[string][]string{
						"host": {"password"},
					}, nil)
				return svc
			},
			testFunc: func(t *testing.T, protector IResourceProtector) {
				res := domain.Resource{
					ID:       1,
					ModelUID: "host",
					Data: mongox.MapStr{
						"password": "ENC:V2:dummy_secret",
					},
				}
				decrypted, err := protector.Decrypt(context.Background(), res)
				assert.NoError(t, err)
				assert.Equal(t, "secret", decrypted.Data["password"])
			},
		},
		{
			name: "自动解密非当前安全属性但含有ENC前缀的历史密文字段",
			mockAttr: func(ctrl *gomock.Controller) *attributemocks.MockService {
				svc := attributemocks.NewMockService(ctrl)
				svc.EXPECT().SearchAttributeFieldsBySecure(gomock.Any(), []string{"host"}).
					Return(map[string][]string{
						"host": {},
					}, nil)
				return svc
			},
			testFunc: func(t *testing.T, protector IResourceProtector) {
				res := domain.Resource{
					ID:       1,
					ModelUID: "host",
					Data: mongox.MapStr{
						"legacy_password": "ENC:V2:dummy_secret",
						"normal_field":    "normal_val",
					},
				}
				decrypted, err := protector.Decrypt(context.Background(), res)
				assert.NoError(t, err)
				assert.Equal(t, "secret", decrypted.Data["legacy_password"])
				assert.Equal(t, "normal_val", decrypted.Data["normal_field"])
			},
		},
		{
			name: "DecryptFields针对指定字段执行批量解密",
			mockAttr: func(ctrl *gomock.Controller) *attributemocks.MockService {
				return attributemocks.NewMockService(ctrl)
			},
			testFunc: func(t *testing.T, protector IResourceProtector) {
				resources := []domain.Resource{
					{
						ID:       1,
						ModelUID: "host",
						Data: mongox.MapStr{
							"password": "ENC:V2:dummy_secret",
							"ip":       "192.168.1.1",
						},
					},
				}
				decrypted, err := protector.DecryptFields(context.Background(), resources, []string{"password"})
				assert.NoError(t, err)
				assert.Equal(t, "secret", decrypted[0].Data["password"])
				assert.Equal(t, "192.168.1.1", decrypted[0].Data["ip"])
			},
		},
		{
			name: "敏感字段安全脱敏替换",
			mockAttr: func(ctrl *gomock.Controller) *attributemocks.MockService {
				svc := attributemocks.NewMockService(ctrl)
				svc.EXPECT().SearchAttributeFieldsBySecure(gomock.Any(), []string{"host"}).
					Return(map[string][]string{
						"host": {"password", "secret_key"},
					}, nil)
				return svc
			},
			testFunc: func(t *testing.T, protector IResourceProtector) {
				res := domain.Resource{
					ID:       1,
					ModelUID: "host",
					Data: mongox.MapStr{
						"password":   "my_pass",
						"secret_key": "my_key",
						"public_ip":  "1.1.1.1",
					},
				}
				masked, err := protector.Mask(context.Background(), res)
				assert.NoError(t, err)
				assert.Equal(t, cryptox.DefaultMask, masked.Data["password"])
				assert.Equal(t, cryptox.DefaultMask, masked.Data["secret_key"])
				assert.Equal(t, "1.1.1.1", masked.Data["public_ip"])
			},
		},
		{
			name: "DecryptValue对单个值解密",
			mockAttr: func(ctrl *gomock.Controller) *attributemocks.MockService {
				return attributemocks.NewMockService(ctrl)
			},
			testFunc: func(t *testing.T, protector IResourceProtector) {
				val, err := protector.DecryptValue("ENC:V2:dummy_secret")
				assert.NoError(t, err)
				assert.Equal(t, "secret", val)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			attrSvc := tc.mockAttr(ctrl)
			protector := NewResourceProtector(attrSvc, &dummyCrypto{})

			tc.testFunc(t, protector)
		})
	}
}

func TestResourceProtector_IsMasked(t *testing.T) {
	protector := NewResourceProtector(nil, &dummyCrypto{})

	assert.True(t, protector.IsMasked(cryptox.DefaultMask))
	assert.False(t, protector.IsMasked("regular_string"))
	assert.False(t, protector.IsMasked(123))
	assert.False(t, protector.IsMasked(nil))
}
