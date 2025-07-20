// Package storage handles local data persistence for ZohoSync
package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bdstest/zohosync/pkg/types"
	"github.com/bdstest/zohosync/internal/utils"
	_ "github.com/mattn/go-sqlite3"
)

// Database represents the local SQLite database
type Database struct {
	db     *sql.DB
	logger *utils.Logger
}

// NewDatabase creates a new database connection
func NewDatabase(dbPath string) (*Database, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal=WAL&_timeout=10000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{
		db:     db,
		logger: utils.GetLogger(),
	}

	if err := database.initialize(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return database, nil
}

// initialize creates the database schema
func (d *Database) initialize() error {
	schema := `
	-- Files table for tracking local and remote file state
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		local_path TEXT NOT NULL UNIQUE,
		remote_id TEXT,
		remote_path TEXT,
		size INTEGER DEFAULT 0,
		modified_time DATETIME,
		hash TEXT,
		is_directory BOOLEAN DEFAULT FALSE,
		sync_status TEXT DEFAULT 'pending',
		last_sync DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Sync operations table for tracking sync history
	CREATE TABLE IF NOT EXISTS sync_operations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id INTEGER,
		operation_type TEXT NOT NULL, -- upload, download, delete, conflict
		status TEXT DEFAULT 'pending', -- pending, success, failed
		error_message TEXT,
		started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME,
		FOREIGN KEY (file_id) REFERENCES files(id)
	);

	-- Configuration table for storing app settings
	CREATE TABLE IF NOT EXISTS config (
		key TEXT PRIMARY KEY,
		value TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Authentication tokens table
	CREATE TABLE IF NOT EXISTS auth_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		access_token TEXT,
		refresh_token TEXT,
		token_type TEXT DEFAULT 'Bearer',
		expires_at DATETIME,
		scope TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Create indexes for better performance
	CREATE INDEX IF NOT EXISTS idx_files_local_path ON files(local_path);
	CREATE INDEX IF NOT EXISTS idx_files_remote_id ON files(remote_id);
	CREATE INDEX IF NOT EXISTS idx_files_sync_status ON files(sync_status);
	CREATE INDEX IF NOT EXISTS idx_sync_operations_file_id ON sync_operations(file_id);
	CREATE INDEX IF NOT EXISTS idx_sync_operations_status ON sync_operations(status);
	`

	if _, err := d.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	d.logger.Info("Database initialized successfully")
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// SaveFileMetadata saves or updates file metadata
func (d *Database) SaveFileMetadata(metadata *types.FileMetadata) error {
	query := `
	INSERT OR REPLACE INTO files 
	(local_path, remote_id, remote_path, size, modified_time, hash, is_directory, sync_status, last_sync, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := d.db.Exec(query,
		metadata.Path,
		metadata.RemoteID,
		metadata.Path, // Assuming same path structure
		metadata.Size,
		metadata.ModifiedTime,
		metadata.Hash,
		metadata.IsDirectory,
		metadata.SyncStatus,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to save file metadata: %w", err)
	}

	d.logger.Debugf("Saved metadata for file: %s", metadata.Path)
	return nil
}

// GetFileMetadata retrieves file metadata by local path
func (d *Database) GetFileMetadata(localPath string) (*types.FileMetadata, error) {
	query := `
	SELECT id, local_path, remote_id, size, modified_time, hash, is_directory, sync_status
	FROM files WHERE local_path = ?
	`

	row := d.db.QueryRow(query, localPath)
	
	var metadata types.FileMetadata
	var id int
	var modifiedTime time.Time

	err := row.Scan(
		&id,
		&metadata.Path,
		&metadata.RemoteID,
		&metadata.Size,
		&modifiedTime,
		&metadata.Hash,
		&metadata.IsDirectory,
		&metadata.SyncStatus,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // File not found
		}
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	metadata.ID = fmt.Sprintf("%d", id)
	metadata.ModifiedTime = modifiedTime

	return &metadata, nil
}

// GetPendingFiles retrieves files that need synchronization
func (d *Database) GetPendingFiles() ([]types.FileMetadata, error) {
	query := `
	SELECT id, local_path, remote_id, size, modified_time, hash, is_directory, sync_status
	FROM files WHERE sync_status IN ('pending', 'conflict', 'error')
	ORDER BY modified_time DESC
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending files: %w", err)
	}
	defer rows.Close()

	var files []types.FileMetadata
	for rows.Next() {
		var metadata types.FileMetadata
		var id int
		var modifiedTime time.Time

		err := rows.Scan(
			&id,
			&metadata.Path,
			&metadata.RemoteID,
			&metadata.Size,
			&modifiedTime,
			&metadata.Hash,
			&metadata.IsDirectory,
			&metadata.SyncStatus,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan file row: %w", err)
		}

		metadata.ID = fmt.Sprintf("%d", id)
		metadata.ModifiedTime = modifiedTime
		files = append(files, metadata)
	}

	return files, nil
}

// LogSyncOperation records a sync operation
func (d *Database) LogSyncOperation(fileID, operationType, status, errorMessage string) error {
	query := `
	INSERT INTO sync_operations (file_id, operation_type, status, error_message, started_at)
	VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	_, err := d.db.Exec(query, fileID, operationType, status, errorMessage)
	if err != nil {
		return fmt.Errorf("failed to log sync operation: %w", err)
	}

	return nil
}

// CompleteSyncOperation marks a sync operation as completed
func (d *Database) CompleteSyncOperation(operationID int, status, errorMessage string) error {
	query := `
	UPDATE sync_operations 
	SET status = ?, error_message = ?, completed_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`

	_, err := d.db.Exec(query, status, errorMessage, operationID)
	if err != nil {
		return fmt.Errorf("failed to complete sync operation: %w", err)
	}

	return nil
}

// SaveAuthToken saves authentication token to database
func (d *Database) SaveAuthToken(token *types.TokenInfo) error {
	// Delete existing tokens
	if _, err := d.db.Exec("DELETE FROM auth_tokens"); err != nil {
		return fmt.Errorf("failed to clear existing tokens: %w", err)
	}

	query := `
	INSERT INTO auth_tokens (access_token, refresh_token, token_type, expires_at, scope, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	_, err := d.db.Exec(query,
		token.AccessToken,
		token.RefreshToken,
		token.TokenType,
		token.ExpiresAt,
		token.Scope,
	)

	if err != nil {
		return fmt.Errorf("failed to save auth token: %w", err)
	}

	d.logger.Info("Authentication token saved to database")
	return nil
}

// GetAuthToken retrieves the stored authentication token
func (d *Database) GetAuthToken() (*types.TokenInfo, error) {
	query := `
	SELECT access_token, refresh_token, token_type, expires_at, scope
	FROM auth_tokens ORDER BY created_at DESC LIMIT 1
	`

	row := d.db.QueryRow(query)
	
	var token types.TokenInfo
	var expiresAt time.Time

	err := row.Scan(
		&token.AccessToken,
		&token.RefreshToken,
		&token.TokenType,
		&expiresAt,
		&token.Scope,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No token found
		}
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	token.ExpiresAt = expiresAt
	token.ExpiresIn = int(time.Until(expiresAt).Seconds())

	return &token, nil
}

// GetSyncStats retrieves synchronization statistics
func (d *Database) GetSyncStats() (*types.SyncStatus, error) {
	query := `
	SELECT 
		COUNT(*) as total_files,
		COUNT(CASE WHEN sync_status = 'synced' THEN 1 END) as synced_files,
		MAX(last_sync) as last_sync
	FROM files
	`

	row := d.db.QueryRow(query)
	
	var totalFiles, syncedFiles int
	var lastSyncPtr *time.Time

	err := row.Scan(&totalFiles, &syncedFiles, &lastSyncPtr)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync stats: %w", err)
	}

	status := &types.SyncStatus{
		State:       types.SyncStateIdle,
		TotalFiles:  totalFiles,
		SyncedFiles: syncedFiles,
		InProgress:  false,
	}

	if lastSyncPtr != nil {
		status.LastSync = *lastSyncPtr
	}

	return status, nil
}

// SetConfigValue stores a configuration value
func (d *Database) SetConfigValue(key, value string) error {
	query := `
	INSERT OR REPLACE INTO config (key, value, updated_at)
	VALUES (?, ?, CURRENT_TIMESTAMP)
	`

	_, err := d.db.Exec(query, key, value)
	if err != nil {
		return fmt.Errorf("failed to set config value: %w", err)
	}

	return nil
}

// GetConfigValue retrieves a configuration value
func (d *Database) GetConfigValue(key string) (string, error) {
	query := "SELECT value FROM config WHERE key = ?"
	
	var value string
	err := d.db.QueryRow(query, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // Key not found
		}
		return "", fmt.Errorf("failed to get config value: %w", err)
	}

	return value, nil
}