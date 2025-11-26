package entity

// User represents authentication subject data.
type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	Role         string
}

// TokenPair contains access and refresh tokens.
type TokenPair struct {
	AccessToken        string `json:"access_token"`
	RefreshToken       string `json:"refresh_token"`
}
