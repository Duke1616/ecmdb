package web

type GetPresignedUrl struct {
	ObjectName string `json:"object_name"`
}

type PutPresignedUrl struct {
	ObjectName string `json:"object_name"`
}

type RemoveObjectReq struct {
	ObjectName string `json:"object_name"`
}
