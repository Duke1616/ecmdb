package dao

import "github.com/xen0n/go-workwx"

const (
	TemplateCollection      = "c_template"
	TemplateGroupCollection = "c_template_group"
)

type Template struct {
	Id                 int64                     `bson:"id"`
	Name               string                    `bson:"name"`
	CreateType         uint8                     `bson:"create_type"`
	Rules              []map[string]interface{}  `bson:"rules"`
	Options            map[string]interface{}    `bson:"options"`
	ExternalTemplateId string                    `bson:"external_template_id"`
	UniqueHash         string                    `bson:"unique_hash"`
	WechatOAControls   workwx.OATemplateControls `bson:"wechat_oa_controls,omitempty"`
	Desc               string                    `bson:"desc,omitempty"`
	Ctime              int64                     `bson:"ctime"`
	Utime              int64                     `bson:"utime"`
}

type TemplateGroup struct {
	Id    int64  `bson:"id"`
	Name  string `bson:"name"`
	Icon  string `bson:"icon"`
	Ctime int64  `bson:"ctime"`
	Utime int64  `bson:"utime"`
}
