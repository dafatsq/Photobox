package main

import (
	"embed"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Enable WebView2 camera/media access without permission prompts
	os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", "--enable-media-stream --use-fake-ui-for-media-stream")

	// Create an instance of the app with platform-specific hardware drivers
	app := NewPlatformApp()

	// Create application with kiosk-oriented options
	err := wails.Run(&options.App{
		Title:     "Photobox",
		Width:     1024,
		Height:    768,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 15, B: 20, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
