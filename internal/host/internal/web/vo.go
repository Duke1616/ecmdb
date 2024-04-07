package web

type CreateHostReq struct {
	Host     Host     `json:"host"`
	Resource Resource `json:"resource"`
}

type Host struct {
	Name      string    `json:"name"`                // 主机名称
	CPU       string    `json:"CPU"`                 // CPU核心数
	Memory    string    `json:"memory"`              // 内存容量
	PrivateIP string    `json:"privateIP,omitempty"` // 私网IP地址
	EIP       string    `json:"EIP,omitempty"`       // 弹性IP地址
	OsType    string    `json:"osType"`              // 操作系统类型
	OsName    string    `json:"osName"`              // 操作系统名称
	GPUInfo   []GPUInfo `json:"GPUInfo,omitempty"`   // GPU信息
}

type GPUInfo struct {
	Name   string `json:"name"`   // GPU 型号
	Memory string `json:"memory"` // GPU 显存大小
}

type Resource struct {
}
