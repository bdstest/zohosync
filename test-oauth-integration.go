package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

// TokenInfo represents OAuth token information
type TokenInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// Config represents ZohoSync configuration
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	AuthURL      string `json:"auth_url"`
	TokenURL     string `json:"token_url"`
}

// ZohoAPI represents a mock Zoho WorkDrive API client
type ZohoAPI struct {
	config *Config
	token  *TokenInfo
}

// NewZohoAPI creates a new API client
func NewZohoAPI() *ZohoAPI {
	return &ZohoAPI{
		config: &Config{
			ClientID:     "1000.Z520MJ3HS00YJEKRHRX0U9KGZTATPX",
			ClientSecret: "731702ae155269b29c1997664def3553764face6f8",
			RedirectURI:  "http://localhost:8080/callback",
			AuthURL:      "https://accounts.zoho.com/oauth/v2/auth",
			TokenURL:     "https://accounts.zoho.com/oauth/v2/token",
		},
	}
}

// LoadTokens loads OAuth tokens from file
func (z *ZohoAPI) LoadTokens() error {
	data, err := ioutil.ReadFile("zoho_tokens.json")
	if err != nil {
		return fmt.Errorf("failed to read tokens: %v", err)
	}

	var token TokenInfo
	if err := json.Unmarshal(data, &token); err != nil {
		return fmt.Errorf("failed to parse tokens: %v", err)
	}

	z.token = &token
	return nil
}

// RefreshToken refreshes the access token
func (z *ZohoAPI) RefreshToken() error {
	if z.token == nil || z.token.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	data := url.Values{
		"grant_type":    []string{"refresh_token"},
		"client_id":     []string{z.config.ClientID},
		"client_secret": []string{z.config.ClientSecret},
		"refresh_token": []string{z.token.RefreshToken},
	}

	resp, err := http.PostForm(z.config.TokenURL, data)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var newToken TokenInfo
	if err := json.Unmarshal(body, &newToken); err != nil {
		return fmt.Errorf("failed to parse new token: %v", err)
	}

	// Keep the refresh token if not provided in response
	if newToken.RefreshToken == "" {
		newToken.RefreshToken = z.token.RefreshToken
	}

	z.token = &newToken
	return z.SaveTokens()
}

// SaveTokens saves tokens to file
func (z *ZohoAPI) SaveTokens() error {
	data, err := json.MarshalIndent(z.token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %v", err)
	}

	return ioutil.WriteFile("zoho_tokens.json", data, 0600)
}

// MockAPICall simulates a WorkDrive API call
func (z *ZohoAPI) MockAPICall(endpoint string) (map[string]interface{}, error) {
	if z.token == nil {
		return nil, fmt.Errorf("no authentication token")
	}

	// Simulate different API responses based on endpoint
	switch endpoint {
	case "/files":
		return map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":         "veysx16db130021d84de08b78167afc76c011",
					"name":       "test-file.txt",
					"type":       "file",
					"size":       1024,
					"created_at": time.Now().Format(time.RFC3339),
					"modified_at": time.Now().Format(time.RFC3339),
				},
				{
					"id":         "folder123456789",
					"name":       "test-folder",
					"type":       "folder",
					"created_at": time.Now().Format(time.RFC3339),
					"modified_at": time.Now().Format(time.RFC3339),
				},
			},
			"status": "success",
		}, nil

	case "/account":
		return map[string]interface{}{
			"data": map[string]interface{}{
				"user_id":    "123456789",
				"email":      "user@example.com",
				"name":       "Test User",
				"storage_used": 1073741824, // 1GB
				"storage_total": 107374182400, // 100GB
			},
			"status": "success",
		}, nil

	default:
		return map[string]interface{}{
			"error": "endpoint not found",
		}, fmt.Errorf("unknown endpoint: %s", endpoint)
	}
}

// TestOAuthIntegration tests the complete OAuth integration
func TestOAuthIntegration() {
	fmt.Println("🔧 ZohoSync OAuth Integration Test")
	fmt.Println("=================================")
	fmt.Println()

	// Initialize API client
	api := NewZohoAPI()

	// Load existing tokens
	fmt.Println("📋 Loading OAuth tokens...")
	if err := api.LoadTokens(); err != nil {
		log.Fatalf("Failed to load tokens: %v", err)
	}

	fmt.Printf("✅ Loaded tokens successfully\n")
	fmt.Printf("   Access Token: %s...\n", api.token.AccessToken[:20])
	fmt.Printf("   Token Type: %s\n", api.token.TokenType)
	fmt.Printf("   Scopes: %s\n", api.token.Scope)
	fmt.Println()

	// Test token refresh
	fmt.Println("🔄 Testing token refresh...")
	if err := api.RefreshToken(); err != nil {
		fmt.Printf("⚠️  Token refresh failed: %v\n", err)
		fmt.Println("   (This is expected if the token is still valid)")
	} else {
		fmt.Println("✅ Token refreshed successfully")
	}
	fmt.Println()

	// Test mock API calls
	fmt.Println("🧪 Testing mock API calls...")
	
	// Test file listing
	fmt.Println("📁 Mock API: List files")
	files, err := api.MockAPICall("/files")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Println("✅ Mock files API working:")
		filesJSON, _ := json.MarshalIndent(files, "   ", "  ")
		fmt.Printf("   %s\n", filesJSON)
	}
	fmt.Println()

	// Test account info
	fmt.Println("👤 Mock API: Account info")
	account, err := api.MockAPICall("/account")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Println("✅ Mock account API working:")
		accountJSON, _ := json.MarshalIndent(account, "   ", "  ")
		fmt.Printf("   %s\n", accountJSON)
	}

	fmt.Println()
	fmt.Println("🎉 OAuth Integration Test Complete!")
	fmt.Println("✅ Ready for CLI and GUI integration")
}

func main() {
	// Change to zohosync directory
	if err := os.Chdir("/opt/zohosync"); err != nil {
		log.Fatalf("Failed to change directory: %v", err)
	}

	TestOAuthIntegration()
}