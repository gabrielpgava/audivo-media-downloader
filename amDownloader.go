package main

import (
	"fmt"
	"os/exec"
)

func (a *App) DownloadAppleMusic(url, format, quality string) {
	fmt.Println("Downloading media from", url, format, quality)

	args := []string{"-o", "~/Downloads/%(title)s.%(ext)s"}

	if format == "audio" {
		args = append(args, "-x", "--audio-format", "mp3")
	}

	if format == "video" {
		args = append(args, "-S", "ext")
		if quality == "best" {
			args = append(args, "-f", "bestvideo+bestaudio")
		}
		args = append(args, "-f", "bestvideo[height<=" + quality + "]+bestaudio/best[height<=" + quality + "]", "mp4")
	}

	args = append(args, url)

	fmt.Println("yt-dlp", args)

	cmd := exec.Command("yt-dlp", args...)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Println(err.Error())
		return
	}

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				output := string(buf[:n])
				fmt.Print(output)
			}
			if err != nil {
				break
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		fmt.Println(err.Error())
		return
	}
}
