package main

import (
	"context"
	"fmt"
	"os"
	goruntime "runtime"

	"TraeProxy/proxy"
	"github.com/denisbrodbeck/machineid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
	srv *proxy.Server
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		srv: proxy.New(),
	}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.srv.SetContext(ctx)
	// Ensure CA generated quietly
	_, _, _ = proxy.EnsureCA()
}

func (a *App) InstallCert() error {
	_, _, err := proxy.EnsureCA()
	if err != nil {
		return err
	}
	return proxy.InstallCA()
}

func (a *App) UninstallCert() error {
	return proxy.UninstallCA()
}

func (a *App) IsCertInstalled() bool {
	return proxy.IsCAInstalled()
}

func (a *App) StartProxy(port int, openaiBase string, anthropicBase string) error {
	_ = a.srv.Stop() // Ensure previous stopped
	certBytes, keyBytes, err := proxy.EnsureCA()
	if err != nil {
		return fmt.Errorf("failed to get CA: %v", err)
	}
	return a.srv.Start(port, openaiBase, anthropicBase, certBytes, keyBytes)
}

func (a *App) StopProxy() error {
	return a.srv.Stop()
}

func (a *App) GetMachineID() (string, error) {
	return machineid.ProtectedID("TraeProxy")
}

func (a *App) GetPlatform() string {
	return goruntime.GOOS
}

func (a *App) HideWindow() {
	runtime.WindowHide(a.ctx)
}

func (a *App) QuitApp() {
	_ = a.srv.Stop()
	runtime.Quit(a.ctx)
}

// UpdateKeyPool updates the key rotation pools. Can be called while proxy is running.
func (a *App) UpdateKeyPool(openaiKeys []string, anthropicKeys []string, generalKeys []string) {
	a.srv.SetKeyPools(openaiKeys, anthropicKeys, generalKeys)
}

// ExportKeysToFile opens a native save dialog and writes JSON content to the chosen file.
func (a *App) ExportKeysToFile(jsonContent string) error {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: "traeproxy-keys.json",
		Title:           "导出密钥",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON 文件 (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil {
		return err
	}
	if path == "" {
		return nil // User cancelled
	}
	return os.WriteFile(path, []byte(jsonContent), 0644)
}

// ImportKeysFromFile opens a native open dialog and returns the file content as a string.
func (a *App) ImportKeysFromFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "导入密钥",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON 文件 (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil // User cancelled
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
