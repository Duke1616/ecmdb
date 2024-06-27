package domain

type Order struct {
	Id        int64
	Applicant string
	Data      map[string]interface{}
}
