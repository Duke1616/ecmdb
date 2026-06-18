package errs

import "errors"

var (
	ErrConcurrentUpdate = errors.New("数据已被其他用户修改，请刷新后重试")
)
