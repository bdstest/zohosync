// Conflict Handler for ZohoSync
// Handles file synchronization conflicts with various resolution strategies
// Author: bdstest

package sync

import (
	"fmt"
	"log"
	"path/filepath"
	"time"
)

// ConflictHandler manages file conflict resolution
type ConflictHandler struct {
	strategy ConflictResolution
}

// NewConflictHandler creates a new conflict handler
func NewConflictHandler(strategy ConflictResolution) *ConflictHandler {
	return &ConflictHandler{
		strategy: strategy,
	}
}

// ResolveConflict determines how to resolve a file conflict
func (ch *ConflictHandler) ResolveConflict(path string, local, remote *FileMetadata) SyncOperation {
	log.Printf("Conflict detected for %s: local(%s) vs remote(%s)", 
		path, local.ModTime.Format(time.RFC3339), remote.ModTime.Format(time.RFC3339))
	
	switch ch.strategy {
	case ResolutionNewest:
		return ch.resolveByNewest(path, local, remote)
	case ResolutionLargest:
		return ch.resolveByLargest(path, local, remote)
	case ResolutionKeepBoth:
		return ch.resolveKeepBoth(path, local, remote)
	case ResolutionManual:
		return ch.resolveManual(path, local, remote)
	default:
		return SyncOperation{Type: OperationSkip}
	}
}

// resolveByNewest keeps the newest file
func (ch *ConflictHandler) resolveByNewest(path string, local, remote *FileMetadata) SyncOperation {
	if local.ModTime.After(remote.ModTime) {
		log.Printf("Resolving conflict: local file is newer, uploading %s", path)
		return SyncOperation{
			Type:       OperationUpload,
			LocalPath:  path,
			RemotePath: path,
			FileSize:   local.Size,
			Metadata:   local,
		}
	} else {
		log.Printf("Resolving conflict: remote file is newer, downloading %s", path)
		return SyncOperation{
			Type:       OperationDownload,
			LocalPath:  path,
			RemotePath: path,
			FileSize:   remote.Size,
			Metadata:   remote,
		}
	}
}

// resolveByLargest keeps the largest file
func (ch *ConflictHandler) resolveByLargest(path string, local, remote *FileMetadata) SyncOperation {
	if local.Size > remote.Size {
		log.Printf("Resolving conflict: local file is larger, uploading %s", path)
		return SyncOperation{
			Type:       OperationUpload,
			LocalPath:  path,
			RemotePath: path,
			FileSize:   local.Size,
			Metadata:   local,
		}
	} else {
		log.Printf("Resolving conflict: remote file is larger, downloading %s", path)
		return SyncOperation{
			Type:       OperationDownload,
			LocalPath:  path,
			RemotePath: path,
			FileSize:   remote.Size,
			Metadata:   remote,
		}
	}
}

// resolveKeepBoth keeps both files with different names
func (ch *ConflictHandler) resolveKeepBoth(path string, local, remote *FileMetadata) SyncOperation {
	timestamp := time.Now().Format("20060102_150405")
	
	// Create conflict filename for local file
	dir := filepath.Dir(path)
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	nameWithoutExt := filename[:len(filename)-len(ext)]
	
	conflictName := fmt.Sprintf("%s_conflict_local_%s%s", nameWithoutExt, timestamp, ext)
	conflictPath := filepath.Join(dir, conflictName)
	
	log.Printf("Resolving conflict: keeping both files, local as %s", conflictPath)
	
	// Download remote file to original path
	// Local file will be renamed to conflict path (handled separately)
	return SyncOperation{
		Type:       OperationDownload,
		LocalPath:  path,
		RemotePath: path,
		FileSize:   remote.Size,
		Metadata:   remote,
	}
}

// resolveManual marks conflict for manual resolution
func (ch *ConflictHandler) resolveManual(path string, local, remote *FileMetadata) SyncOperation {
	log.Printf("Conflict marked for manual resolution: %s", path)
	
	return SyncOperation{
		Type:       OperationConflict,
		LocalPath:  path,
		RemotePath: path,
		FileSize:   0,
		Metadata:   local,
	}
}

// ConflictInfo represents information about a conflict
type ConflictInfo struct {
	Path         string
	LocalFile    *FileMetadata
	RemoteFile   *FileMetadata
	Timestamp    time.Time
	Resolution   string
	AutoResolved bool
}

// ConflictLog maintains a log of conflicts for review
type ConflictLog struct {
	conflicts []ConflictInfo
}

// NewConflictLog creates a new conflict log
func NewConflictLog() *ConflictLog {
	return &ConflictLog{
		conflicts: make([]ConflictInfo, 0),
	}
}

// LogConflict adds a conflict to the log
func (cl *ConflictLog) LogConflict(info ConflictInfo) {
	cl.conflicts = append(cl.conflicts, info)
}

// GetConflicts returns all logged conflicts
func (cl *ConflictLog) GetConflicts() []ConflictInfo {
	return cl.conflicts
}

// GetUnresolvedConflicts returns conflicts that need manual resolution
func (cl *ConflictLog) GetUnresolvedConflicts() []ConflictInfo {
	var unresolved []ConflictInfo
	for _, conflict := range cl.conflicts {
		if !conflict.AutoResolved {
			unresolved = append(unresolved, conflict)
		}
	}
	return unresolved
}