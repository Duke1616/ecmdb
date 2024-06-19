package web

type CreateCodebookReq struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	Language string `json:"language"`
}

type DetailCodebookReq struct {
	Id int64 `json:"id"`
}

type Page struct {
	Offset int64 `json:"offset,omitempty"`
	Limit  int64 `json:"limit,omitempty"`
}

type ListCodebookReq struct {
	Page
}

type Codebook struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Language string `json:"language"`
}

type RetrieveCodebooks struct {
	Total     int64      `json:"total"`
	Codebooks []Codebook `json:"codebooks"`
}
