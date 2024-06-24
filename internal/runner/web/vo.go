package web

type PublishRunnerReq struct {
	Name     string `json:"name"`
	UUID     string `json:"uuid"`
	Language string `json:"language"`
	Code     string `json:"code"`
	Topic    string `json:"topic"`
}
