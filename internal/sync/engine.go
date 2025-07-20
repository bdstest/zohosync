// Package sync provides the core synchronization engine for ZohoSync
package sync

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bdstest/zohosync/internal/api"
	"github.com/bdstest/zohosync/internal/storage"
	"github.com/bdstest/zohosync/pkg/types"
	"github.com/bdstest/zohosync/internal/utils"
	"github.com/fsnotify/fsnotify"
)

// Engine represents the synchronization engine
type Engine struct {
	apiClient    *api.Client
	database     *storage.Database
	watcher      *fsnotify.Watcher
	config       *types.Config
	logger       *utils.Logger
	isRunning    bool
	stopChan     chan struct{}
	mu           sync.RWMutex
	syncFolders  []types.FolderConfig
}

// NewEngine creates a new synchronization engine
func NewEngine(apiClient *api.Client, database *storage.Database, config *types.Config) *Engine {
	return &Engine{
		apiClient:   apiClient,
		database:    database,
		config:      config,
		logger:      utils.GetLogger(),
		stopChan:    make(chan struct{}),
		syncFolders: config.Folders,
	}
}

// Start begins the synchronization process
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.isRunning {
		return fmt.Errorf("sync engine is already running")
	}

	// Initialize file system watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	e.watcher = watcher

	// Add folders to watch
	for _, folder := range e.syncFolders {
		if folder.Enabled {
			if err := e.addWatchRecursive(folder.Local); err != nil {
				e.logger.Errorf("Failed to watch folder %s: %v", folder.Local, err)
			} else {
				e.logger.Infof("Watching folder: %s", folder.Local)
			}
		}
	}

	e.isRunning = true
	
	// Start background goroutines
	go e.watchFileChanges(ctx)
	go e.periodicSync(ctx)

	e.logger.Info("Sync engine started successfully")
	return nil
}

// Stop stops the synchronization engine
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isRunning {
		return nil
	}

	close(e.stopChan)
	
	if e.watcher != nil {
		e.watcher.Close()
	}

	e.isRunning = false
	e.logger.Info("Sync engine stopped")
	return nil
}

// addWatchRecursive adds a directory and all its subdirectories to the watcher
func (e *Engine) addWatchRecursive(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return e.watcher.Add(path)
		}
		return nil
	})
}

// watchFileChanges monitors file system changes
func (e *Engine) watchFileChanges(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return
		case event, ok := <-e.watcher.Events:
			if !ok {
				return
			}
			e.handleFileEvent(event)
		case err, ok := <-e.watcher.Errors:
			if !ok {
				return
			}
			e.logger.Errorf("File watcher error: %v", err)
		}
	}
}

// handleFileEvent processes file system events
func (e *Engine) handleFileEvent(event fsnotify.Event) {
	e.logger.Debugf("File event: %s %s", event.Op.String(), event.Name)

	// Skip temporary files and hidden files
	if e.shouldIgnoreFile(event.Name) {
		return
	}

	// Determine operation type
	var syncRequired bool
	
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		syncRequired = true
		e.logger.Debugf("File created: %s", event.Name)
	case event.Op&fsnotify.Write == fsnotify.Write:
		syncRequired = true
		e.logger.Debugf("File modified: %s", event.Name)
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		syncRequired = true
		e.logger.Debugf("File removed: %s", event.Name)
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		syncRequired = true
		e.logger.Debugf("File renamed: %s", event.Name)
	}

	if syncRequired {
		// Queue file for synchronization
		go e.queueFileForSync(event.Name, event.Op)
	}
}

// shouldIgnoreFile determines if a file should be ignored during sync
func (e *Engine) shouldIgnoreFile(path string) bool {
	name := filepath.Base(path)
	
	// Ignore hidden files
	if strings.HasPrefix(name, ".") {
		return true
	}
	
	// Ignore temporary files
	tmpExtensions := []string{".tmp", ".temp", ".swp", ".swo", "~"}
	for _, ext := range tmpExtensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	
	// Ignore system files
	systemFiles := []string{"Thumbs.db", ".DS_Store", "desktop.ini"}
	for _, sysFile := range systemFiles {
		if name == sysFile {
			return true
		}
	}
	
	return false
}

