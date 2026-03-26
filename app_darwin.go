//go:build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) SelectTraePath() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 Trae 或 Trae CN 的应用程序",
		Filters: []runtime.FileFilter{
			{DisplayName: "应用程序 (*.app)", Pattern: "*.app"},
			{DisplayName: "所有文件 (*.*)", Pattern: "*.*"},
		},
	})
}

// resolveMacOSBinary extracts the actual executable path from a .app bundle.
// e.g. /Applications/Trae.app => /Applications/Trae.app/Contents/MacOS/Trae
func resolveMacOSBinary(appPath string) (string, error) {
	if !strings.HasSuffix(appPath, ".app") {
		// Already a direct binary path, use as-is
		return appPath, nil
	}

	macosDir := filepath.Join(appPath, "Contents", "MacOS")
	entries, err := os.ReadDir(macosDir)
	if err != nil {
		return "", fmt.Errorf("无法读取应用包内容: %v", err)
	}

	// Find the first executable file in the MacOS directory
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		binPath := filepath.Join(macosDir, entry.Name())
		info, err := os.Stat(binPath)
		if err != nil {
			continue
		}
		// Check if executable
		if info.Mode()&0111 != 0 {
			return binPath, nil
		}
	}

	return "", fmt.Errorf("在应用包中未找到可执行文件: %s", macosDir)
}

func (a *App) LaunchTrae(path string, port int) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("找不到该路径下的程序: %v", err)
	}

	// Resolve .app bundle to actual binary
	binaryPath, err := resolveMacOSBinary(path)
	if err != nil {
		return err
	}

	proxyURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Electron apps need the proxy-server flag for the Chromium network stack
	cmd := exec.Command(binaryPath, fmt.Sprintf("--proxy-server=%s", proxyURL))

	// Inject the proxy variables for Node.js modules and other subprocesses
	cmd.Env = append(os.Environ(),
		"HTTP_PROXY="+proxyURL,
		"HTTPS_PROXY="+proxyURL,
		"http_proxy="+proxyURL,
		"https_proxy="+proxyURL,
		"ALL_PROXY="+proxyURL,
		"all_proxy="+proxyURL,
	)

	return cmd.Start()
}
