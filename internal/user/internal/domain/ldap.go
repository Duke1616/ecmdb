package domain

type Profile struct {
	DN          string   `json:"dn"`
	Email       string   `json:"email"`
	Username    string   `json:"username"`
	Title       string   `json:"title"`
	DisplayName string   `json:"display_name"`
	Groups      []string `json:"groups"`
}
