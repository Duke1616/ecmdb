package service

import (
	"context"
	"testing"

	"github.com/Duke1616/ecmdb/pkg/cryptox"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	attributemocks "github.com/Duke1616/ecmdb/internal/service/attribute/mocks"
)

type dummyCrypto struct{}

func (d *dummyCrypto) Encrypt(plainText string) (string, error) {
	return "ENC:V2:dummy_" + plainText, nil
}

func (d *dummyCrypto) Decrypt(encryptedText string) (string, error) {
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
		modelUID string
		data     map[string]any
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
			modelUID: "host",
			data: map[string]any{
				"ip":       "192.168.1.1",
				"password": "my_password",
			},
			testFunc: func(t *testing.T, protector IResourceProtector) {
				encrypted, err := protector.EncryptResource(context.Background(), "host", map[string]any{
					"ip":       "192.168.1.1",
					"password": "my_password",
				})
				assert.NoError(t, err)
				assert.Equal(t, "192.168.1.1", encrypted["ip"])
				assert.Equal(t, "ENC:V2:dummy_my_password", encrypted["password"])
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
			modelUID: "host",
			testFunc: func(t *testing.T, protector IResourceProtector) {
				encrypted, err := protector.EncryptResource(context.Background(), "host", map[string]any{
					"password": cryptox.DefaultMask,
				})
				assert.NoError(t, err)
				assert.Equal(t, cryptox.DefaultMask, encrypted["password"])
			},
		},
		{
			name: "成功解密敏感字段",
			mockAttr: func(ctrl *gomock.Controller) *attributemocks.MockService {
				svc := attributemocks.NewMockService(ctrl)
				svc.EXPECT().SearchAttributeFieldsBySecure(gomock.Any(), []string{"host"}).
					Return(map[string][]string{
						"host": {"password"},
					}, nil)
				return svc
			},
			modelUID: "host",
			testFunc: func(t *testing.T, protector IResourceProtector) {
				decrypted, err := protector.DecryptResource(context.Background(), "host", map[string]any{
					"password": "ENC:V2:dummy_secret",
				})
				assert.NoError(t, err)
				assert.Equal(t, "secret", decrypted["password"])
			},
		},
		{
			name: "敏感字段安全脱敏替换",
			mockAttr: func(ctrl *gomock.Controller) *attributemocks.MockService {
				svc := attributemocks.NewMockService(ctrl)
				svc.EXPECT().SearchAttributeFieldsBySecure(gomock.Any(), []string{"host"}).
					Return(map[string][]string{
						"host": {"password"},
					}, nil)
				return svc
			},
			modelUID: "host",
			testFunc: func(t *testing.T, protector IResourceProtector) {
				masked, err := protector.MaskResource(context.Background(), "host", map[string]any{
					"ip":       "10.0.0.1",
					"password": "super_secret_password",
				})
				assert.NoError(t, err)
				assert.Equal(t, "10.0.0.1", masked["ip"])
				assert.Equal(t, cryptox.DefaultMask, masked["password"])
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
	assert.False(t, protector.IsMasked("normal_password"))
	assert.False(t, protector.IsMasked(12345))
}
