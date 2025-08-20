package errs

var (
	SystemError   = ErrorCode{Code: 503001, Msg: "系统错误"}
	UserPassError = ErrorCode{Code: 504002, Msg: "账号或密码输入不正确"}
)

type ErrorCode struct {
	Code int
	Msg  string
}
