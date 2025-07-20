package api

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"encoding/json"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkDriveClient(t *testing.T) {
	// Mock WorkDrive API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Zoho-oauthtoken test_token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/api/v1/users/me":
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"id":    "12345",
					"name":  "Test User",
					"email": "test@example.com",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

		case "/api/v1/teams":
			teams := []map[string]interface{}{
				{
					"id":   "team1",
					"name": "Test Team",
					"type": "team",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": teams,
			})

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &WorkDriveClient{
		BaseURL:     server.URL,
		AccessToken: "test_token",
		HTTPClient:  &http.Client{},
	}

	ctx := context.Background()

	// Test user info retrieval
	user, err := client.GetUserInfo(ctx)
	require.NoError(t, err)
	assert.Equal(t, "12345", user.ID)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "test@example.com", user.Email)

	// Test teams listing
	teams, err := client.ListTeams(ctx)
	require.NoError(t, err)
	assert.Len(t, teams, 1)
	assert.Equal(t, "team1", teams[0].ID)
	assert.Equal(t, "Test Team", teams[0].Name)
}

func TestAPIErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedError  string
	}{
		{
			name:          "Unauthorized",
			statusCode:    401,
			responseBody:  `{"error":"invalid_token"}`,
			expectedError: "unauthorized",
		},
		{
			name:          "Rate Limited",
			statusCode:    429,
			responseBody:  `{"error":"rate_limit_exceeded"}`,
			expectedError: "rate limit",
		},
		{
			name:          "Server Error",
			statusCode:    500,
			responseBody:  `{"error":"internal_server_error"}`,
			expectedError: "server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &WorkDriveClient{
				BaseURL:     server.URL,
				AccessToken: "test_token",
				HTTPClient:  &http.Client{},
			}

			ctx := context.Background()
			_, err := client.GetUserInfo(ctx)
			
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestRetryLogic(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		
		// Success on third attempt
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "12345",
				"name": "Test User",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &WorkDriveClient{
		BaseURL:     server.URL,
		AccessToken: "test_token",
		HTTPClient:  &http.Client{},
		MaxRetries:  3,
	}

	ctx := context.Background()
	user, err := client.GetUserInfo(ctx)
	
	require.NoError(t, err)
	assert.Equal(t, "12345", user.ID)
	assert.Equal(t, 3, attemptCount) // Should have retried twice
}

func TestFileOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Zoho-oauthtoken test_token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch {
		case r.URL.Path == "/api/v1/files" && r.Method == "GET":
			files := []map[string]interface{}{
				{
					"id":           "file1",
					"name":         "document.txt",
					"type":         "file",
					"size":         1024,
					"modified_time": "2024-01-15T10:30:00Z",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": files,
			})

		case r.URL.Path == "/api/v1/download/file1":
			w.Write([]byte("file content"))

		case r.URL.Path == "/api/v1/upload" && r.Method == "POST":
			w.WriteHeader(http.StatusCreated)
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"id":   "new_file_id",
					"name": "uploaded_file.txt",
				},
			}
			json.NewEncoder(w).Encode(response)

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &WorkDriveClient{
		BaseURL:     server.URL,
		AccessToken: "test_token",
		HTTPClient:  &http.Client{},
	}

	ctx := context.Background()

	// Test file listing
	files, err := client.ListFiles(ctx, "team1")
	require.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "document.txt", files[0].Name)

	// Test file download
	content, err := client.DownloadFile(ctx, "file1")
	require.NoError(t, err)
	assert.Equal(t, []byte("file content"), content)

	// Test file upload
	uploadContent := []byte("test upload content")
	fileID, err := client.UploadFile(ctx, "team1", "test.txt", uploadContent)
	require.NoError(t, err)
	assert.Equal(t, "new_file_id", fileID)
}