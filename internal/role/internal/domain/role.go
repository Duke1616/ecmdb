package domain

type Role struct {
	Id      int64
	Code    string
	Name    string
	Desc    string
	Status  bool
	MenuIds []int64
}
