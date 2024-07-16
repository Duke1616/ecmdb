package domain

import (
	"github.com/xen0n/go-workwx"
)

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
	Id                 int64
	Name               string
	WorkflowId         int64
	GroupId            int64
	Icon               string
	CreateType         CreateType
	UniqueHash         string
	ExternalTemplateId string
	WechatOAControls   workwx.OATemplateControls
	Rules              []map[string]interface{}
	Options            map[string]interface{}
	Desc               string
}

type TemplateCombination struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	Icon       string `json:"icon"`
	Total      int    `json:"total"`
	CreateType CreateType
	Templates  []Template `json:"templates"`
}
