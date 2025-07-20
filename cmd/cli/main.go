// ZohoSync CLI - Command line interface for ZohoSync
package main

import (
	"fmt"
	"os"

	"github.com/bdstest/zohosync/internal/config"
	"github.com/bdstest/zohosync/internal/utils"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	buildDate = "unknown"
	commit    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "zohosync-cli",
	Short: "ZohoSync CLI - Sync your Zoho WorkDrive files",
	Long: `ZohoSync CLI provides command-line access to Zoho WorkDrive synchronization.
	
Secure, lightweight sync client for Linux that keeps your files synchronized
between your local machine and Zoho WorkDrive.`,
	Version: fmt.Sprintf("%s (Built: %s, Commit: %s)", version, buildDate, commit),
}

func init() {
	// Add commands here as we implement them
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ZohoSync CLI %s\n", version)
		fmt.Printf("Build Date: %s\n", buildDate)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Go Version: %s\n", "1.21+")
	},
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
	}
	
	// Initialize logger
	utils.InitLogger(cfg.App.LogLevel)
	
	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
