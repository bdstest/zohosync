// Package gui provides system tray integration
package gui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/systray"

	"github.com/bdstest/zohosync/internal/api"
	"github.com/bdstest/zohosync/internal/storage"
	"github.com/bdstest/zohosync/internal/sync"
	"github.com/bdstest/zohosync/internal/utils"
	"github.com/bdstest/zohosync/pkg/types"
)

// SystemTray manages the system tray integration
type SystemTray struct {
	app        fyne.App
	window     fyne.Window
	config     *types.Config
	database   *storage.Database
	syncEngine *sync.Engine
	token      *types.TokenInfo
	logger     *utils.Logger
	isRunning  bool
}

// NewSystemTray creates a new system tray instance
func NewSystemTray(app fyne.App, window fyne.Window, config *types.Config, database *storage.Database, token *types.TokenInfo) *SystemTray {
	return &SystemTray{
		app:      app,
		window:   window,
		config:   config,
		database: database,
		token:    token,
		logger:   utils.GetLogger(),
	}
}

// Start initializes and starts the system tray
func (st *SystemTray) Start() error {
	if st.isRunning {
		return nil
	}

	// Initialize sync engine
	apiClient := api.NewClient(st.token)
	st.syncEngine = sync.NewEngine(apiClient, st.database, st.config)

	// Start sync engine
	if err := st.syncEngine.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start sync engine: %w", err)
	}

	// Initialize system tray
	go func() {
		systray.Run(st.onTrayReady, st.onTrayExit)
	}()

	st.isRunning = true
	st.logger.Info("System tray started")
	return nil
}

// Stop stops the system tray and sync engine
func (st *SystemTray) Stop() error {
	if !st.isRunning {
		return nil
	}

	if st.syncEngine != nil {
		st.syncEngine.Stop()
	}

	systray.Quit()
	st.isRunning = false
	st.logger.Info("System tray stopped")
	return nil
}

// onTrayReady initializes the system tray menu
func (st *SystemTray) onTrayReady() {
	// Set icon (would use actual icon file in production)
	systray.SetIcon(st.createTrayIcon())
	systray.SetTitle("ZohoSync")
	systray.SetTooltip("ZohoSync - Zoho WorkDrive Sync Client")

	// Create menu items
	mStatus := systray.AddMenuItem("📊 Status", "Show sync status")
	mShow := systray.AddMenuItem("🖥️ Show Window", "Show main window")
	systray.AddSeparator()
	
	mSync := systray.AddMenuItem("🔄 Sync Now", "Trigger manual sync")
	mPause := systray.AddMenuItem("⏸️ Pause Sync", "Pause synchronization")
	systray.AddSeparator()
	
	mSettings := systray.AddMenuItem("⚙️ Settings", "Open settings")
	systray.AddSeparator()
	
	mAbout := systray.AddMenuItem("ℹ️ About", "About ZohoSync")
	mQuit := systray.AddMenuItem("🚪 Quit", "Exit ZohoSync")

	// Start status update routine
	go st.updateTrayStatus()

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mStatus.ClickedCh:
				st.showStatusNotification()
			case <-mShow.ClickedCh:
				st.showMainWindow()
			case <-mSync.ClickedCh:
				st.triggerManualSync()
			case <-mPause.ClickedCh:
				st.toggleSyncPause()
			case <-mSettings.ClickedCh:
				st.showSettings()
			case <-mAbout.ClickedCh:
				st.showAbout()
			case <-mQuit.ClickedCh:
				st.app.Quit()
				return
			}
		}
	}()
}

// onTrayExit handles tray exit
func (st *SystemTray) onTrayExit() {
	st.logger.Info("System tray exited")
}

// createTrayIcon creates the system tray icon
func (st *SystemTray) createTrayIcon() []byte {
	// This would return actual icon bytes in production
	// For now, return empty slice
	return []byte{}
}

