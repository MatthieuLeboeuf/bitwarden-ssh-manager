package main

import (
	"embed"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Init config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.bsm")
	err := viper.ReadInConfig()

	if err != nil {
		println("Config error:", err.Error())
	}

	if viper.GetString("device.identifier") == "" {
		viper.Set("device.identifier", uuid.New().String())
		_ = viper.WriteConfig()
	}

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "Bitwarden ssh manager",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
