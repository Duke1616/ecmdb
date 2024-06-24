package errs

var (
	SystemError = ErrorCode{Code: 504001, Msg: "系统错误"}
)

type ErrorCode struct {
	Code int
	Msg  string
}

func (e ErrorCode) Error() string {
	//TODO implement me
	panic("implement me")
}
