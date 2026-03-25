package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

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
	_, _, err := proxy.EnsureCA() // This part is moved from the malformed snippet into InstallCert
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

func (a *App) SelectTraePath() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 Trae 或 Trae CN 的执行文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "执行程序 (*.exe)", Pattern: "*.exe"},
		},
	})
}

func (a *App) LaunchTrae(path string, port int) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("找不到该路径下的程序: %v", err)
	}
	proxyURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	
	// Electron apps often need the proxy-server flag directly for the Chromium network stack
	cmd := exec.Command(path, fmt.Sprintf("--proxy-server=%s", proxyURL))
	
	// Inject the proxy variables for Node.js modules and other subprocesses
	cmd.Env = append(os.Environ(),
		"HTTP_PROXY="+proxyURL,
		"HTTPS_PROXY="+proxyURL,
		"http_proxy="+proxyURL,
		"https_proxy="+proxyURL,
		"ALL_PROXY="+proxyURL,
		"all_proxy="+proxyURL,
	)
	
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	
	return cmd.Start()
}
