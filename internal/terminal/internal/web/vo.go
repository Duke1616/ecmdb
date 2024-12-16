package web

type CreateTunnelReq struct {
	Width  string `json:"width"`
	Height string `json:"height"`
	Dpi    string `json:"dpi"`
}
