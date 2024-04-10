package domain

type Resource struct {
	ID      int64
	ModelID int64
	Data    map[string]interface{}
}
