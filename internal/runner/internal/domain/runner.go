package domain

type Action uint8

func (s Action) ToUint8() uint8 {
	return uint8(s)
}

const (
	// REGISTER 注册
	REGISTER Action = 1
	// UNREGISTER 注销
	UNREGISTER Action = 2
)

type Runner struct {
	Id             int64
	Name           string
	CodebookUid    string
	CodebookSecret string
	WorkerName     string
	Topic          string
	Tags           []string
	Desc           string
	Action         Action
	Variables      []Variables
}

type Variables struct {
	Key    string
	Value  any
	Secret bool
}

type RunnerTags struct {
	CodebookUid      string
	TagsMappingTopic map[string]string
}