// queueFileForSync adds a file to the sync queue
func (e *Engine) queueFileForSync(filePath string, operation fsnotify.Op) {
	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		e.logger.Errorf("Failed to get file info for %s: %v", filePath, err)
		return
	}

	// Create file metadata
	metadata := &types.FileMetadata{
		Path:        filePath,
		IsDirectory: fileInfo != nil && fileInfo.IsDir(),
		SyncStatus:  "pending",
	}

	if fileInfo != nil {
		metadata.Size = fileInfo.Size()
		metadata.ModifiedTime = fileInfo.ModTime()
		
		// Calculate hash for files (not directories)
		if !metadata.IsDirectory {
			hash, err := e.calculateFileHash(filePath)
			if err != nil {
				e.logger.Errorf("Failed to calculate hash for %s: %v", filePath, err)
			} else {
				metadata.Hash = hash
			}
		}
	}

	// Save to database
	if err := e.database.SaveFileMetadata(metadata); err != nil {
		e.logger.Errorf("Failed to save file metadata: %v", err)
	}

	e.logger.Debugf("Queued file for sync: %s", filePath)
}

// calculateFileHash calculates MD5 hash of a file
func (e *Engine) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// periodicSync performs periodic synchronization
func (e *Engine) periodicSync(ctx context.Context) {
	interval := time.Duration(e.config.Sync.Interval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return
		case <-ticker.C:
			e.performSync(ctx)
		}
	}
}

// performSync executes a synchronization cycle
func (e *Engine) performSync(ctx context.Context) {
	e.logger.Info("Starting sync cycle")
	
	// Get pending files
	pendingFiles, err := e.database.GetPendingFiles()
	if err != nil {
		e.logger.Errorf("Failed to get pending files: %v", err)
		return
	}

	if len(pendingFiles) == 0 {
		e.logger.Debug("No pending files to sync")
		return
	}

	e.logger.Infof("Found %d files to sync", len(pendingFiles))

	// Process files with limited concurrency
	maxConcurrent := e.config.Sync.MaxConcurrentSyncs
	if maxConcurrent <= 0 {
		maxConcurrent = 3
	}

	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for _, file := range pendingFiles {
		wg.Add(1)
		go func(f types.FileMetadata) {
			defer wg.Done()
			sem <- struct{}{} // Acquire
			defer func() { <-sem }() // Release

			e.syncFile(ctx, &f)
		}(file)
	}

	wg.Wait()
	e.logger.Info("Sync cycle completed")
}

// syncFile synchronizes a single file
func (e *Engine) syncFile(ctx context.Context, metadata *types.FileMetadata) {
	e.logger.Debugf("Syncing file: %s", metadata.Path)

	// Log sync operation start
	if err := e.database.LogSyncOperation(metadata.ID, "sync", "started", ""); err != nil {
		e.logger.Errorf("Failed to log sync operation: %v", err)
	}

	// Check if file exists locally
	_, err := os.Stat(metadata.Path)
	fileExists := err == nil

	var syncErr error

	switch {
	case fileExists && metadata.RemoteID == "":
		// Local file, needs upload
		syncErr = e.uploadFile(ctx, metadata)
	case !fileExists && metadata.RemoteID != "":
		// Remote file, needs download
		syncErr = e.downloadFile(ctx, metadata)
	case fileExists && metadata.RemoteID != "":
		// File exists both locally and remotely, check for conflicts
		syncErr = e.resolveConflict(ctx, metadata)
	default:
		// File doesn't exist anywhere, mark as synced
		metadata.SyncStatus = "synced"
		syncErr = e.database.SaveFileMetadata(metadata)
	}

	// Update sync status
	if syncErr != nil {
		e.logger.Errorf("Failed to sync file %s: %v", metadata.Path, syncErr)
		metadata.SyncStatus = "error"
		e.database.LogSyncOperation(metadata.ID, "sync", "failed", syncErr.Error())
	} else {
		metadata.SyncStatus = "synced"
		e.database.LogSyncOperation(metadata.ID, "sync", "success", "")
	}

	e.database.SaveFileMetadata(metadata)
}

