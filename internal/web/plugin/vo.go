package web

type ListResourceActionsBatchReq struct {
	ResourceIDs []int64 `json:"resource_ids" binding:"required"`
}
