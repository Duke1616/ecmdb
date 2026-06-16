package errs

// 定义统一的错误类型
var (
	ErrUniqueDuplicate = ErrorCode{Code: 500001, Msg: "唯一标识冲突"}
	ErrNotFound        = ErrorCode{Code: 500002, Msg: "数据不存在"}
	SystemError        = ErrorCode{Code: 502001, Msg: "系统错误"}
)

type ErrorCode struct {
	Code int
	Msg  string
}

func (e ErrorCode) Error() string {
	return e.Msg
}

func (e ErrorCode) GetCode() int {
	return e.Code
}

func (e ErrorCode) GetMsg() string {
	return e.Msg
}

func (e ErrorCode) Is(target error) bool {
	t, ok := target.(ErrorCode)
	if !ok {
		return false
	}
	return e.Code == t.Code
}
