package types

import "time"

// SyncStatus represents the synchronization status
type SyncStatus struct {
	State        SyncState     `json:"state"`
	LastSync     time.Time     `json:"last_sync"`
	NextSync     time.Time     `json:"next_sync"`
	InProgress   bool          `json:"in_progress"`
	TotalFiles   int           `json:"total_files"`
	SyncedFiles  int           `json:"synced_files"`
	Errors       []SyncError   `json:"errors,omitempty"`
}

// SyncState represents the current sync state
type SyncState string

const (
	SyncStateIdle     SyncState = "idle"
	SyncStateSyncing  SyncState = "syncing"
	SyncStatePaused   SyncState = "paused"
	SyncStateError    SyncState = "error"
)

// SyncError represents a synchronization error
type SyncError struct {
	Path      string    `json:"path"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// FileMetadata represents file metadata for sync tracking
type FileMetadata struct {
	ID           string    `json:"id"`
	Path         string    `json:"path"`
	RemoteID     string    `json:"remote_id"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modified_time"`
	Hash         string    `json:"hash"`
	IsDirectory  bool      `json:"is_directory"`
	SyncStatus   string    `json:"sync_status"`
}
