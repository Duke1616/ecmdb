package web

type CreateTunnelReq struct {
	Width  string `json:"width"`
	Height string `json:"height"`
	Dpi    string `json:"dpi"`
}

// ConnectType 表示终端连接的类型，用于替代裸字符串，便于统一管理和扩展。
type ConnectType string

const (
	ConnectTypeSSH     ConnectType = "SSH"
	ConnectTypeWebSftp ConnectType = "Web Sftp"
	ConnectTypeRDP     ConnectType = "RDP"
	ConnectTypeVNC     ConnectType = "VNC"
)

type ConnectReq struct {
	Type       ConnectType `json:"type"`
	ResourceId int64       `json:"resource_id"`
}
