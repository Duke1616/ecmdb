package errs

var (
	SystemError   = ErrorCode{Code: 503001, Msg: "系统错误"}
	ValidateError = ErrorCode{Code: 503002, Msg: "校验错误"}
)

type ErrorCode struct {
	Code int
	Msg  string
}
