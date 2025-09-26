package incr

import "context"

// InitialIncr 版本增量更新接口
type InitialIncr interface {
	Version() string                    // 返回版本号
	Rollback(ctx context.Context) error // 回滚逻辑
	Commit(ctx context.Context) error   // 更新逻辑
	Before(ctx context.Context) error   // 更新前逻辑
	After(ctx context.Context) error    // 更新后逻辑
}
