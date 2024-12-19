package main

import (
	"context"
	"fmt"
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

func (a *App) DownloadYoutube(url string, format string, quality string) {
	fmt.Println("Downloading media from", url, format,quality)
}
func (a *App) DownloadAppleMusic(url string, format string, quality string) {
	fmt.Println("Downloading media from", url, format,quality)
}
