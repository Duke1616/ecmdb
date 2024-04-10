package domain

type Attribute struct {
	ID         int64
	ModelID    int64
	Identifies string
	Name       string
	FieldType  string
	Required   bool
}
