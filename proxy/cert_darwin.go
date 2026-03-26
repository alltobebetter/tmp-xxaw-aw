//go:build darwin

package proxy

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func InstallCA() error {
	certPath, err := GetCertPath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(certPath); err != nil {
		return errors.New("certificate file not generated yet")
	}

	// Add cert to the user's login keychain and mark it as trusted for SSL.
	// This will prompt the user for their macOS login password (similar to UAC on Windows).
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	keychainPath := filepath.Join(homeDir, "Library", "Keychains", "login.keychain-db")

	cmd := exec.Command("security", "add-trusted-cert", "-r", "trustRoot", "-k", keychainPath, certPath)
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

	// Search for our cert in all keychains
	cmd := exec.Command("security", "find-certificate", "-c", "TraeProxy Root CA", "-a")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "TraeProxy Root CA")
}

func UninstallCA() error {
	// Delete the certificate by its Common Name from the login keychain
	cmd := exec.Command("security", "delete-certificate", "-c", "TraeProxy Root CA")
	return cmd.Run()
}
