//go:build unit

package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestJwtAuth_Encode(t *testing.T) {
	// 创建 JwtAuth 实例
	testKey := "test_key"
	jwtAuth := NewJwtAuth(testKey)

	// 测试场景
	tests := []struct {
		name         string
		customClaims jwt.MapClaims
		wantErr      bool
	}{
		{
			name:         "基本令牌生成",
			customClaims: jwt.MapClaims{},
			wantErr:      false,
		},
		{
			name: "带用户ID的令牌",
			customClaims: jwt.MapClaims{
				"biz_id": float64(23456),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtAuth.Encode(tt.customClaims)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// 验证生成的令牌可以被解析
			claims, err := jwtAuth.Decode(token)
			assert.NoError(t, err)

			// 验证标准声明存在
			assert.NotEmpty(t, claims["iat"])
			assert.NotEmpty(t, claims["exp"])
			assert.Equal(t, "notification-platform", claims["iss"])

			// 验证自定义声明存在
			for k, v := range tt.customClaims {
				assert.Equal(t, v, claims[k])
			}
		})
	}
}

func TestJwtAuth_Decode(t *testing.T) {
	// 创建 JwtAuth 实例
	testKey := "test-secret-key"
	jwtAuth := NewJwtAuth(testKey)

	// 创建一个有效的令牌用于测试
	validClaims := jwt.MapClaims{
		"user_id": "123456",
		"role":    "admin",
	}
	validToken, err := jwtAuth.Encode(validClaims)
	assert.NoError(t, err)

	// 创建一个已过期的令牌
	expiredClaims := jwt.MapClaims{
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
	}
	expiredToken, err := jwtAuth.Encode(expiredClaims)
	assert.NoError(t, err)

	// 测试场景
	tests := []struct {
		name       string
		tokenInput string
		wantErr    bool
	}{
		{
			name:       "有效令牌解析",
			tokenInput: validToken,
			wantErr:    false,
		},
		{
			name:       "带Bearer前缀的有效令牌",
			tokenInput: "Bearer " + validToken,
			wantErr:    false,
		},
		{
			name:       "已过期令牌",
			tokenInput: expiredToken,
			wantErr:    true,
		},
		{
			name:       "无效令牌格式",
			tokenInput: "invalid.token.format",
			wantErr:    true,
		},
		{
			name:       "空令牌",
			tokenInput: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtAuth.Decode(tt.tokenInput)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, claims)

			// 当令牌有效时，验证声明值
			if !tt.wantErr && tt.tokenInput == validToken || tt.tokenInput == "Bearer "+validToken {
				assert.Equal(t, "123456", claims["user_id"])
				assert.Equal(t, "admin", claims["role"])
			}
		})
	}
}

func TestNewJwtAuth(t *testing.T) {
	testKey := "test-secret-key"
	jwtAuth := NewJwtAuth(testKey)

	assert.NotNil(t, jwtAuth)
	// 生成令牌测试实例是否正常工作
	token, err := jwtAuth.Encode(jwt.MapClaims{})
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}
