package config

// Default configuration values
const (
	DefaultAppName     = "ZohoSync"
	DefaultLogLevel    = "info"
	DefaultSyncInterval = 300 // seconds
	DefaultTimeout     = 30   // seconds
	DefaultMaxRetries  = 3
	
	// OAuth endpoints
	AuthURL  = "https://accounts.zoho.com/oauth/v2/auth"
	TokenURL = "https://accounts.zoho.com/oauth/v2/token"
	
	// API endpoints
	APIBaseURL     = "https://workdrive.zoho.com/api/v1"
	UploadBaseURL  = "https://upload.zoho.com/workdrive-api/v1"
	DownloadBaseURL = "https://download.zoho.com/v1/workdrive"
)
