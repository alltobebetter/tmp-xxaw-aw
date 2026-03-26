package main

import (
	"context"
	"embed"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// macOS uses native titlebar (rounded corners + traffic lights)
	// Windows uses custom frameless titlebar
	frameless := runtime.GOOS != "darwin"

	// Create application with options
	err := wails.Run(&options.App{
		Title:         "TraeProxy",
		Width:         440,
		Height:        640,
		DisableResize: true,
		Frameless:     frameless,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			// Emit event to frontend so it can show the close confirmation dialog
			wailsRuntime.EventsEmit(ctx, "close-requested")
			return true // Prevent default close; frontend will call Quit() after confirmation
		},
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarDefault(),
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

