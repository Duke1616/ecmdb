package errs

import "errors"

// 定义统一的错误类型
var (
	ErrUniqueDuplicate = errors.New("唯一标识冲突")
	ErrNotFound        = errors.New("数据不存在")
)
