package errs

import "errors"

// 定义统一的错误类型
var (
	ErrUniqueDuplicate = errors.New("唯一标识冲突")
	ErrNotFound        = errors.New("数据不存在")
	SystemError        = ErrorCode{Code: 502001, Msg: "系统错误"}
)

type ErrorCode struct {
	Code int
	Msg  string
}

func (e ErrorCode) Error() string {
	return e.Msg
}
