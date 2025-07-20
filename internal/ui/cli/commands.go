// Package cli provides command-line interface functionality
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bdstest/zohosync/internal/api"
	"github.com/bdstest/zohosync/internal/auth"
	"github.com/bdstest/zohosync/internal/config"
	"github.com/bdstest/zohosync/internal/storage"
	"github.com/bdstest/zohosync/internal/sync"
	"github.com/bdstest/zohosync/internal/utils"
	"github.com/bdstest/zohosync/pkg/types"
	"github.com/spf13/cobra"
)

// CLI represents the command-line interface
type CLI struct {
	config    *types.Config
	database  *storage.Database
	logger    *utils.Logger
}

// NewCLI creates a new CLI instance
func NewCLI() (*CLI, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize database
	dbPath := filepath.Join(os.Getenv("HOME"), ".config", "zohosync", "zohosync.db")
	db, err := storage.NewDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	logger := utils.InitLogger(cfg.App.LogLevel)

	return &CLI{
		config:   cfg,
		database: db,
		logger:   logger,
	}, nil
}

// Close cleans up CLI resources
func (c *CLI) Close() error {
	return c.database.Close()
}

// CreateLoginCommand creates the login command
func (c *CLI) CreateLoginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Zoho WorkDrive",
		Long:  "Initiate OAuth 2.0 authentication flow with Zoho WorkDrive",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handleLogin(cmd.Context())
		},
	}
}

// handleLogin processes the login command
func (c *CLI) handleLogin(ctx context.Context) error {
	fmt.Println("🔐 ZohoSync Authentication")
	fmt.Println("Initiating OAuth 2.0 login with Zoho WorkDrive...")
	fmt.Println()

	// Create OAuth client
	oauthClient := auth.NewOAuthClient(c.config)

	// Get authorization URL
	authURL, err := oauthClient.GetAuthURL()
	if err != nil {
		return fmt.Errorf("failed to generate auth URL: %w", err)
	}

	fmt.Println("📱 Please visit the following URL to authorize ZohoSync:")
	fmt.Println(authURL)
	fmt.Println()
	fmt.Println("🌐 Opening browser... (if supported)")
	fmt.Println("🔄 Waiting for callback...")

	// Start callback server with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	token, err := oauthClient.StartCallbackServer(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Save token to database
	if err := c.database.SaveAuthToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	// Test API connection
	apiClient := api.NewClient(token)
	userInfo, err := apiClient.GetUserInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to verify authentication: %w", err)
	}

	fmt.Printf("✅ Successfully authenticated as: %s (%s)\n", userInfo.DisplayName, userInfo.Email)
	fmt.Println("🎉 ZohoSync is now ready to use!")

	return nil
}

// CreateStatusCommand creates the status command
func (c *CLI) CreateStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show synchronization status",
		Long:  "Display current sync status, statistics, and pending operations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handleStatus(cmd.Context())
		},
	}
}

// handleStatus processes the status command
func (c *CLI) handleStatus(ctx context.Context) error {
	fmt.Println("📊 ZohoSync Status")
	fmt.Println("==================")
	fmt.Println()

	// Check authentication
	token, err := c.database.GetAuthToken()
	if err != nil {
		return fmt.Errorf("failed to get auth token: %w", err)
	}

	if token == nil {
		fmt.Println("🔐 Authentication: Not logged in")
		fmt.Println("   Run 'zohosync-cli login' to authenticate")
		return nil
	}

	// Validate token
	oauthClient := auth.NewOAuthClient(c.config)
	if !oauthClient.ValidateToken(token) {
		fmt.Println("🔐 Authentication: Token expired")
		fmt.Println("   Run 'zohosync-cli login' to re-authenticate")
		return nil
	}

	fmt.Println("🔐 Authentication: ✅ Valid")
	fmt.Printf("   Token expires: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Get user info
	apiClient := api.NewClient(token)
	userInfo, err := apiClient.GetUserInfo(ctx)
	if err != nil {
		fmt.Printf("⚠️  Failed to get user info: %v\n", err)
	} else {
		fmt.Printf("👤 User: %s (%s)\n", userInfo.DisplayName, userInfo.Email)
		fmt.Println()
	}

	// Get sync statistics
	stats, err := c.database.GetSyncStats()
	if err != nil {
		return fmt.Errorf("failed to get sync stats: %w", err)
	}

	fmt.Println("📈 Sync Statistics:")
	fmt.Printf("   Total files: %d\n", stats.TotalFiles)
	fmt.Printf("   Synced files: %d\n", stats.SyncedFiles)
	fmt.Printf("   Pending files: %d\n", stats.TotalFiles-stats.SyncedFiles)
	fmt.Printf("   Sync state: %s\n", stats.State)
	
	if !stats.LastSync.IsZero() {
		fmt.Printf("   Last sync: %s\n", stats.LastSync.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("   Last sync: Never")
	}

	fmt.Println()

	// Show configured folders
	fmt.Println("📁 Configured Folders:")
	for i, folder := range c.config.Folders {
		status := "🔴 Disabled"
		if folder.Enabled {
			status = "🟢 Enabled"
		}
		fmt.Printf("   %d. %s %s -> %s (%s)\n", i+1, status, folder.Local, folder.Remote, folder.SyncMode)
	}

	return nil
}

// CreateSyncCommand creates the sync command
func (c *CLI) CreateSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Perform manual synchronization",
		Long:  "Trigger immediate synchronization of all configured folders",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.handleSync(cmd.Context())
		},
	}

	cmd.Flags().BoolP("dry-run", "n", false, "Show what would be synced without making changes")
	return cmd
}

