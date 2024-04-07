package domain

type Host struct {
	ID         int64
	ResourceID int64
	Name       string
	CPU        string
	Memory     string
	PrivateIP  string
	EIP        string
	OsType     string // 操作系统类型
	OsName     string // 操作系统名称
	GPUInfo    []GPUInfo
}

type GPUInfo struct {
	Name   string
	Memory string
}
