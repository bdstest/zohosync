package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// MockFile represents a file in the mock WorkDrive
type MockFile struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"` // "file" or "folder"
	Size       int64     `json:"size,omitempty"`
	ParentID   string    `json:"parent_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
	DownloadURL string   `json:"download_url,omitempty"`
}

// MockAPI represents the mock WorkDrive API
type MockAPI struct {
	files map[string]*MockFile
}

// NewMockAPI creates a new mock API instance
func NewMockAPI() *MockAPI {
	now := time.Now()
	
	// Create sample data matching your actual WorkDrive content
	files := map[string]*MockFile{
		"root": {
			ID:         "root",
			Name:       "My WorkDrive",
			Type:       "folder",
			CreatedAt:  now.AddDate(0, -1, 0),
			ModifiedAt: now.AddDate(0, -1, 0),
		},
		"folder123456789": {
			ID:         "folder123456789",
			Name:       "test-folder",
			Type:       "folder",
			ParentID:   "root",
			CreatedAt:  now.Add(-time.Hour),
			ModifiedAt: now.Add(-time.Hour),
		},
		"veysx16db130021d84de08b78167afc76c011": {
			ID:         "veysx16db130021d84de08b78167afc76c011",
			Name:       "test-file.txt",
			Type:       "file",
			Size:       1024,
			ParentID:   "folder123456789",
			CreatedAt:  now.Add(-30 * time.Minute),
			ModifiedAt: now.Add(-30 * time.Minute),
			DownloadURL: "https://mock-workdrive.local/download/veysx16db130021d84de08b78167afc76c011",
		},
		"file456789": {
			ID:         "file456789",
			Name:       "document.pdf",
			Type:       "file",
			Size:       2048,
			ParentID:   "root",
			CreatedAt:  now.Add(-2 * time.Hour),
			ModifiedAt: now.Add(-time.Hour),
			DownloadURL: "https://mock-workdrive.local/download/file456789",
		},
	}
	
	return &MockAPI{files: files}
}

// authenticate checks the authorization header
func (m *MockAPI) authenticate(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	// Accept both formats
	return strings.HasPrefix(auth, "Zoho-oauthtoken") || strings.HasPrefix(auth, "Bearer")
}

// sendError sends an error response
func (m *MockAPI) sendError(w http.ResponseWriter, code int, errorID, title string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	
	response := map[string]interface{}{
		"errors": []map[string]string{
			{
				"id":    errorID,
				"title": title,
			},
		},
	}
	
	json.NewEncoder(w).Encode(response)
}

// sendSuccess sends a success response
func (m *MockAPI) sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"data":   data,
		"status": "success",
	}
	
	json.NewEncoder(w).Encode(response)
}

// handleFiles handles /workdrive/api/v1/files requests
func (m *MockAPI) handleFiles(w http.ResponseWriter, r *http.Request) {
	if !m.authenticate(r) {
		m.sendError(w, http.StatusUnauthorized, "F000", "INVALID_TICKET")
		return
	}
	
	switch r.Method {
	case "GET":
		// List files
		parentID := r.URL.Query().Get("parent_id")
		if parentID == "" {
			parentID = "root"
		}
		
		var files []*MockFile
		for _, file := range m.files {
			if file.ParentID == parentID {
				files = append(files, file)
			}
		}
		
		m.sendSuccess(w, files)
		
	case "POST":
		// Upload file (mock)
		m.sendSuccess(w, map[string]interface{}{
			"id":      "new-file-" + fmt.Sprintf("%d", time.Now().Unix()),
			"message": "File uploaded successfully",
		})
		
	default:
		m.sendError(w, http.StatusMethodNotAllowed, "F6004", "Invalid Method")
	}
}

// handleFile handles /workdrive/api/v1/files/{id} requests
func (m *MockAPI) handleFile(w http.ResponseWriter, r *http.Request) {
	if !m.authenticate(r) {
		m.sendError(w, http.StatusUnauthorized, "F000", "INVALID_TICKET")
		return
	}
	
	// Extract file ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		m.sendError(w, http.StatusBadRequest, "F6001", "Invalid file ID")
		return
	}
	
	fileID := pathParts[4]
	file, exists := m.files[fileID]
	if !exists {
		m.sendError(w, http.StatusNotFound, "F6002", "File not found")
		return
	}
	
	switch r.Method {
	case "GET":
		m.sendSuccess(w, file)
	case "DELETE":
		delete(m.files, fileID)
		m.sendSuccess(w, map[string]string{"message": "File deleted successfully"})
	default:
		m.sendError(w, http.StatusMethodNotAllowed, "F6004", "Invalid Method")
	}
}

// handleAccount handles /workdrive/api/v1/account requests
func (m *MockAPI) handleAccount(w http.ResponseWriter, r *http.Request) {
	if !m.authenticate(r) {
		m.sendError(w, http.StatusUnauthorized, "F000", "INVALID_TICKET")
		return
	}
	
	account := map[string]interface{}{
		"user_id":       "123456789",
		"email":         "user@example.com",
		"name":          "Test User",
		"storage_used":  1073741824,   // 1GB
		"storage_total": 107374182400, // 100GB
	}
	
	m.sendSuccess(w, account)
}

// handleWorkspaces handles /workdrive/api/v1/workspaces requests
func (m *MockAPI) handleWorkspaces(w http.ResponseWriter, r *http.Request) {
	if !m.authenticate(r) {
		m.sendError(w, http.StatusUnauthorized, "F000", "INVALID_TICKET")
		return
	}
	
	workspaces := []map[string]interface{}{
		{
			"id":          "root",
			"name":        "My WorkDrive",
			"type":        "privatespace",
			"permissions": []string{"read", "write", "delete"},
		},
	}
	
	m.sendSuccess(w, workspaces)
}

func main() {
	fmt.Println("🚀 Starting Mock Zoho WorkDrive API Server")
	fmt.Println("=========================================")
	fmt.Println()
	
	api := NewMockAPI()
	
	// Set up routes
	http.HandleFunc("/workdrive/api/v1/files", api.handleFiles)
	http.HandleFunc("/workdrive/api/v1/files/", api.handleFile)
	http.HandleFunc("/workdrive/api/v1/account", api.handleAccount)
	http.HandleFunc("/workdrive/api/v1/workspaces", api.handleWorkspaces)
	
	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Mock WorkDrive API is running"))
	})
	
	// CORS and logging middleware
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		fmt.Printf("[%s] %s %s\n", time.Now().Format("15:04:05"), r.Method, r.URL.Path)
		
		// Route to appropriate handler
		if strings.HasPrefix(r.URL.Path, "/workdrive/api/v1/files/") && len(strings.Split(r.URL.Path, "/")) > 4 {
			api.handleFile(w, r)
		} else if r.URL.Path == "/workdrive/api/v1/files" {
			api.handleFiles(w, r)
		} else if r.URL.Path == "/workdrive/api/v1/account" {
			api.handleAccount(w, r)
		} else if r.URL.Path == "/workdrive/api/v1/workspaces" {
			api.handleWorkspaces(w, r)
		} else if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Mock WorkDrive API is running"))
		} else {
			api.sendError(w, http.StatusNotFound, "F6016", "URL Rule is not configured")
		}
	})
	
	port := ":8090"
	fmt.Printf("🌐 Mock API running on http://localhost%s\n", port)
	fmt.Println("📋 Available endpoints:")
	fmt.Println("   GET  /workdrive/api/v1/files")
	fmt.Println("   GET  /workdrive/api/v1/files/{id}")
	fmt.Println("   GET  /workdrive/api/v1/account")
	fmt.Println("   GET  /workdrive/api/v1/workspaces")
	fmt.Println("   GET  /health")
	fmt.Println()
	fmt.Println("🔑 Use any Authorization header (Zoho-oauthtoken or Bearer)")
	fmt.Println("⏳ Server ready for ZohoSync testing...")
	
	log.Fatal(http.ListenAndServe(port, nil))
}