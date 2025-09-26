package jwt

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ClientInterceptorBuilder 客户端拦截器构建器
type ClientInterceptorBuilder struct {
	jwtKey string
}

// NewClientInterceptorBuilder 创建客户端拦截器构建器
func NewClientInterceptorBuilder(jwtKey string) *ClientInterceptorBuilder {
	return &ClientInterceptorBuilder{
		jwtKey: jwtKey,
	}
}

// UnaryClientInterceptor 创建一元客户端拦截器
func (b *ClientInterceptorBuilder) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 检查 context 中是否已经有 JWT 信息
		if b.hasJWTInContext(ctx) {
			// 如果已经有 JWT 信息，直接调用
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		// 自动注入 JWT context
		jwtCtx := b.injectJWTContext(ctx)
		return invoker(jwtCtx, method, req, reply, cc, opts...)
	}
}

// hasJWTInContext 检查 context 中是否已经有 JWT 信息
func (b *ClientInterceptorBuilder) hasJWTInContext(ctx context.Context) bool {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return false
	}

	authHeaders := md.Get("Authorization")
	return len(authHeaders) > 0
}

// injectJWTContext 注入 JWT context
func (b *ClientInterceptorBuilder) injectJWTContext(ctx context.Context) context.Context {
	// 使用项目已有的JWT包创建令牌
	jwtAuth := NewJwtAuth(b.jwtKey)

	// 创建包含业务ID的声明
	claims := jwt.MapClaims{
		"biz_id": float64(1),
	}

	// 使用JWT认证包的Encode方法生成令牌
	tokenString, err := jwtAuth.Encode(claims)
	if err != nil {
		// 如果生成失败，返回原始 context
		return ctx
	}

	// 创建带有授权信息的元数据
	md := metadata.New(map[string]string{
		"Authorization": "Bearer " + tokenString,
	})
	return metadata.NewOutgoingContext(ctx, md)
}
