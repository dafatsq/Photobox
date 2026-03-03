package main

import (
	"embed"
	"os"
	"path/filepath"

	"photobox/admin"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Enable WebView2 camera/media access without permission prompts
	os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", "--enable-media-stream --use-fake-ui-for-media-stream")

	// Ensure PhotoboxData storage directory exists
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, "PhotoboxData")
	framesDir := filepath.Join(dataDir, "frames")
	os.MkdirAll(framesDir, 0755)

	// Create shared admin config
	adminCfg := admin.NewAdminConfig(framesDir)

	// Start admin dashboard on port 8080 (separate browser window)
	go admin.StartAdminServer(adminCfg, 8080)

	// Create an instance of the app with platform-specific hardware drivers
	app := NewPlatformApp(adminCfg)

	// Create application with kiosk-oriented options
	err := wails.Run(&options.App{
		Title:             "Photobox",
		Width:             1024,
		Height:            768,
		DisableResize:     true,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
