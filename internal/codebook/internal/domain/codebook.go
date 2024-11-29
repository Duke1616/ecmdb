package domain

type Codebook struct {
	Id         int64
	Name       string
	Owner      int64
	Code       string
	Language   string
	Secret     string
	Identifier string
}
