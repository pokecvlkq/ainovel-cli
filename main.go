package main

import (
	"embed"
	"log"

	"github.com/voocel/ainovel-cli/assets"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var dist embed.FS

func main() {
	// 1. Tải cấu hình
	cfg, err := bootstrap.LoadConfig("")
	if err != nil {
		log.Fatalf("Lỗi tải cấu hình: %v", err)
	}

	// 2. Tải assets
	bundle := assets.Load(cfg.Style)

	// 3. Khởi tạo App Wails
	app := NewApp(cfg, bundle)

	// 4. Khởi chạy Wails
	err = wails.Run(&options.App{
		Title:  "AINovel Writer",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: dist,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
