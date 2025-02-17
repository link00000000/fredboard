package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	internal_errors "accidentallycoded.com/fredboard/v3/internal/errors"
	"accidentallycoded.com/fredboard/v3/internal/exec/ffmpeg"
	"accidentallycoded.com/fredboard/v3/internal/exec/ytdlp"
	"accidentallycoded.com/fredboard/v3/internal/optional"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

func main() {
	logger := logging.NewLogger()
	logger.SetPanicOnError(true)
	logger.AddHandler(logging.NewPrettyHandler(os.Stderr, logging.LevelDebug))
	defer logger.Close()

	l := logger.NewChildLogger()
	defer l.Close()

	f, err := os.Create("./output.pcms16le")
	if err != nil {
		logger.Fatal("error while creating output file", "error", err)
	}
	defer f.Close()

	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	err = DownloadAudio(ctx, l, "https://www.youtube.com/watch?v=OkktfeAR-Rk", f)
	if err != nil {
		logger.Panic("error while downloading audio", "error", err)
	}
}

func DownloadAudio(ctx context.Context, logger *logging.Logger, url string, w io.Writer) error {
	ytdlpCmd, err := ytdlp.NewVideoCmd(ctx, &ytdlp.Config{
		ExePath:     optional.Make("/nix/store/8bk9vw8bk10x3g0r60mp1yxinfbwx4gd-yt-dlp-2025.1.26/bin/yt-dlp"),
		CookiesPath: optional.Make("/home/logan/Source/link00000000/fredboard/.env/cookies.txt"),
	}, url)
	if err != nil {
		logger.Warn("error while executing ytdlp.NewVideoCmd", "error", err, "url", url)
		return fmt.Errorf("error while executing ytdlp.NewVideoCmd: %w", err)
	}

	ytdlpStdout, err := ytdlpCmd.StdoutPipe()
	if err != nil {
		logger.Warn("error while creating stdout pipe for subprocess yt-dlp", "error", err)
		return fmt.Errorf("error while creating stdout pipe for subprocess yt-dlp: %w", err)
	}

	ytdlpStderr, err := ytdlpCmd.StderrPipe()
	if err != nil {
		logger.Warn("error while creating stderr pipe for subprocess yt-dlp", "error", err)
		return fmt.Errorf("error while creating stderr pipe for subprocess yt-dlp: %w", err)
	}

	ffmpegCmd, err := ffmpeg.NewEncodeCmd(ctx, &ffmpeg.Config{
		ExePath: optional.Make("/nix/store/hdgkfddym117iib1w67dxayr54kp7b1s-ffmpeg-7.1-bin/bin/ffmpeg"),
	}, "ogg", 48000, 2)
	if err != nil {
		logger.Warn("error while executing ffmpeg.NewEncodeCmd", "error", err)
		return fmt.Errorf("error while executing ffmpeg.NewEncodeCmd: %w", err)
	}

	ffmpegStdout, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		logger.Warn("error while creating stdout pipe for subprocess ffmpeg", "error", err)
		return fmt.Errorf("error while creating stdout pipe for subprocess ffmpeg: %w", err)
	}

	ffmpegStderr, err := ffmpegCmd.StderrPipe()
	if err != nil {
		logger.Warn("error while creating stderr pipe for subprocess ffmpeg", "error", err)
		return fmt.Errorf("error while creating stderr pipe for subprocess ffmpeg: %w", err)
	}

	go func() {
		_, err := io.Copy(w, ffmpegStdout)
		if err != nil {
			logger.Warn("error while copying yt-dlp stdout to ffmpeg stdin: %w", "error", err)
		}
	}()

	go func() {
		err := logger.LogReader(ytdlpStderr, logging.LevelDebug, "[ytdlp stderr] %s")
		if err != nil {
			logger.Warn("error while copying yt-dlp stdout to ffmpeg stdin: %w", err)
		}
	}()

	go func() {
		err := logger.LogReader(ffmpegStderr, logging.LevelDebug, "[ffmpeg stderr] %s")
		if err != nil {
			logger.Warn("error while reading stderr from ffmpeg")
		}
	}()

	ffmpegCmd.Stdin = ytdlpStdout

	err = ytdlpCmd.Start()
	if err != nil {
		logger.Warn("error while starting yt-dlp subprocess", "error", err)
		return fmt.Errorf("error while starting yt-dlp subprocess: %w", err)
	}

	logger.Debug("started yt-dlp subprocess", "exe", ytdlpCmd.Path, "args", ytdlpCmd.Args)

	err = ffmpegCmd.Start()
	if err != nil {
		logger.Warn("error while starting ffmpeg subprocess", "error", err)
		fmt.Errorf("error while starting ffmpeg subprocess: %w", err)
	}

	logger.Debug("started ffmpeg subprocess", "exe", ffmpegCmd.Path, "args", ffmpegCmd.Args)

	exitErrs := internal_errors.NewErrorList()

	err = ytdlpCmd.Wait()
	if err != nil {
		logger.Warn("error while waiting for yt-dlp subprocess", "error", err)
		exitErrs.Add(fmt.Errorf("error while waiting for yt-dlp subprocess: %w", err))

		err = ffmpegCmd.Process.Kill()
		if err != nil {
			logger.Warn("error while killing ffmpeg process", "error", err)
			return fmt.Errorf("error while killing ffmpeg process: %w", err)
		}
	}

	err = ffmpegCmd.Wait()
	if err != nil {
		logger.Warn("error while waiting for ffmpeg process", "error", err)
		exitErrs.Add(fmt.Errorf("error while waiting for ffmpeg process: %w", err))
	}

	return exitErrs.Join()
}
