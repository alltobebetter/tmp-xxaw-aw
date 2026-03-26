//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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
