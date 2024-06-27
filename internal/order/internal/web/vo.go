package web

type CreateOrderReq struct {
	Applicant string                 `json:"applicant"`
	Data      map[string]interface{} `json:"data"`
}
