package main

import (
	"context"
	"fmt"
	"runtime"

	"TraeProxy/proxy"
	"github.com/denisbrodbeck/machineid"
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
	return runtime.GOOS
}
