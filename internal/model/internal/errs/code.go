package errs

var (
	SystemError              = ErrorCode{Code: 501001, Msg: "系统错误"}
	RelationIsNotFountResult = ErrorCode{Code: 501001, Msg: "模型关联关系不为空"}
)

type ErrorCode struct {
	Code int
	Msg  string
}

func (e ErrorCode) Error() string {
	//TODO implement me
	panic("implement me")
}
