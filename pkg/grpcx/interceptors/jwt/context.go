package jwt

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/metadata"
)

func ContextWithJWT(ctx context.Context, key string) context.Context {
	// 使用项目已有的JWT包创建令牌
	jwtAuth := NewJwtAuth(key)

	// 创建包含业务ID的声明
	claims := jwt.MapClaims{
		"biz_id": float64(1),
	}

	// 使用JWT认证包的Encode方法生成令牌
	tokenString, _ := jwtAuth.Encode(claims)

	// 创建带有授权信息的元数据
	md := metadata.New(map[string]string{
		"Authorization": "Bearer " + tokenString,
	})
	return metadata.NewOutgoingContext(ctx, md)
}
