package web

type PluginDetailReq struct {
	UID string `form:"uid" json:"uid"`
}

type DeletePluginReq struct {
	UID string `json:"uid"`
}
