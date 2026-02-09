package YoutubeDownloader

import (
	"audivo-media-downloader/tools"
	"context"
	"os"
	"os/exec"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func DownloadYoutube(ctx context.Context, url, format, quality string) {

    
    dependencyPaths, err := tools.DependencyDownloader()
        if err != nil {
            wailsruntime.EventsEmit(ctx, "download-progress", "Error downloading yt-dlp: "+err.Error())
            return
        }
        
    ytDlpPath := dependencyPaths.YtDlpPath

    if err := os.Chmod(ytDlpPath, 0755); err != nil {
        wailsruntime.EventsEmit(ctx, "download-progress", "Error setting permissions: "+err.Error())
        return
    }

    args := []string{"-o", "~/Downloads/%(artist)s-%(title)s.%(ext)s"}

    if format == "audio" {
        args = append(args, "-x", "--audio-format", "mp3", "--audio-quality", "0")
    }

    if format == "video" {
        args = append(args, "-S", "ext")
        if quality == "best" {
            args = append(args, "-f", "bv*+ba/b", "--remux-video", "mp4", "--merge-output-format", "mp4")
        } else {
            args = append(args, "-f", "bestvideo[height<="+quality+"]+bestaudio/best[height<="+quality+"]", "--merge-output-format", "mp4")
        }
    }

    args = append(args, url)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func(){
		wailsruntime.EventsOn(ctx, "cancel-download", func(optionalData ...interface{}) {
            cancel()
            wailsruntime.EventsEmit(ctx, "download-progress", "Download cancelled by user")
        })
	}()

    cmd := exec.CommandContext(ctx, ytDlpPath, args...)
    stdoutPipe, err := cmd.StdoutPipe()
    if err != nil {
        wailsruntime.EventsEmit(ctx, "download-progress", "Error getting stdout: "+err.Error())
        return
    }

    if err := cmd.Start(); err != nil {
        wailsruntime.EventsEmit(ctx, "download-progress", "Error starting the command: "+err.Error())
        return
    }

    go func() {
        buf := make([]byte, 1024)
        for {
            n, err := stdoutPipe.Read(buf)
            if n > 0 {
                output := string(buf[:n])
                wailsruntime.EventsEmit(ctx, "download-progress", output)
            }
            if err != nil {
                break
            }
        }
    }()

    if err := cmd.Wait(); err != nil {
        wailsruntime.EventsEmit(ctx, "download-progress", "Error: "+err.Error())
        return
    }
}

