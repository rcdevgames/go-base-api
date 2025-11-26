package entity

// LoginRequest captures credentials required to authenticate.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest carries the refresh token for renewing access tokens.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}
