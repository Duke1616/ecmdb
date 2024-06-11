package domain

type CreateType uint8

func (s CreateType) ToUint8() uint8 {
	return uint8(s)
}

const (
	// SystemCreate 系统创建
	SystemCreate CreateType = 1
	// WechatCreate 企业微信创建 OR 同步
	WechatCreate CreateType = 2
)

type Template struct {
	Id           int64
	CreateType   CreateType
	WechatOAInfo WechatInfo
	UniqueHash   string
}
