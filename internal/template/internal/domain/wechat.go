package domain

import (
	"github.com/xen0n/go-workwx"
)

type WechatInfo struct {
	Id       string
	Name     string
	Controls workwx.OATemplateControls
}
