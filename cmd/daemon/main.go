// ZohoSync Daemon - Background synchronization service
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bdstest/zohosync/internal/config"
	"github.com/bdstest/zohosync/internal/utils"
)

var (
	version   = "dev"
	buildDate = "unknown"
	commit    = "unknown"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	
	// Initialize logger
	logger := utils.InitLogger(cfg.App.LogLevel)
	logger.Info("Starting ZohoSync daemon")
	logger.Infof("Version: %s, Build: %s, Commit: %s", version, buildDate, commit)
	
	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Main daemon loop
	logger.Info("Daemon started successfully")
	
	// Wait for shutdown signal
	sig := <-sigChan
	logger.Infof("Received signal: %v, shutting down...", sig)
	
	// Cleanup
	logger.Info("Daemon stopped")
}
