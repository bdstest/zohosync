// Progress Tracker for ZohoSync
// Tracks synchronization progress and provides real-time updates
// Author: bdstest

package sync

import (
	"fmt"
	"sync"
	"time"
)

// ProgressTracker tracks sync operation progress
type ProgressTracker struct {
	totalFiles      int64
	completedFiles  int64
	totalBytes      int64
	transferredBytes int64
	startTime       time.Time
	currentFile     string
	errors          []error
	mutex           sync.RWMutex
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		startTime: time.Now(),
		errors:    make([]error, 0),
	}
}

// SetTotals sets the total files and bytes to be processed
func (pt *ProgressTracker) SetTotals(files int64, bytes int64) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	
	pt.totalFiles = files
	pt.totalBytes = bytes
}

// UpdateFileProgress updates progress for a specific file
func (pt *ProgressTracker) UpdateFileProgress(filename string, bytesTransferred int64) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	
	pt.currentFile = filename
	pt.transferredBytes += bytesTransferred
}

// CompleteFile marks a file as completed
func (pt *ProgressTracker) CompleteFile(filename string) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	
	pt.completedFiles++
	pt.currentFile = ""
}

// AddError adds an error to the progress tracker
func (pt *ProgressTracker) AddError(err error) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	
	pt.errors = append(pt.errors, err)
}

// GetProgress returns current progress information
func (pt *ProgressTracker) GetProgress() ProgressInfo {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()
	
	elapsed := time.Since(pt.startTime)
	
	var fileProgress float64
	if pt.totalFiles > 0 {
		fileProgress = float64(pt.completedFiles) / float64(pt.totalFiles) * 100
	}
	
	var byteProgress float64
	if pt.totalBytes > 0 {
		byteProgress = float64(pt.transferredBytes) / float64(pt.totalBytes) * 100
	}
	
	var speed float64
	if elapsed.Seconds() > 0 {
		speed = float64(pt.transferredBytes) / elapsed.Seconds()
	}
	
	var eta time.Duration
	if speed > 0 && pt.totalBytes > pt.transferredBytes {
		remainingBytes := pt.totalBytes - pt.transferredBytes
		eta = time.Duration(float64(remainingBytes)/speed) * time.Second
	}
	
	return ProgressInfo{
		TotalFiles:       pt.totalFiles,
		CompletedFiles:   pt.completedFiles,
		TotalBytes:       pt.totalBytes,
		TransferredBytes: pt.transferredBytes,
		FileProgress:     fileProgress,
		ByteProgress:     byteProgress,
		CurrentFile:      pt.currentFile,
		ElapsedTime:      elapsed,
		Speed:            speed,
		ETA:              eta,
		ErrorCount:       int64(len(pt.errors)),
	}
}

// GetErrors returns all errors encountered
func (pt *ProgressTracker) GetErrors() []error {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()
	
	return append([]error(nil), pt.errors...)
}

// ProgressInfo contains current progress information
type ProgressInfo struct {
	TotalFiles       int64
	CompletedFiles   int64
	TotalBytes       int64
	TransferredBytes int64
	FileProgress     float64 // Percentage
	ByteProgress     float64 // Percentage
	CurrentFile      string
	ElapsedTime      time.Duration
	Speed            float64 // Bytes per second
	ETA              time.Duration
	ErrorCount       int64
}

// String returns a formatted progress string
func (pi ProgressInfo) String() string {
	return fmt.Sprintf(
		"Progress: %.1f%% files (%d/%d), %.1f%% bytes (%s/%s), Speed: %s/s, ETA: %v, Errors: %d",
		pi.FileProgress,
		pi.CompletedFiles,
		pi.TotalFiles,
		pi.ByteProgress,
		formatBytes(pi.TransferredBytes),
		formatBytes(pi.TotalBytes),
		formatBytes(int64(pi.Speed)),
		pi.ETA.Round(time.Second),
		pi.ErrorCount,
	)
}

// formatBytes formats bytes in human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ProgressCallback is a function type for progress notifications
type ProgressCallback func(ProgressInfo)

// ProgressNotifier manages progress notifications
type ProgressNotifier struct {
	callbacks []ProgressCallback
	interval  time.Duration
	ticker    *time.Ticker
	tracker   *ProgressTracker
	stopChan  chan struct{}
}

// NewProgressNotifier creates a new progress notifier
func NewProgressNotifier(tracker *ProgressTracker, interval time.Duration) *ProgressNotifier {
	return &ProgressNotifier{
		callbacks: make([]ProgressCallback, 0),
		interval:  interval,
		tracker:   tracker,
		stopChan:  make(chan struct{}),
	}
}

// AddCallback adds a progress callback
func (pn *ProgressNotifier) AddCallback(callback ProgressCallback) {
	pn.callbacks = append(pn.callbacks, callback)
}

// Start starts the progress notifier
func (pn *ProgressNotifier) Start() {
	pn.ticker = time.NewTicker(pn.interval)
	
	go func() {
		for {
			select {
			case <-pn.ticker.C:
				progress := pn.tracker.GetProgress()
				for _, callback := range pn.callbacks {
					callback(progress)
				}
			case <-pn.stopChan:
				return
			}
		}
	}()
}

// Stop stops the progress notifier
func (pn *ProgressNotifier) Stop() {
	if pn.ticker != nil {
		pn.ticker.Stop()
	}
	close(pn.stopChan)
}