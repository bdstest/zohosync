// ZohoSync CLI - Command line interface for ZohoSync
package main

import (
	"fmt"
	"os"

	"github.com/bdstest/zohosync/internal/config"
	"github.com/bdstest/zohosync/internal/ui/cli"
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
	// Initialize CLI
	cliInstance, err := cli.NewCLI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize CLI: %v\n", err)
		os.Exit(1)
	}

	// Add commands
	rootCmd.AddCommand(cliInstance.CreateLoginCommand())
	rootCmd.AddCommand(cliInstance.CreateStatusCommand())
	rootCmd.AddCommand(cliInstance.CreateSyncCommand())
	rootCmd.AddCommand(cliInstance.CreateListCommand())
	rootCmd.AddCommand(cliInstance.CreateVersionCommand(version, buildDate, commit))
}

func main() {
	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
