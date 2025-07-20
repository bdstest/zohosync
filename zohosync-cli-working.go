package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// TokenInfo represents OAuth token information
type TokenInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// APIClient represents the WorkDrive API client
type APIClient struct {
	baseURL string
	token   *TokenInfo
	client  *http.Client
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// LoadTokens loads OAuth tokens from file
func (c *APIClient) LoadTokens() error {
	data, err := ioutil.ReadFile("zoho_tokens.json")
	if err != nil {
		return fmt.Errorf("failed to read tokens (run 'zohosync login' first): %v", err)
	}

	var token TokenInfo
	if err := json.Unmarshal(data, &token); err != nil {
		return fmt.Errorf("failed to parse tokens: %v", err)
	}

	c.token = &token
	return nil
}

// makeRequest makes an authenticated API request
func (c *APIClient) makeRequest(method, endpoint string) (map[string]interface{}, error) {
	if c.token == nil {
		return nil, fmt.Errorf("not authenticated - run 'zohosync login' first")
	}

	url := c.baseURL + endpoint
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Zoho-oauthtoken "+c.token.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Check for API errors
	if errors, exists := result["errors"]; exists {
		if errorList, ok := errors.([]interface{}); ok && len(errorList) > 0 {
			if errorObj, ok := errorList[0].(map[string]interface{}); ok {
				errorTitle := errorObj["title"]
				return nil, fmt.Errorf("API error: %v", errorTitle)
			}
		}
		return nil, fmt.Errorf("API error: %v", errors)
	}

	return result, nil
}

// CLI Commands

var rootCmd = &cobra.Command{
	Use:   "zohosync",
	Short: "ZohoSync - Secure Zoho WorkDrive sync client",
	Long: `ZohoSync CLI provides command-line access to Zoho WorkDrive synchronization.

Secure, lightweight sync client for Linux that keeps your files synchronized
between your local machine and Zoho WorkDrive.`,
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Zoho WorkDrive",
	Long:  "Start the OAuth 2.0 authentication flow to connect to Zoho WorkDrive",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔐 ZohoSync Authentication")
		fmt.Println("========================")
		fmt.Println()
		
		// Check if already authenticated
		client := NewAPIClient("http://localhost:8090")
		if err := client.LoadTokens(); err == nil {
			fmt.Println("✅ Already authenticated!")
			fmt.Println("   Run 'zohosync status' to check your connection")
			fmt.Println("   Run 'zohosync logout' to remove authentication")
			return
		}
		
		fmt.Println("🚀 OAuth 2.0 Authentication Required")
		fmt.Println()
		fmt.Println("1. Visit this URL in your browser:")
		fmt.Println("   https://accounts.zoho.com/oauth/v2/auth?response_type=code&client_id=1000.Z520MJ3HS00YJEKRHRX0U9KGZTATPX&scope=WorkDrive.files.ALL,WorkDrive.workspace.READ,WorkDrive.organization.READ&redirect_uri=http://localhost:8080/callback&access_type=offline")
		fmt.Println()
		fmt.Println("2. Authorize ZohoSync and copy the authorization code")
		fmt.Println("3. Run the token exchange script:")
		fmt.Println("   ./exchange-tokens.sh")
		fmt.Println()
		fmt.Println("4. Then run 'zohosync status' to verify connection")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication and sync status",
	Long:  "Display current authentication status and connection to Zoho WorkDrive",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📊 ZohoSync Status")
		fmt.Println("=================")
		fmt.Println()
		
		client := NewAPIClient("http://localhost:8090")
		
		// Check authentication
		if err := client.LoadTokens(); err != nil {
			fmt.Println("❌ Not authenticated")
			fmt.Printf("   Error: %v\n", err)
			fmt.Println("   Run 'zohosync login' to authenticate")
			return
		}
		
		fmt.Println("✅ Authentication Status: Connected")
		fmt.Printf("   Access Token: %s...\n", client.token.AccessToken[:20])
		fmt.Printf("   Token Type: %s\n", client.token.TokenType)
		fmt.Printf("   Scopes: %s\n", client.token.Scope)
		fmt.Println()
		
		// Test API connection
		fmt.Println("🔍 Testing API Connection...")
		account, err := client.makeRequest("GET", "/workdrive/api/v1/account")
		if err != nil {
			fmt.Printf("❌ API Connection Failed: %v\n", err)
			fmt.Println("   Note: Using mock API on localhost:8090")
			fmt.Println("   Start mock server: go run mock-workdrive-server.go")
			return
		}
		
		fmt.Println("✅ API Connection: Working")
		if data, exists := account["data"]; exists {
			if userInfo, ok := data.(map[string]interface{}); ok {
				fmt.Printf("   User: %v\n", userInfo["name"])
				fmt.Printf("   Email: %v\n", userInfo["email"])
			}
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List files and folders",
	Long:  "List files and folders in your Zoho WorkDrive",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📁 ZohoSync File Listing")
		fmt.Println("=======================")
		fmt.Println()
		
		client := NewAPIClient("http://localhost:8090")
		
		if err := client.LoadTokens(); err != nil {
			fmt.Printf("❌ Not authenticated: %v\n", err)
			fmt.Println("   Run 'zohosync login' first")
			return
		}
		
		files, err := client.makeRequest("GET", "/workdrive/api/v1/files")
		if err != nil {
			fmt.Printf("❌ Failed to list files: %v\n", err)
			return
		}
		
		if data, exists := files["data"]; exists {
			if fileList, ok := data.([]interface{}); ok {
				fmt.Printf("Found %d items:\n\n", len(fileList))
				
				for _, item := range fileList {
					if file, ok := item.(map[string]interface{}); ok {
						fileType := file["type"]
						fileName := file["name"]
						fileID := file["id"]
						
						icon := "📄"
						if fileType == "folder" {
							icon = "📁"
						}
						
						fmt.Printf("  %s %s\n", icon, fileName)
						fmt.Printf("     ID: %s\n", fileID)
						fmt.Printf("     Type: %s\n", fileType)
						
						if size, exists := file["size"]; exists {
							fmt.Printf("     Size: %v bytes\n", size)
						}
						fmt.Println()
					}
				}
			}
		} else {
			fmt.Println("No files found")
		}
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize files",
	Long:  "Start file synchronization between local directory and Zoho WorkDrive",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔄 ZohoSync File Synchronization")
		fmt.Println("===============================")
		fmt.Println()
		
		client := NewAPIClient("http://localhost:8090")
		
		if err := client.LoadTokens(); err != nil {
			fmt.Printf("❌ Not authenticated: %v\n", err)
			fmt.Println("   Run 'zohosync login' first")
			return
		}
		
		fmt.Println("🧪 Sync simulation (mock mode)")
		fmt.Println("✅ Checking local directory...")
		fmt.Println("✅ Checking remote WorkDrive...")
		fmt.Println("✅ Comparing file timestamps...")
		fmt.Println("✅ No conflicts detected")
		fmt.Println("📊 Sync summary:")
		fmt.Println("   - Local files: 0")
		fmt.Println("   - Remote files: 2")
		fmt.Println("   - Files to download: 0")
		fmt.Println("   - Files to upload: 0")
		fmt.Println()
		fmt.Println("🎉 Sync completed successfully!")
		fmt.Println("   Note: This is a simulation using mock API")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display ZohoSync version and build information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ZohoSync CLI v1.0.0-dev")
		fmt.Println("Build: Phase 1 Development")
		fmt.Println("OAuth: Integrated")
		fmt.Println("API: Mock Mode")
		fmt.Println()
		fmt.Println("🚀 Ready for Zoho WorkDrive API integration")
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove authentication",
	Long:  "Clear stored OAuth tokens and disconnect from Zoho WorkDrive",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚪 ZohoSync Logout")
		fmt.Println("=================")
		fmt.Println()
		
		if _, err := os.Stat("zoho_tokens.json"); os.IsNotExist(err) {
			fmt.Println("ℹ️  Not currently authenticated")
			return
		}
		
		if err := os.Remove("zoho_tokens.json"); err != nil {
			fmt.Printf("❌ Failed to remove tokens: %v\n", err)
			return
		}
		
		fmt.Println("✅ Successfully logged out")
		fmt.Println("   OAuth tokens removed")
		fmt.Println("   Run 'zohosync login' to reconnect")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(logoutCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}