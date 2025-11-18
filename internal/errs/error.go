package errs

import "errors"

// ErrUniqueDuplicate 定义统一的错误类型
var (
	ErrUniqueDuplicate = errors.New("唯一标识冲突")
)
