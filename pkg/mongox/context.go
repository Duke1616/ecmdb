package mongox

import "context"

const (
	// IgnoreTenantKey 跳过租户隔离校验的 Context Key
	IgnoreTenantKey = "mongox:ignore_tenant"
)

// IgnoreTenantContext 将跳过租户隔离标记注入 Context，允许全局访问数据
func IgnoreTenantContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, IgnoreTenantKey, true)
}

// IsIgnoreTenant 检查是否处于跳过隔离校验模式
func IsIgnoreTenant(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	val, ok := ctx.Value(IgnoreTenantKey).(bool)
	return ok && val
}
