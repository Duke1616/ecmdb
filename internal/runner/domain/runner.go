package domain

type Runner struct {
	Topic    string
	Name     string // 执行名称
	UUID     string // 唯一标识
	Language string // 语言
	Code     string // 代码
}
