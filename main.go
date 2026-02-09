package main

import (
	"audivo-media-downloader/internal/AppleMusicDownloader"
	"audivo-media-downloader/internal/YoutubeDownloader"
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Audivo Media Downloader",
		Width:  800,
		Height: 600, 
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:        app.Startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func DownloadAppleMusic(ctx context.Context, url, format, quality string) {
	AppleMusicDownloader.DownloadAppleMusic(ctx, url, format, quality)
}

func DownloadYoutube(ctx context.Context, url, format, quality string) {
	YoutubeDownloader.DownloadYoutube(ctx, url, format, quality)
}
