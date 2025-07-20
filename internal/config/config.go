// Package config handles application configuration
package config

import (
	"os"
	"path/filepath"

	"github.com/bdstest/zohosync/pkg/types"
	"github.com/spf13/viper"
)

// LoadConfig loads the application configuration
func LoadConfig() (*types.Config, error) {
	// Set config name and type
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	
	// Add config paths
	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Join(os.Getenv("HOME"), ".config", "zohosync"))
	viper.AddConfigPath("/etc/zohosync")
	
	// Set defaults
	setDefaults()
	
	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		// Create default config if not exists
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return createDefaultConfig()
		}
		return nil, err
	}
	
	// Unmarshal config
	var config types.Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	
	return &config, nil
}

func setDefaults() {
	viper.SetDefault("app.name", "ZohoSync")
	viper.SetDefault("app.version", "0.1.0")
	viper.SetDefault("app.log_level", "info")
	
	viper.SetDefault("auth.redirect_uri", "http://localhost:8080/callback")
	viper.SetDefault("auth.scopes", []string{"WorkDrive.files.ALL", "WorkDrive.folders.ALL"})
	
	viper.SetDefault("sync.interval", 300)
	viper.SetDefault("sync.conflict_resolution", "newer")
	viper.SetDefault("sync.max_concurrent_syncs", 5)
	
	viper.SetDefault("network.timeout", 30)
	viper.SetDefault("network.max_retries", 3)
	
	viper.SetDefault("ui.theme", "light")
	viper.SetDefault("ui.show_notifications", true)
	viper.SetDefault("ui.minimize_to_tray", true)
}

func createDefaultConfig() (*types.Config, error) {
	config := &types.Config{
		App: types.AppConfig{
			Name:     "ZohoSync",
			Version:  "0.1.0",
			LogLevel: "info",
		},
		Auth: types.AuthConfig{
			RedirectURI: "http://localhost:8080/callback",
			Scopes:      []string{"WorkDrive.files.ALL", "WorkDrive.folders.ALL"},
		},
		Sync: types.SyncConfig{
			Interval:           300,
			ConflictResolution: "newer",
			MaxConcurrentSyncs: 5,
		},
		Network: types.NetworkConfig{
			Timeout:    30,
			MaxRetries: 3,
		},
		UI: types.UIConfig{
			Theme:             "light",
			ShowNotifications: true,
			MinimizeToTray:    true,
		},
	}
	
	return config, nil
}
