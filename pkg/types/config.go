// Package types contains shared type definitions for ZohoSync
package types

// Config represents the application configuration
type Config struct {
	App      AppConfig      `yaml:"app" json:"app"`
	Auth     AuthConfig     `yaml:"auth" json:"auth"`
	Sync     SyncConfig     `yaml:"sync" json:"sync"`
	Network  NetworkConfig  `yaml:"network" json:"network"`
	UI       UIConfig       `yaml:"ui" json:"ui"`
	Folders  []FolderConfig `yaml:"folders" json:"folders"`
}

// AppConfig contains general application settings
type AppConfig struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
	LogLevel string `yaml:"log_level" json:"log_level"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	ClientID     string `yaml:"client_id" json:"client_id"`
	ClientSecret string `yaml:"client_secret" json:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri" json:"redirect_uri"`
	Scopes       []string `yaml:"scopes" json:"scopes"`
}

// SyncConfig contains synchronization settings
type SyncConfig struct {
	Interval            int    `yaml:"interval" json:"interval"`
	ConflictResolution  string `yaml:"conflict_resolution" json:"conflict_resolution"`
	MaxConcurrentSyncs  int    `yaml:"max_concurrent_syncs" json:"max_concurrent_syncs"`
}

// NetworkConfig contains network settings
type NetworkConfig struct {
	ProxyURL         string `yaml:"proxy_url" json:"proxy_url"`
	Timeout          int    `yaml:"timeout" json:"timeout"`
	MaxRetries       int    `yaml:"max_retries" json:"max_retries"`
	BandwidthLimit   int    `yaml:"bandwidth_limit" json:"bandwidth_limit"`
}

// UIConfig contains UI settings
type UIConfig struct {
	Theme              string `yaml:"theme" json:"theme"`
	ShowNotifications  bool   `yaml:"show_notifications" json:"show_notifications"`
	MinimizeToTray     bool   `yaml:"minimize_to_tray" json:"minimize_to_tray"`
}

// FolderConfig represents a sync folder configuration
type FolderConfig struct {
	Local     string `yaml:"local" json:"local"`
	Remote    string `yaml:"remote" json:"remote"`
	SyncMode  string `yaml:"sync_mode" json:"sync_mode"`
	Enabled   bool   `yaml:"enabled" json:"enabled"`
}
