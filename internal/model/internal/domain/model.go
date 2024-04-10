package domain

type Model struct {
	ID      int64
	GroupId int64
	Name    string
}

type ModelGroup struct {
	ID   int64
	Name string
}
