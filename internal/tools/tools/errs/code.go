package errs

var (
	SystemError     = ErrorCode{Code: 503001, Msg: "系统错误"}
	ValidationError = ErrorCode{Code: 503002, Msg: "验证错误"}
)

type ErrorCode struct {
	Code int
	Msg  string
}
