// Package auth handles OAuth 2.0 authentication with Zoho WorkDrive
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"errors"
	"strings"

	"github.com/bdstest/zohosync/pkg/types"
	"github.com/bdstest/zohosync/internal/config"
	"github.com/bdstest/zohosync/internal/utils"
	"golang.org/x/oauth2"
)

// OAuthClient handles OAuth 2.0 authentication flow
type OAuthClient struct {
	config      *oauth2.Config
	verifier    string
	challenge   string
	state       string
	redirectURI string
	logger      *utils.Logger
}

// NewOAuthClient creates a new OAuth client
func NewOAuthClient(cfg *types.Config) *OAuthClient {
	return &OAuthClient{
		config: &oauth2.Config{
			ClientID:     cfg.Auth.ClientID,
			ClientSecret: cfg.Auth.ClientSecret,
			RedirectURL:  cfg.Auth.RedirectURI,
			Scopes:       cfg.Auth.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  config.AuthURL,
				TokenURL: config.TokenURL,
			},
		},
		redirectURI: cfg.Auth.RedirectURI,
		logger:      utils.GetLogger(),
	}
}

// GeneratePKCE generates PKCE code verifier and challenge
func (o *OAuthClient) GeneratePKCE() error {
	// Generate code verifier (43-128 characters)
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return fmt.Errorf("failed to generate code verifier: %w", err)
	}
	o.verifier = base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Generate code challenge (SHA256 hash of verifier)
	hash := sha256.Sum256([]byte(o.verifier))
	o.challenge = base64.RawURLEncoding.EncodeToString(hash[:])

	return nil
}

// GenerateState generates a random state parameter
func (o *OAuthClient) GenerateState() error {
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}
	o.state = base64.RawURLEncoding.EncodeToString(stateBytes)
	return nil
}

// GetAuthURL returns the OAuth authorization URL with PKCE
func (o *OAuthClient) GetAuthURL() (string, error) {
	if err := o.GeneratePKCE(); err != nil {
		return "", err
	}
	
	if err := o.GenerateState(); err != nil {
		return "", err
	}

	authURL := o.config.AuthCodeURL(o.state,
		oauth2.SetAuthURLParam("code_challenge", o.challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("access_type", "offline"),
	)

	o.logger.Info("Generated OAuth URL with PKCE")
	return authURL, nil
}

// ExchangeCodeForToken exchanges authorization code for access token
func (o *OAuthClient) ExchangeCodeForToken(ctx context.Context, code, state string) (*types.TokenInfo, error) {
	// Verify state parameter
	if state != o.state {
		return nil, fmt.Errorf("invalid state parameter")
	}

	// Exchange code for token with PKCE
	token, err := o.config.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", o.verifier),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Convert to our token format
	tokenInfo := &types.TokenInfo{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry,
		Scope:        "",
	}

	if token.Valid() {
		tokenInfo.ExpiresIn = int(time.Until(token.Expiry).Seconds())
	}

	o.logger.Info("Successfully exchanged code for token")
	return tokenInfo, nil
}

// RefreshToken refreshes an expired access token
func (o *OAuthClient) RefreshToken(ctx context.Context, refreshToken string) (*types.TokenInfo, error) {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := o.config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	tokenInfo := &types.TokenInfo{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		TokenType:    newToken.TokenType,
		ExpiresAt:    newToken.Expiry,
	}

	if newToken.Valid() {
		tokenInfo.ExpiresIn = int(time.Until(newToken.Expiry).Seconds())
	}

	o.logger.Info("Successfully refreshed token")
	return tokenInfo, nil
}

// ValidateOAuthConfig validates OAuth configuration
func ValidateOAuthConfig(config *OAuthConfig) error {
	if config.ClientID == "" {
		return errors.New("client ID is required")
	}
	if config.ClientSecret == "" {
		return errors.New("client secret is required")
	}
	if config.RedirectURI == "" {
		return errors.New("redirect URI is required")
	}
	
	// Validate redirect URI format
	if _, err := url.Parse(config.RedirectURI); err != nil {
		return fmt.Errorf("invalid redirect URI: %w", err)
	}
	
	return nil
}

// IsTokenValid checks if a token is still valid
func IsTokenValid(token *Token) bool {
	if token == nil || token.AccessToken == "" {
		return false
	}
	return time.Now().Before(token.ExpiresAt)
}

// GenerateCodeVerifier generates a PKCE code verifier
func GenerateCodeVerifier() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	b := make([]byte, 128)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// GenerateCodeChallenge generates a PKCE code challenge from verifier
func GenerateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.URLEncoding.EncodeToString(hash[:])
	return strings.TrimRight(challenge, "=")
}

// RefreshToken refreshes an access token using refresh token
func RefreshToken(config *OAuthConfig, token *Token) (*Token, error) {
	if config == nil || token == nil || token.RefreshToken == "" {
		return nil, errors.New("invalid config or token")
	}
	
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", token.RefreshToken)
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	
	resp, err := http.PostForm(config.TokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status: %d", resp.StatusCode)
	}
	
	// Parse response and return new token
	// Implementation would parse JSON response and create new Token
	return &Token{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token", 
		ExpiresAt:    time.Now().Add(time.Hour),
	}, nil
}

// OAuthConfig represents OAuth configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	TokenURL     string
	Scopes       []string
}

// Token represents an OAuth token
type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// ValidateToken validates if a token is still valid
func (o *OAuthClient) ValidateToken(token *types.TokenInfo) bool {
	if token == nil || token.AccessToken == "" {
		return false
	}

	// Check if token is expired (with 5 minute buffer)
	if time.Now().Add(5 * time.Minute).After(token.ExpiresAt) {
		return false
	}

	return true
}

// StartCallbackServer starts a local HTTP server for OAuth callback
func (o *OAuthClient) StartCallbackServer(ctx context.Context) (*types.TokenInfo, error) {
	resultChan := make(chan *types.TokenInfo, 1)
	errorChan := make(chan error, 1)

	// Parse redirect URI to get port
	redirectURL, err := url.Parse(o.redirectURI)
	if err != nil {
		return nil, fmt.Errorf("invalid redirect URI: %w", err)
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%s", redirectURL.Port()),
	}

	// Handle OAuth callback
	http.HandleFunc(redirectURL.Path, func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		errorParam := r.URL.Query().Get("error")

		if errorParam != "" {
			errorChan <- fmt.Errorf("OAuth error: %s", errorParam)
			fmt.Fprintf(w, "<h1>Authentication Failed</h1><p>Error: %s</p>", errorParam)
			return
		}

		if code == "" {
			errorChan <- fmt.Errorf("no authorization code received")
			fmt.Fprintf(w, "<h1>Authentication Failed</h1><p>No authorization code received</p>")
			return
		}

		// Exchange code for token
		token, err := o.ExchangeCodeForToken(r.Context(), code, state)
		if err != nil {
			errorChan <- err
			fmt.Fprintf(w, "<h1>Authentication Failed</h1><p>Error: %s</p>", err.Error())
			return
		}

		resultChan <- token
		fmt.Fprintf(w, "<h1>Authentication Successful!</h1><p>You can now close this window and return to ZohoSync.</p>")
	})

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errorChan <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	// Wait for result or timeout
	select {
	case token := <-resultChan:
		server.Close()
		return token, nil
	case err := <-errorChan:
		server.Close()
		return nil, err
	case <-ctx.Done():
		server.Close()
		return nil, fmt.Errorf("authentication timeout")
	}
}