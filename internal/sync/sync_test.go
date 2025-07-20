package sync

import (
	"testing"
	"time"
	"context"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncEngine(t *testing.T) {
	// Mock WorkDrive API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/files":
			files := []map[string]interface{}{
				{
					"id":           "file1",
					"name":         "test.txt",
					"type":         "file",
					"size":         1024,
					"modified_time": time.Now().Unix(),
				},
				{
					"id":           "file2", 
					"name":         "document.pdf",
					"type":         "file",
					"size":         2048,
					"modified_time": time.Now().Unix(),
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": files,
			})
		case "/api/v1/download/file1":
			w.Write([]byte("test file content"))
		case "/api/v1/download/file2":
			w.Write([]byte("pdf file content"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	engine := &Engine{
		APIBaseURL: server.URL,
		LocalPath:  "/tmp/zohosync-test",
	}

	ctx := context.Background()
	
	// Test file listing
	files, err := engine.ListRemoteFiles(ctx)
	require.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Equal(t, "test.txt", files[0].Name)
	assert.Equal(t, "document.pdf", files[1].Name)
}

func TestSyncConflictResolution(t *testing.T) {
	tests := []struct {
		name           string
		localModified  time.Time
		remoteModified time.Time
		expectedAction string
	}{
		{
			name:           "Remote newer",
			localModified:  time.Now().Add(-time.Hour),
			remoteModified: time.Now(),
			expectedAction: "download",
		},
		{
			name:           "Local newer",
			localModified:  time.Now(),
			remoteModified: time.Now().Add(-time.Hour),
			expectedAction: "upload",
		},
		{
			name:           "Same timestamp",
			localModified:  time.Now(),
			remoteModified: time.Now(),
			expectedAction: "skip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localFile := &FileInfo{
				Name:         "test.txt",
				ModifiedTime: tt.localModified,
			}
			remoteFile := &FileInfo{
				Name:         "test.txt", 
				ModifiedTime: tt.remoteModified,
			}

			action := ResolveConflict(localFile, remoteFile)
			assert.Equal(t, tt.expectedAction, action)
		})
	}
}

func TestSyncProgress(t *testing.T) {
	progress := NewSyncProgress()
	
	// Test initial state
	assert.Equal(t, 0, progress.TotalFiles)
	assert.Equal(t, 0, progress.CompletedFiles)
	assert.Equal(t, float64(0), progress.Percentage())

	// Test progress updates
	progress.SetTotal(10)
	assert.Equal(t, 10, progress.TotalFiles)
	assert.Equal(t, float64(0), progress.Percentage())

	progress.IncrementCompleted()
	progress.IncrementCompleted()
	assert.Equal(t, 2, progress.CompletedFiles)
	assert.Equal(t, float64(20), progress.Percentage())

	progress.SetCompleted(10)
	assert.Equal(t, float64(100), progress.Percentage())
}

func TestFileHashing(t *testing.T) {
	testContent := []byte("test file content for hashing")
	
	hash1 := CalculateFileHash(testContent)
	hash2 := CalculateFileHash(testContent)
	
	// Same content should produce same hash
	assert.Equal(t, hash1, hash2)
	assert.NotEmpty(t, hash1)
	
	// Different content should produce different hash
	differentContent := []byte("different content")
	hash3 := CalculateFileHash(differentContent)
	assert.NotEqual(t, hash1, hash3)
}

func TestSyncErrorHandling(t *testing.T) {
	// Test network error handling
	engine := &Engine{
		APIBaseURL: "http://invalid-url-that-does-not-exist",
		LocalPath:  "/tmp/zohosync-test",
	}

	ctx := context.Background()
	_, err := engine.ListRemoteFiles(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network")

	// Test timeout handling
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Simulate slow response
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	engine.APIBaseURL = slowServer.URL
	engine.Timeout = 1 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err = engine.ListRemoteFiles(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}