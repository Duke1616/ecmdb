package web

type PluginDetailReq struct {
	UID string `form:"uid" json:"uid"`
}

type TogglePluginReq struct {
	UID     string `json:"uid"`
	Enabled bool   `json:"enabled"`
}

type DeletePluginReq struct {
	UID string `json:"uid"`
}
