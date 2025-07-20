package auth

import (
	"testing"
	"time"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthFlow(t *testing.T) {
	tests := []struct {
		name           string
		clientID       string
		clientSecret   string
		redirectURI    string
		expectedError  bool
	}{
		{
			name:          "Valid OAuth configuration",
			clientID:      "test_client_id",
			clientSecret:  "test_client_secret", 
			redirectURI:   "http://localhost:8080/callback",
			expectedError: false,
		},
		{
			name:          "Empty client ID",
			clientID:      "",
			clientSecret:  "test_client_secret",
			redirectURI:   "http://localhost:8080/callback",
			expectedError: true,
		},
		{
			name:          "Invalid redirect URI",
			clientID:      "test_client_id",
			clientSecret:  "test_client_secret",
			redirectURI:   "invalid-uri",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &OAuthConfig{
				ClientID:     tt.clientID,
				ClientSecret: tt.clientSecret,
				RedirectURI:  tt.redirectURI,
				Scopes:       []string{"WorkDrive.workspace.READ", "WorkDrive.files.ALL"},
			}

			err := ValidateOAuthConfig(config)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTokenRefresh(t *testing.T) {
	// Mock Zoho token endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/v2/token" {
			response := map[string]interface{}{
				"access_token":  "new_access_token",
				"refresh_token": "new_refresh_token",
				"expires_in":    3600,
				"token_type":    "Bearer",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	config := &OAuthConfig{
		ClientID:     "test_client",
		ClientSecret: "test_secret",
		TokenURL:     server.URL + "/oauth/v2/token",
	}

	token := &Token{
		AccessToken:  "old_access_token",
		RefreshToken: "old_refresh_token",
		ExpiresAt:    time.Now().Add(-time.Hour), // Expired
	}

	newToken, err := RefreshToken(config, token)
	require.NoError(t, err)
	assert.Equal(t, "new_access_token", newToken.AccessToken)
	assert.Equal(t, "new_refresh_token", newToken.RefreshToken)
	assert.True(t, newToken.ExpiresAt.After(time.Now()))
}

func TestTokenValidation(t *testing.T) {
	tests := []struct {
		name     string
		token    *Token
		isValid  bool
	}{
		{
			name: "Valid token",
			token: &Token{
				AccessToken: "valid_token",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
			isValid: true,
		},
		{
			name: "Expired token", 
			token: &Token{
				AccessToken: "expired_token",
				ExpiresAt:   time.Now().Add(-time.Hour),
			},
			isValid: false,
		},
		{
			name: "Empty token",
			token: &Token{
				AccessToken: "",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := IsTokenValid(tt.token)
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

func TestPKCE(t *testing.T) {
	// Test PKCE code verifier generation
	verifier := GenerateCodeVerifier()
	assert.NotEmpty(t, verifier)
	assert.GreaterOrEqual(t, len(verifier), 43)
	assert.LessOrEqual(t, len(verifier), 128)

	// Test PKCE code challenge generation
	challenge := GenerateCodeChallenge(verifier)
	assert.NotEmpty(t, challenge)
	assert.NotEqual(t, verifier, challenge)

	// Verify challenge is URL-safe base64
	assert.NotContains(t, challenge, "+")
	assert.NotContains(t, challenge, "/")
	assert.NotContains(t, challenge, "=")
}