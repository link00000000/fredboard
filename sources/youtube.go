package sources

import (
	"fmt"
	"os/exec"
)

type YouTubeStreamQuality string

const (
	YOUTUBESTREAMQUALITY_WORST YouTubeStreamQuality = "worstaudio"
	YOUTUBESTREAMQUALITY_BEST                       = "bestaudio"
)

type YouTube struct {
  ytdlpCmd *exec.Cmd
  ffmpegCmd *exec.Cmd
}

func NewYouTubeSource(url string, quality YouTubeStreamQuality) (*YouTube, error) {
  ytdlpCmd := exec.Command("yt-dlp",
		"--abort-on-error",
		"--quiet",
		"--no-warnings",
		"--format", fmt.Sprintf("%s[acodec=opus]", quality),
		"--output", "-",
		url)

  ffmpegCmd := exec.Command("ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-i", "pipe:0",
		"-f", "s16le",
    "-ar", "48000",
    "-ac", "2",
		"pipe:1")

	if ytdlpStdout, err := ytdlpCmd.StdoutPipe(); err != nil {
		return nil, err
	} else {
		ffmpegCmd.Stdin = ytdlpStdout
	}

  if _, err := ffmpegCmd.StdoutPipe(); err != nil {
    return nil, err
  }

  // TODO: Print ytdlpCmd.Stderr
  // TODO: Print ffmpegCmd.Stderr

  return &YouTube{ytdlpCmd: ytdlpCmd, ffmpegCmd: ffmpegCmd}, nil
}

// Implements [io.Reader]
func (youtube *YouTube) Read(p []byte) (int, error) {
  return youtube.ffmpegCmd.Stdout.Write(p)
}

