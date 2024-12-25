package sources

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"accidentallycoded.com/fredboard/v3/telemetry/logging"
)

type YouTubeStreamQuality string

const (
	YOUTUBESTREAMQUALITY_WORST YouTubeStreamQuality = "worstaudio"
	YOUTUBESTREAMQUALITY_BEST                       = "bestaudio"
)

func newYtdlpCmd(url string, quality YouTubeStreamQuality, logger *logging.Logger) (*exec.Cmd, context.Context, context.CancelFunc, error) {
	context, cancel := context.WithCancel(context.Background())
	ytdlp := exec.CommandContext(context, "yt-dlp",
		"--abort-on-error",
		"--quiet",
		"--no-warnings",
		"--format", fmt.Sprintf("%s[acodec=opus]", quality),
		"--output", "-",
		url)

	stderr, err := ytdlp.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	go func() {
		buf := make([]byte, 0, 0xffff)

		for {
			buf = buf[:0]
			_, err := stderr.Read(buf)

			if err == io.EOF || err == os.ErrClosed {
				return
			}

			if err != nil {
				logger.ErrorWithErr("failed to read from ytdlp stderr pipe", err)
				return
			}

			if len(buf) > 0 {
				logger.Debug(fmt.Sprintf("ytdlp stderr: %s", string(buf)))
			}
		}
	}()

	return ytdlp, context, cancel, nil
}

func newFfmpegCmd(logger *logging.Logger) (*exec.Cmd, context.Context, context.CancelFunc, error) {
	context, cancel := context.WithCancel(context.Background())
	ffmpeg := exec.CommandContext(context, "ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-i", "pipe:0",
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"pipe:1")

	stderr, err := ffmpeg.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	go func() {
		buf := make([]byte, 0, 0xffff)

		for {
			buf = buf[:0]
			_, err := stderr.Read(buf)

			if err == io.EOF || err == os.ErrClosed {
				return
			}

			if err != nil {
				logger.ErrorWithErr("failed to read from ffmpeg stderr pipe", err)
				return
			}

			if len(buf) > 0 {
				logger.Debug(fmt.Sprintf("ffmpeg stderr: %s", string(buf)))
			}
		}
	}()

	return ffmpeg, context, cancel, nil
}

type YouTube struct {
	ytdlp        *exec.Cmd
	ytdlpContext context.Context
	cancelYtdlp  context.CancelFunc

	ffmpeg        *exec.Cmd
	ffmpegContext context.Context
	cancelFfmpeg  context.CancelFunc

	ffmpegStdout io.Reader

	stopped bool
}

func NewYouTubeSource(url string, quality YouTubeStreamQuality, logger *logging.Logger) (*YouTube, error) {
	ytdlp, ytdlpContext, cancelYtdlp, err := newYtdlpCmd(url, quality, logger)
	if err != nil {
		return nil, err
	}

	ffmpeg, ffmpegContext, cancelFfmpeg, err := newFfmpegCmd(logger)
	if err != nil {
		return nil, err
	}

	ytdlpStdout, err := ytdlp.StdoutPipe()
	if err != nil {
		return nil, err
	}

	ffmpeg.Stdin = ytdlpStdout

	ffmpegStdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		return nil, err
	}

	return &YouTube{
		ytdlp,
		ytdlpContext,
		cancelYtdlp,
		ffmpeg,
		ffmpegContext,
		cancelFfmpeg,
		ffmpegStdout,
		false,
	}, nil
}

// Implements [Source]
func (youtube *YouTube) Read(p []byte) (int, error) {
	return youtube.ffmpegStdout.Read(p)
}

// Implements [Source]
func (youtube *YouTube) Start() error {
	err := youtube.ytdlp.Start()
	if err != nil {
		return err
	}

	err = youtube.ffmpeg.Start()
	if err != nil {
		return err
	}

	return nil
}

// Implements [Source]
func (youtube *YouTube) Stop() error {
	if !youtube.stopped {
		youtube.cancelFfmpeg()
		youtube.cancelYtdlp()
	}

	return nil
}

// Implements [Source]
func (youtube *YouTube) Wait() error {
	err := youtube.ytdlp.Wait()
	if err != nil {
		return err
	}

	err = youtube.ffmpeg.Wait()
	if err != nil {
		return err
	}

	return nil
}
