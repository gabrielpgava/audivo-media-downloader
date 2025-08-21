package tools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)


type DependencyPaths struct {
	YtDlpPath  string
	FfmpegPath string
	GamdlPath  string
}

func DependencyDownloader() (DependencyPaths, error) {
    
    binDir, err := checkDirOrCreate()
    if err != nil {
        return DependencyPaths{}, fmt.Errorf("error checking or creating directory: %w", err)
    }

    ytDlpPath := filepath.Join(binDir, "yt-dlp")
    ffmpegPath := filepath.Join(binDir, "ffmpeg")
    gamdlPath := filepath.Join(binDir, "gamdl")
    if runtime.GOOS == "windows" {
        ytDlpPath += ".exe"
        ffmpegPath += ".exe"
        gamdlPath += ".exe"
    }

    // Download yt-dlp if necessary
    if _, err := os.Stat(ytDlpPath); err != nil {
        var ytDlpURL string
        switch runtime.GOOS {
        case "windows":
            ytDlpURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe"
        case "darwin":
            ytDlpURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos"
        case "linux":
            ytDlpURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp"
        default:
            return DependencyPaths{}, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
        }
        if err := downloadFile(ytDlpURL, ytDlpPath); err != nil {
            return DependencyPaths{}, fmt.Errorf("error downloading yt-dlp: %w", err)
        }
        os.Chmod(ytDlpPath, 0755)
    }

    // Download ffmpeg if necessary
    if _, err := os.Stat(ffmpegPath); err != nil {
        var ffmpegURL string
        switch runtime.GOOS {
        case "windows":
            ffmpegURL = "https://github.com/eugeneware/ffmpeg-static/releases/latest/download/ffmpeg-win32.exe"
        case "darwin":
            ffmpegURL = "https://github.com/eugeneware/ffmpeg-static/releases/latest/download/ffmpeg-macos"
        case "linux":
            ffmpegURL = "https://github.com/eugeneware/ffmpeg-static/releases/latest/download/ffmpeg-linux"
        default:
            return DependencyPaths{}, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
        }
        if err := downloadFile(ffmpegURL, ffmpegPath); err != nil {
            return DependencyPaths{}, fmt.Errorf("error downloading ffmpeg: %w", err)
        }
        os.Chmod(ffmpegPath, 0755)
    }

    // Download gamdl if necessary
    if _, err := os.Stat(gamdlPath); err != nil {
        var gamdlURL string
        switch runtime.GOOS {
        case "windows":
            gamdlURL = "https://github.com/gabrielpasilva/gamdl/releases/latest/download/gamdl-windows.exe"
        case "darwin":
            gamdlURL = "https://github.com/gabrielpasilva/gamdl/releases/latest/download/gamdl-macos"
        case "linux":
            gamdlURL = "https://github.com/gabrielpasilva/gamdl/releases/latest/download/gamdl-linux"
        default:
            return DependencyPaths{}, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
        }
        if err := downloadFile(gamdlURL, gamdlPath); err != nil {
            return DependencyPaths{}, fmt.Errorf("error downloading gamdl: %w", err)
        }
        os.Chmod(gamdlPath, 0755)
    }

    return DependencyPaths{
		YtDlpPath:  ytDlpPath,
		FfmpegPath: ffmpegPath,
		GamdlPath:  gamdlPath,
	}, nil
}

func checkDirOrCreate()(string, error) {
    var binDir string

    home, err := os.UserHomeDir()

    if err != nil {
        return "", fmt.Errorf("failed to get user home directory: %w", err)
    }
    if runtime.GOOS == "windows" {
        appData := os.Getenv("APPDATA")
        if appData != "" {
            binDir = filepath.Join(appData, "audivo-media-downloader")
        } else {
            binDir = filepath.Join(home, ".audivo-media-downloader")
        }
    } else {
        binDir = filepath.Join(home, ".audivo-media-downloader")
    }

    if err := os.MkdirAll(binDir, 0755); err != nil {
        return "", fmt.Errorf("failed to create assets directory: %w", err)
    }

    return binDir, nil
}

//TODO: SEND TO FRONTEND THE DEPENDENCY DOWNLOAD PROGRESS

func downloadFile(url, dest string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    out, err := os.Create(dest)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, resp.Body)
    return err
}