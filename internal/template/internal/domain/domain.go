package domain

import "github.com/xen0n/go-workwx"

type Template struct {
	Id                 string
	CreateType         string
	WechatApprovalInfo workwx.OAApprovalInfo
}
