package domain

type Attribute struct {
	ID              int64
	ModelIdentifies string
	Identifies      string
	Name            string
	FieldType       string
	Required        bool
}
