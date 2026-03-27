package main

import (
	"context"
	"embed"
	"fmt"
	"net"
	"os"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.design/x/hotkey"
)

//go:embed all:frontend/dist
var assets embed.FS

const singleInstanceAddr = "127.0.0.1:19199"

func main() {
	app := NewApp()

	// --- Single-instance detection ---
	// Try to listen on the fixed port. If it fails, another instance is running.
	listener, err := net.Listen("tcp", singleInstanceAddr)
	if err != nil {
		// Another instance owns the port – tell it to show its window, then exit.
		conn, dialErr := net.Dial("tcp", singleInstanceAddr)
		if dialErr == nil {
			conn.Write([]byte("SHOW"))
			conn.Close()
		}
		os.Exit(0)
	}

	// Accept connections in background – when we receive "SHOW", bring window up.
	var showWindowFunc func()
	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			buf := make([]byte, 16)
			n, _ := conn.Read(buf)
			conn.Close()
			if string(buf[:n]) == "SHOW" && showWindowFunc != nil {
				showWindowFunc()
			}
		}
	}()

	// macOS frameless windows have square corners by default.
	// Making the window background transparent lets CSS border-radius work.
	bgAlpha := uint8(1)
	if runtime.GOOS == "darwin" {
		bgAlpha = 0
	}

	err = wails.Run(&options.App{
		Title:         "TraeProxy",
		Width:         700,
		Height:        660,
		DisableResize: true,
		Frameless:     true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: bgAlpha},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)

			// Wire up the show-window callback for single-instance listener
			showWindowFunc = func() {
				wailsRuntime.WindowShow(ctx)
				wailsRuntime.WindowCenter(ctx)
			}

			// --- Global hotkey registration ---
			go func() {
				mods := hotkeyModifiers()
				hk := hotkey.New(mods, hotkey.KeyT)
				if err := hk.Register(); err != nil {
					fmt.Println("Hotkey register failed:", err)
					return
				}
				for range hk.Keydown() {
					wailsRuntime.WindowShow(ctx)
					wailsRuntime.WindowCenter(ctx)
				}
			}()
		},
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			WebviewIsTransparent: true,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
