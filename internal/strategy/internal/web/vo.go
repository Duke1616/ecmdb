package web

type GetSpecifiedTemplate struct {
	Id int64
}

type CreateStrategyReq struct {
	Key   string
	Value string
	Op    op
}

// op 代表操作符
type op string

const (
	opAND = "AND"
	opOR  = "OR"
)

func (o op) String() string {
	return string(o)
}
