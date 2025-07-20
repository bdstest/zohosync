// Package api provides Zoho WorkDrive API client functionality
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/bdstest/zohosync/pkg/types"
	"github.com/bdstest/zohosync/internal/config"
	"github.com/bdstest/zohosync/internal/utils"
)

// Client represents the Zoho WorkDrive API client
type Client struct {
	httpClient  *http.Client
	baseURL     string
	uploadURL   string
	downloadURL string
	token       *types.TokenInfo
	logger      *utils.Logger
}

// NewClient creates a new Zoho WorkDrive API client
func NewClient(token *types.TokenInfo) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:     config.APIBaseURL,
		uploadURL:   config.UploadBaseURL,
		downloadURL: config.DownloadBaseURL,
		token:       token,
		logger:      utils.GetLogger(),
	}
}

// SetToken updates the authentication token
func (c *Client) SetToken(token *types.TokenInfo) {
	c.token = token
}

// makeRequest performs an authenticated HTTP request
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// UserInfo represents user information from Zoho
type UserInfo struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	TimeZone    string `json:"time_zone"`
	Language    string `json:"language"`
}

// GetUserInfo retrieves current user information
func (c *Client) GetUserInfo(ctx context.Context) (*UserInfo, error) {
	resp, err := c.makeRequest(ctx, "GET", "/users/me", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var result struct {
		Data UserInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Info("Successfully retrieved user information")
	return &result.Data, nil
}

// FileInfo represents file metadata from Zoho WorkDrive
type FileInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Size         int64     `json:"size"`
	CreatedTime  time.Time `json:"created_time"`
	ModifiedTime time.Time `json:"modified_time"`
	ParentID     string    `json:"parent_id"`
	Path         string    `json:"path"`
	IsFolder     bool      `json:"is_folder"`
	DownloadURL  string    `json:"download_url"`
	Permission   string    `json:"permission"`
}

// ListFiles retrieves files from a specific folder
func (c *Client) ListFiles(ctx context.Context, folderID string, limit int) ([]FileInfo, error) {
	endpoint := fmt.Sprintf("/files/%s/files", folderID)
	
	// Add query parameters
	params := url.Values{}
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var result struct {
		Data []FileInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Infof("Retrieved %d files from folder %s", len(result.Data), folderID)
	return result.Data, nil
}

// GetRootFolder retrieves the root folder information
func (c *Client) GetRootFolder(ctx context.Context) (*FileInfo, error) {
	resp, err := c.makeRequest(ctx, "GET", "/files", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var result struct {
		Data FileInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Info("Retrieved root folder information")
	return &result.Data, nil
}

// DownloadFile downloads a file from Zoho WorkDrive
func (c *Client) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error) {
	endpoint := fmt.Sprintf("/files/%s/download", fileID)
	
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	c.logger.Infof("Started download for file %s", fileID)
	return resp.Body, nil
}

// CreateFolder creates a new folder
func (c *Client) CreateFolder(ctx context.Context, parentID, name string) (*FileInfo, error) {
	body := map[string]interface{}{
		"name":      name,
		"parent_id": parentID,
		"type":      "folder",
	}

	resp, err := c.makeRequest(ctx, "POST", "/files", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("folder creation failed with status %d", resp.StatusCode)
	}

	var result struct {
		Data FileInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Infof("Created folder '%s' in parent %s", name, parentID)
	return &result.Data, nil
}

// FileUploadInfo represents upload session information
type FileUploadInfo struct {
	UploadID    string `json:"upload_id"`
	UploadURL   string `json:"upload_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// InitiateUpload initiates a file upload session
func (c *Client) InitiateUpload(ctx context.Context, filename string, fileSize int64, parentID string) (*FileUploadInfo, error) {
	body := map[string]interface{}{
		"filename":  filename,
		"file_size": fileSize,
		"parent_id": parentID,
	}

	endpoint := "/upload/initiate"
	req, err := http.NewRequestWithContext(ctx, "POST", c.uploadURL+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	jsonBody, _ := json.Marshal(body)
	req.Body = io.NopCloser(bytes.NewBuffer(jsonBody))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload initiation failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload initiation failed with status %d", resp.StatusCode)
	}

	var result struct {
		Data FileUploadInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Infof("Initiated upload for file '%s'", filename)
	return &result.Data, nil
}

// DeleteFile deletes a file or folder
func (c *Client) DeleteFile(ctx context.Context, fileID string) error {
	endpoint := fmt.Sprintf("/files/%s", fileID)
	
	resp, err := c.makeRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete failed with status %d", resp.StatusCode)
	}

	c.logger.Infof("Deleted file %s", fileID)
	return nil
}

// GetFileInfo retrieves metadata for a specific file
func (c *Client) GetFileInfo(ctx context.Context, fileID string) (*FileInfo, error) {
	endpoint := fmt.Sprintf("/files/%s", fileID)
	
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var result struct {
		Data FileInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Data, nil
}