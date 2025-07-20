// Package gui provides graphical user interface components
package gui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/bdstest/zohosync/internal/api"
	"github.com/bdstest/zohosync/internal/auth"
	"github.com/bdstest/zohosync/internal/storage"
	"github.com/bdstest/zohosync/internal/utils"
	"github.com/bdstest/zohosync/pkg/types"
)

// AuthWindow handles authentication UI
type AuthWindow struct {
	window    fyne.Window
	config    *types.Config
	database  *storage.Database
	logger    *utils.Logger
	onSuccess func(*types.TokenInfo)
}

// NewAuthWindow creates a new authentication window
func NewAuthWindow(parent fyne.Window, config *types.Config, database *storage.Database, onSuccess func(*types.TokenInfo)) *AuthWindow {
	return &AuthWindow{
		window:    parent,
		config:    config,
		database:  database,
		logger:    utils.GetLogger(),
		onSuccess: onSuccess,
	}
}

// Show displays the authentication window
func (a *AuthWindow) Show() {
	// Check if already authenticated
	existingToken, err := a.database.GetAuthToken()
	if err == nil && existingToken != nil {
		oauthClient := auth.NewOAuthClient(a.config)
		if oauthClient.ValidateToken(existingToken) {
			a.showAlreadyAuthenticated(existingToken)
			return
		}
	}

	a.showLoginForm()
}

// showAlreadyAuthenticated displays status for already authenticated user
func (a *AuthWindow) showAlreadyAuthenticated(token *types.TokenInfo) {
	// Get user info
	apiClient := api.NewClient(token)
	userInfo, err := apiClient.GetUserInfo(context.Background())
	
	var userText string
	if err != nil {
		userText = "Authenticated user (unable to fetch details)"
	} else {
		userText = fmt.Sprintf("%s (%s)", userInfo.DisplayName, userInfo.Email)
	}

	content := container.NewVBox(
		widget.NewCard("Authentication Status", "Already authenticated",
			container.NewVBox(
				widget.NewIcon(fyne.CurrentApp().Icon()),
				widget.NewLabel("✅ You are already logged in to ZohoSync"),
				widget.NewLabel("User: "+userText),
				widget.NewLabel("Token expires: "+token.ExpiresAt.Format("2006-01-02 15:04:05")),
			),
		),
		container.NewHBox(
			widget.NewButton("Continue", func() {
				if a.onSuccess != nil {
					a.onSuccess(token)
				}
			}),
			widget.NewButton("Re-authenticate", func() {
				a.showLoginForm()
			}),
		),
	)

	dialog.ShowCustom("Authentication Status", "Close", content, a.window)
}

// showLoginForm displays the login form
func (a *AuthWindow) showLoginForm() {
	// Create login UI
	titleLabel := widget.NewLabelWithStyle("ZohoSync Authentication", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	descLabel := widget.NewLabel("Connect to your Zoho WorkDrive account to start syncing files.")
	descLabel.Wrapping = fyne.TextWrapWord

	instructionLabel := widget.NewRichTextFromMarkdown(`
**Steps to authenticate:**
1. Click "Login with Zoho WorkDrive" below
2. Your browser will open with Zoho's login page
3. Sign in with your Zoho account
4. Grant permissions to ZohoSync
5. Return to this application

**Security:** ZohoSync uses OAuth 2.0 with PKCE for secure authentication. Your credentials are never stored locally.
`)

	loginButton := widget.NewButton("🔐 Login with Zoho WorkDrive", func() {
		a.handleLogin()
	})
	loginButton.Importance = widget.HighImportance

	quitButton := widget.NewButton("Cancel", func() {
		fyne.CurrentApp().Quit()
	})

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		descLabel,
		widget.NewCard("", "", instructionLabel),
		container.NewHBox(loginButton, quitButton),
	)

	dialog.ShowCustom("ZohoSync Login", "", content, a.window)
}

// handleLogin processes the OAuth login flow
func (a *AuthWindow) handleLogin() {
	// Create OAuth client
	oauthClient := auth.NewOAuthClient(a.config)

	// Generate auth URL
	authURL, err := oauthClient.GetAuthURL()
	if err != nil {
		a.showError("Failed to generate authentication URL", err)
		return
	}

	// Show login progress dialog
	progressDialog := a.showLoginProgress(authURL)

	// Start authentication in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		token, err := oauthClient.StartCallbackServer(ctx)
		
		// Close progress dialog
		progressDialog.Hide()

		if err != nil {
			a.showError("Authentication failed", err)
			return
		}

		// Save token
		if err := a.database.SaveAuthToken(token); err != nil {
			a.showError("Failed to save authentication token", err)
			return
		}

		// Verify token by getting user info
		apiClient := api.NewClient(token)
		userInfo, err := apiClient.GetUserInfo(ctx)
		if err != nil {
			a.showError("Failed to verify authentication", err)
			return
		}

		// Show success
		a.showLoginSuccess(userInfo, token)
	}()
}

// showLoginProgress displays the login progress dialog
func (a *AuthWindow) showLoginProgress(authURL string) *dialog.CustomDialog {
	progressBar := widget.NewProgressBarInfinite()
	progressBar.Start()

	statusLabel := widget.NewLabel("🌐 Opening browser...")
	
	// Try to open browser (would need platform-specific implementation)
	// For now, just show the URL
	urlEntry := widget.NewEntry()
	urlEntry.SetText(authURL)
	urlEntry.MultiLine = true

	content := container.NewVBox(
		widget.NewCard("Authenticating with Zoho WorkDrive", "",
			container.NewVBox(
				progressBar,
				statusLabel,
				widget.NewLabel("Please visit this URL in your browser:"),
				urlEntry,
				widget.NewLabel("⏱️ Waiting for authentication callback..."),
			),
		),
	)

	progressDialog := dialog.NewCustom("Authentication in Progress", "Cancel", content, a.window)
	progressDialog.Show()

	return progressDialog
}

// showLoginSuccess displays successful login message
func (a *AuthWindow) showLoginSuccess(userInfo *api.UserInfo, token *types.TokenInfo) {
	content := container.NewVBox(
		widget.NewCard("Authentication Successful! 🎉", "",
			container.NewVBox(
				widget.NewIcon(fyne.CurrentApp().Icon()),
				widget.NewLabel("✅ Successfully connected to Zoho WorkDrive"),
				widget.NewLabel("Welcome, "+userInfo.DisplayName+"!"),
				widget.NewLabel("Email: "+userInfo.Email),
				widget.NewLabel("ZohoSync is now ready to sync your files."),
			),
		),
		widget.NewButton("Continue", func() {
			if a.onSuccess != nil {
				a.onSuccess(token)
			}
		}),
	)

	dialog.ShowCustom("Success", "", content, a.window)
	a.logger.Infof("User %s successfully authenticated", userInfo.Email)
}

// showError displays an error dialog
func (a *AuthWindow) showError(title string, err error) {
	content := container.NewVBox(
		widget.NewCard("❌ "+title, "",
			container.NewVBox(
				widget.NewLabel("An error occurred:"),
				widget.NewLabel(err.Error()),
				widget.NewLabel("Please try again or check your internet connection."),
			),
		),
		container.NewHBox(
			widget.NewButton("Retry", func() {
				a.showLoginForm()
			}),
			widget.NewButton("Cancel", func() {
				fyne.CurrentApp().Quit()
			}),
		),
	)

	dialog.ShowCustom("Error", "", content, a.window)
	a.logger.Errorf("%s: %v", title, err)
}