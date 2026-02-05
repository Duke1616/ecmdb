package errs

var (
	SystemError     = ErrorCode{Code: 503001, Msg: "系统错误"}
	MenuHasChildren = ErrorCode{Code: 503002, Msg: "存在子菜单，无法删除"}
)

type ErrorCode struct {
	Code int
	Msg  string
}

func (e ErrorCode) Error() string {
	return e.Msg
}
