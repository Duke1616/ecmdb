package web

type CreateTunnelReq struct {
	Width  string `json:"width"`
	Height string `json:"height"`
	Dpi    string `json:"dpi"`
}

type ConnectReq struct {
	Type       string `json:"type"`
	ResourceId int64  `json:"resource_id"`
}
