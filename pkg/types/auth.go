package types

import "time"

// TokenInfo represents OAuth token information
type TokenInfo struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope"`
}

// AuthState represents the current authentication state
type AuthState struct {
	IsAuthenticated bool       `json:"is_authenticated"`
	UserID          string     `json:"user_id"`
	UserEmail       string     `json:"user_email"`
	Token           *TokenInfo `json:"token,omitempty"`
}
