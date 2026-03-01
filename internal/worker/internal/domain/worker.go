package domain

type Execute struct {
	TaskId    int64
	Topic     string
	Handler   string
	Language  string
	Code      string
	Args      map[string]interface{}
	Variables string
}
