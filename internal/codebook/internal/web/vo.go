package web

type CreateCodebookReq struct {
	Name       string `json:"name"`
	Code       string `json:"code"`
	Language   string `json:"language"`
	Identifier string `json:"identifier"`
}

type DetailCodebookReq struct {
	Id int64 `json:"id"`
}

type UpdateCodebookReq struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Language string `json:"language"`
}

type DeleteCodebookReq struct {
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
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	Code       string `json:"code"`
	Language   string `json:"language"`
	Secret     string `json:"secret"`
}

type RetrieveCodebooks struct {
	Total     int64      `json:"total"`
	Codebooks []Codebook `json:"codebooks"`
}
