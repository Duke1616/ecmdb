package domain

type Codebook struct {
	Id         int64
	Name       string
	Owner      string
	Code       string
	Language   string
	Secret     string
	Identifier string
}
