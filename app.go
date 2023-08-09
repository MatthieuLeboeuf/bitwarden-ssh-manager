package main

import (
	"context"
	"encoding/json"
	"github.com/MatthieuLeboeuf/bitwarden-ssh-manager/backend"
	"github.com/spf13/viper"
	"os/exec"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetEmail() string {
	return viper.GetString("vault.email")
}

func (a *App) IsLogged() bool {
	return backend.GlobalData.AccessToken != ""
}

// Login return true if two-factor authentication is required
func (a *App) Login(password string, otpCode int, otpRemember bool) int {
	// Get access token to backend
	backend.PreLogin()
	return backend.Login(password, otpCode, otpRemember)
}

func (a *App) Sync() string {
	// Sync data
	return backend.Sync()
}

func (a *App) LaunchTerminal(data string) {
	host := &backend.FrontendHost{}
	json.Unmarshal([]byte(data), &host)

	var command string

	if host.Password == "" {
		command = "ssh " + host.Username + "@" + host.Host + " -p " + host.Port
	} else {
		command = "sshpass -p " + host.Password + " ssh " + host.Username + "@" + host.Host + " -p " + host.Port
	}

	cmd := exec.Command("kgx", "-e", command)
	_ = cmd.Run()
}
