package web

type UpdateBindingEnabledReq struct {
	UID     string `json:"uid" binding:"required"`
	Enabled bool   `json:"enabled"`
}
