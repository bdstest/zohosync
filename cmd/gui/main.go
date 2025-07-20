// ZohoSync GUI - Desktop application for ZohoSync
package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	
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
	}
	
	// Initialize logger
	logger := utils.InitLogger(cfg.App.LogLevel)
	logger.Info("Starting ZohoSync GUI")
	
	// Create Fyne application
	myApp := app.New()
	myApp.Settings().SetTheme(&zohoTheme{})
	
	// Create main window
	myWindow := myApp.NewWindow("ZohoSync")
	myWindow.Resize(fyne.NewSize(800, 600))
	
	// Create basic UI
	hello := widget.NewLabel("Welcome to ZohoSync!")
	content := container.NewVBox(
		hello,
		widget.NewButton("Connect to Zoho WorkDrive", func() {
			hello.SetText("Connecting... (Not implemented yet)")
		}),
	)
	
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

// Basic theme placeholder
type zohoTheme struct{}

func (z zohoTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

func (z zohoTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (z zohoTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (z zohoTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