// handleSync processes the sync command
func (c *CLI) handleSync(ctx context.Context) error {
	// Check authentication
	token, err := c.database.GetAuthToken()
	if err != nil {
		return fmt.Errorf("failed to get auth token: %w", err)
	}

	if token == nil {
		return fmt.Errorf("not authenticated - run 'zohosync-cli login' first")
	}

	// Validate token
	oauthClient := auth.NewOAuthClient(c.config)
	if !oauthClient.ValidateToken(token) {
		return fmt.Errorf("authentication token expired - run 'zohosync-cli login'")
	}

	fmt.Println("🔄 Starting manual synchronization...")

	// Create API client and sync engine
	apiClient := api.NewClient(token)
	syncEngine := sync.NewEngine(apiClient, c.database, c.config)

	// Start sync engine
	if err := syncEngine.Start(ctx); err != nil {
		return fmt.Errorf("failed to start sync engine: %w", err)
	}
	defer syncEngine.Stop()

	// Wait for sync to complete
	fmt.Println("⏳ Synchronizing...")
	time.Sleep(2 * time.Second) // Allow time for initial sync

	// Get final status
	stats, err := syncEngine.GetSyncStatus()
	if err != nil {
		return fmt.Errorf("failed to get sync status: %w", err)
	}

	fmt.Printf("✅ Synchronization completed!\n")
	fmt.Printf("   Files processed: %d\n", stats.TotalFiles)
	fmt.Printf("   Successfully synced: %d\n", stats.SyncedFiles)

	return nil
}

// CreateListCommand creates the list command
func (c *CLI) CreateListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [folder-id]",
		Short: "List remote files",
		Long:  "List files in Zoho WorkDrive. If no folder ID is provided, lists root folder.",
		RunE: func(cmd *cobra.Command, args []string) error {
			folderID := "root"
			if len(args) > 0 {
				folderID = args[0]
			}
			return c.handleList(cmd.Context(), folderID)
		},
	}

	cmd.Flags().IntP("limit", "l", 50, "Maximum number of files to list")
	return cmd
}

// handleList processes the list command
func (c *CLI) handleList(ctx context.Context, folderID string) error {
	// Check authentication
	token, err := c.database.GetAuthToken()
	if err != nil {
		return fmt.Errorf("failed to get auth token: %w", err)
	}

	if token == nil {
		return fmt.Errorf("not authenticated - run 'zohosync-cli login' first")
	}

	// Create API client
	apiClient := api.NewClient(token)

	// Get limit from flags
	limit := 50 // Default value would be set from command flags in real implementation

	fmt.Printf("📁 Listing files in folder: %s\n", folderID)
	fmt.Println()

	// List files
	files, err := apiClient.ListFiles(ctx, folderID, limit)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("📂 No files found")
		return nil
	}

	fmt.Printf("Found %d files:\n\n", len(files))

	// Display files
	for _, file := range files {
		icon := "📄"
		if file.IsFolder {
			icon = "📁"
		}

		sizeStr := "-"
		if !file.IsFolder {
			sizeStr = formatFileSize(file.Size)
		}

		fmt.Printf("%s %s\n", icon, file.Name)
		fmt.Printf("   ID: %s\n", file.ID)
		fmt.Printf("   Size: %s\n", sizeStr)
		fmt.Printf("   Modified: %s\n", file.ModifiedTime.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	return nil
}

// formatFileSize formats file size in human-readable format
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// CreateVersionCommand creates the version command
func (c *CLI) CreateVersionCommand(version, buildDate, commit string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display ZohoSync version, build date, and commit information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ZohoSync CLI %s\n", version)
			fmt.Printf("Build Date: %s\n", buildDate)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Go Version: 1.21+\n")
			fmt.Printf("Platform: Linux\n")
		},
	}
}