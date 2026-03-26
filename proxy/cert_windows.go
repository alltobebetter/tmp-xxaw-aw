//go:build windows

package proxy

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func InstallCA() error {
	certPath, err := GetCertPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(certPath); err != nil {
		return errors.New("certificate file not generated yet")
	}

	// This triggers UAC on Windows
	cmd := exec.Command("certutil", "-addstore", "-user", "root", certPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

func IsCAInstalled() bool {
	certPath, err := GetCertPath()
	if err != nil {
		return false
	}
	if _, err := os.Stat(certPath); err != nil {
		return false
	}
	// Check store
	cmd := exec.Command("certutil", "-user", "-store", "root", "TraeProxy Root CA")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "TraeProxy")
}

func UninstallCA() error {
	// This triggers UAC on Windows
	cmd := exec.Command("certutil", "-delstore", "-user", "root", "TraeProxy Root CA")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}
