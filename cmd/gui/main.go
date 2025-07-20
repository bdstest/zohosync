// ZohoSync GUI - Desktop application for ZohoSync
package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	
	"github.com/bdstest/zohosync/internal/auth"
	"github.com/bdstest/zohosync/internal/config"
	"github.com/bdstest/zohosync/internal/storage"
	"github.com/bdstest/zohosync/internal/ui/gui"
	"github.com/bdstest/zohosync/internal/utils"
	"github.com/bdstest/zohosync/pkg/types"
)

var (
	version   = "dev"
	buildDate = "unknown"
	commit    = "unknown"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
	}
	
	// Initialize logger
	logger := utils.InitLogger(cfg.App.LogLevel)
	logger.Info("Starting ZohoSync GUI")
	
	// Create Fyne application
	myApp := app.New()
	myApp.Settings().SetTheme(&zohoTheme{})
	
	// Create main window
	myWindow := myApp.NewWindow("ZohoSync")
	myWindow.Resize(fyne.NewSize(800, 600))
	
	// Initialize database
	dbPath := filepath.Join(os.Getenv("HOME"), ".config", "zohosync", "zohosync.db")
	database, err := storage.NewDatabase(dbPath)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Check authentication status
	token, err := database.GetAuthToken()
	if err != nil {
		logger.Errorf("Failed to check auth status: %v", err)
	}

	// Create OAuth client for token validation
	oauthClient := auth.NewOAuthClient(cfg)
	isAuthenticated := token != nil && oauthClient.ValidateToken(token)

	if !isAuthenticated {
		// Show authentication window
		authWindow := gui.NewAuthWindow(myWindow, cfg, database, func(newToken *types.TokenInfo) {
			logger.Info("Authentication successful, starting main application")
			showMainWindow(myWindow, cfg, database, newToken)
		})
		authWindow.Show()
	} else {
		// User is already authenticated, show main window
		showMainWindow(myWindow, cfg, database, token)
	}

	myWindow.ShowAndRun()
}

// showMainWindow displays the main application window
func showMainWindow(window fyne.Window, config *types.Config, database *storage.Database, token *types.TokenInfo) {
	// Create main UI
	welcomeLabel := widget.NewLabelWithStyle("ZohoSync", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	
	// Status card
	statusCard := widget.NewCard("Sync Status", "", 
		container.NewVBox(
			widget.NewLabel("✅ Connected to Zoho WorkDrive"),
			widget.NewLabel("🔄 Monitoring for changes..."),
		),
	)

	// Quick actions
	syncButton := widget.NewButton("🔄 Sync Now", func() {
		// TODO: Implement manual sync
		statusCard.SetContent(container.NewVBox(
			widget.NewLabel("✅ Connected to Zoho WorkDrive"),
			widget.NewLabel("⏳ Syncing..."),
		))
	})

	settingsButton := widget.NewButton("⚙️ Settings", func() {
		// TODO: Implement settings window
	})

	logoutButton := widget.NewButton("🚪 Logout", func() {
		// Clear stored token
		database.SaveAuthToken(nil)
		fyne.CurrentApp().Quit()
	})

	// Layout
	content := container.NewVBox(
		welcomeLabel,
		widget.NewSeparator(),
		statusCard,
		container.NewHBox(syncButton, settingsButton),
		widget.NewSeparator(),
		logoutButton,
	)

	window.SetContent(content)
}

// Basic theme placeholder
type zohoTheme struct{}

func (z zohoTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

func (z zohoTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (z zohoTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (z zohoTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