// uploadFile uploads a local file to remote storage
func (e *Engine) uploadFile(ctx context.Context, metadata *types.FileMetadata) error {
	e.logger.Infof("Uploading file: %s", metadata.Path)

	if metadata.IsDirectory {
		// Create directory remotely
		// This is a simplified implementation - would need proper parent resolution
		folderInfo, err := e.apiClient.CreateFolder(ctx, "root", filepath.Base(metadata.Path))
		if err != nil {
			return fmt.Errorf("failed to create remote folder: %w", err)
		}
		metadata.RemoteID = folderInfo.ID
		return nil
	}

	// For files, initiate upload
	fileInfo, err := os.Stat(metadata.Path)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	uploadInfo, err := e.apiClient.InitiateUpload(ctx, filepath.Base(metadata.Path), fileInfo.Size(), "root")
	if err != nil {
		return fmt.Errorf("failed to initiate upload: %w", err)
	}

	// Upload would continue here with actual file transfer
	// This is a skeleton implementation
	e.logger.Infof("Upload initiated for %s with ID: %s", metadata.Path, uploadInfo.UploadID)
	
	return nil
}

// downloadFile downloads a remote file to local storage
func (e *Engine) downloadFile(ctx context.Context, metadata *types.FileMetadata) error {
	e.logger.Infof("Downloading file: %s", metadata.Path)

	// Get remote file info
	remoteInfo, err := e.apiClient.GetFileInfo(ctx, metadata.RemoteID)
	if err != nil {
		return fmt.Errorf("failed to get remote file info: %w", err)
	}

	if remoteInfo.IsFolder {
		// Create local directory
		return os.MkdirAll(metadata.Path, 0755)
	}

	// Download file content
	reader, err := e.apiClient.DownloadFile(ctx, metadata.RemoteID)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer reader.Close()

	// Ensure local directory exists
	if err := os.MkdirAll(filepath.Dir(metadata.Path), 0755); err != nil {
		return fmt.Errorf("failed to create local directory: %w", err)
	}

	// Create local file
	localFile, err := os.Create(metadata.Path)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer localFile.Close()

	// Copy content
	if _, err := io.Copy(localFile, reader); err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	e.logger.Infof("Downloaded file: %s", metadata.Path)
	return nil
}

// resolveConflict handles conflicts between local and remote files
func (e *Engine) resolveConflict(ctx context.Context, metadata *types.FileMetadata) error {
	e.logger.Debugf("Resolving conflict for: %s", metadata.Path)

	// Get remote file info
	remoteInfo, err := e.apiClient.GetFileInfo(ctx, metadata.RemoteID)
	if err != nil {
		return fmt.Errorf("failed to get remote file info: %w", err)
	}

	// Get local file info
	localInfo, err := os.Stat(metadata.Path)
	if err != nil {
		return fmt.Errorf("failed to get local file info: %w", err)
	}

	// Simple conflict resolution based on modification time
	switch e.config.Sync.ConflictResolution {
	case "newer":
		if localInfo.ModTime().After(remoteInfo.ModifiedTime) {
			return e.uploadFile(ctx, metadata)
		} else {
			return e.downloadFile(ctx, metadata)
		}
	case "local":
		return e.uploadFile(ctx, metadata)
	case "remote":
		return e.downloadFile(ctx, metadata)
	default:
		// Mark as conflict for manual resolution
		metadata.SyncStatus = "conflict"
		return nil
	}
}

// GetSyncStatus returns current synchronization status
func (e *Engine) GetSyncStatus() (*types.SyncStatus, error) {
	return e.database.GetSyncStats()
}

// IsRunning returns whether the sync engine is currently running
func (e *Engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isRunning
}