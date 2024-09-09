package easyflow

type Rule struct {
	Type  string `json:"type"`
	Field string `json:"field"`
	Title string `json:"title"`
}

type Rules []Rule
