package web

type GetPresignedUrl struct {
	ObjectName string `json:"object_name"`
	Bucket     string `json:"bucket"`
	Prefix     string `json:"prefix"`
}

type PutPresignedUrl struct {
	ObjectName string `json:"object_name"`
	Bucket     string `json:"bucket"`
	Prefix     string `json:"prefix"`
}

type RemoveObjectReq struct {
	ObjectName string `json:"object_name"`
	Bucket     string `json:"bucket"`
	Prefix     string `json:"prefix"`
}