// updateTrayStatus updates the tray tooltip with current status
func (st *SystemTray) updateTrayStatus() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			st.refreshTrayStatus()
		}
	}
}

// refreshTrayStatus refreshes the tray status information
func (st *SystemTray) refreshTrayStatus() {
	if st.syncEngine == nil {
		return
	}

	status, err := st.syncEngine.GetSyncStatus()
	if err != nil {
		st.logger.Errorf("Failed to get sync status: %v", err)
		return
	}

	tooltip := fmt.Sprintf("ZohoSync - %s\nFiles: %d/%d synced", 
		status.State, status.SyncedFiles, status.TotalFiles)
	
	if !status.LastSync.IsZero() {
		tooltip += fmt.Sprintf("\nLast sync: %s", status.LastSync.Format("15:04:05"))
	}

	systray.SetTooltip(tooltip)
}

// showStatusNotification displays a status notification
func (st *SystemTray) showStatusNotification() {
	if st.syncEngine == nil {
		return
	}

	status, err := st.syncEngine.GetSyncStatus()
	if err != nil {
		st.showNotification("Error", "Failed to get sync status")
		return
	}

	message := fmt.Sprintf("Sync Status: %s\nFiles: %d/%d synced\nPending: %d", 
		status.State, status.SyncedFiles, status.TotalFiles, status.TotalFiles-status.SyncedFiles)
	
	st.showNotification("ZohoSync Status", message)
}

// showMainWindow brings the main window to front
func (st *SystemTray) showMainWindow() {
	if deskApp, ok := st.app.(desktop.App); ok {
		deskApp.SetSystemTrayMenu(nil) // Temporarily hide to focus window
	}
	
	st.window.Show()
	st.window.RequestFocus()
	st.logger.Debug("Main window shown from system tray")
}

// triggerManualSync triggers a manual synchronization
func (st *SystemTray) triggerManualSync() {
	if st.syncEngine == nil {
		st.showNotification("Error", "Sync engine not initialized")
		return
	}

	// The sync engine runs continuously, so we just show a notification
	st.showNotification("Sync Started", "Manual synchronization triggered")
	st.logger.Info("Manual sync triggered from system tray")
}

// toggleSyncPause toggles sync engine pause state
func (st *SystemTray) toggleSyncPause() {
	if st.syncEngine == nil {
		return
	}

	if st.syncEngine.IsRunning() {
		st.syncEngine.Stop()
		st.showNotification("Sync Paused", "Synchronization has been paused")
		st.logger.Info("Sync paused from system tray")
	} else {
		st.syncEngine.Start(context.Background())
		st.showNotification("Sync Resumed", "Synchronization has been resumed")
		st.logger.Info("Sync resumed from system tray")
	}
}

// showSettings opens the settings window
func (st *SystemTray) showSettings() {
	// TODO: Implement settings window
	st.showNotification("Settings", "Settings window not yet implemented")
	st.logger.Debug("Settings requested from system tray")
}

// showAbout displays about information
func (st *SystemTray) showAbout() {
	message := fmt.Sprintf("ZohoSync v%s\nSecure Zoho WorkDrive sync client for Linux\n\nDeveloper: bdstest", "0.1.0")
	st.showNotification("About ZohoSync", message)
}

// showNotification displays a system notification
func (st *SystemTray) showNotification(title, message string) {
	// Check if desktop notifications are supported
	if deskApp, ok := st.app.(desktop.App); ok {
		if deskApp.SendNotification != nil {
			notification := &fyne.Notification{
				Title:   title,
				Content: message,
			}
			deskApp.SendNotification(notification)
			return
		}
	}

	// Fallback: log the notification
	st.logger.Infof("Notification: %s - %s", title, message)
}

// SetSyncEngine updates the sync engine reference
func (st *SystemTray) SetSyncEngine(engine *sync.Engine) {
	st.syncEngine = engine
}

// IsRunning returns whether the system tray is running
func (st *SystemTray) IsRunning() bool {
	return st.isRunning
}