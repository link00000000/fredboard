package sources

import (
	"errors"
	"fmt"
	"io"
	"os/exec"

	"accidentallycoded.com/fredboard/v3/telemetry/logging"
)

type YouTubeStreamQuality string

type StdStreams struct {
	cmd *exec.Cmd

	stdin       io.Writer
	stdinLogger *logging.Logger

	stdout       io.Reader
	stdoutLogger *logging.Logger

	stderr       io.Reader
	stderrLogger *logging.Logger
}

// Implements [io.Closer]
func (streams *StdStreams) Close() error {
	errs := make([]error, 0, 3)

	if streams.stdinLogger != nil {
		err := streams.stderrLogger.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if streams.stdout != nil {
		err := streams.stdoutLogger.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if streams.stderr != nil {
		err := streams.stderrLogger.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func NewStdStreams(cmd *exec.Cmd, logger *logging.Logger) (*StdStreams, error) {
	stdStreams := &StdStreams{cmd: cmd}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdStreams.stdinLogger = logger.NewChildLogger()
	stdStreams.stdin = logging.LogWriter(stdin, stdStreams.stdinLogger, logging.Debug)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdStreams.stdoutLogger = logger.NewChildLogger()
	stdStreams.stdout = logging.LogReader(stdout, stdStreams.stdoutLogger, logging.Debug)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	stdStreams.stderrLogger = logger.NewChildLogger()
	stdStreams.stderr = logging.LogReader(stderr, stdStreams.stderrLogger, logging.Debug)

	return stdStreams, nil
}

const (
	YOUTUBESTREAMQUALITY_WORST YouTubeStreamQuality = "worstaudio"
	YOUTUBESTREAMQUALITY_BEST                       = "bestaudio"
)

type YouTube struct {
	ytdlp  *StdStreams
	ffmpeg *StdStreams
}

func NewYouTubeSource(url string, quality YouTubeStreamQuality, logger *logging.Logger) (*YouTube, error) {
	fmt.Println("----------------------------------------------------- 1")

	ytdlp, err := NewStdStreams(exec.Command("yt-dlp",
		"--abort-on-error",
		"--quiet",
		"--no-warnings",
		"--format", fmt.Sprintf("%s[acodec=opus]", quality),
		"--output", "-",
		url), logger)

	if err != nil {
		return nil, err
	}

	ffmpeg, err := NewStdStreams(exec.Command("ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-i", "pipe:0",
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"pipe:1"), logger)

	if err != nil {
		return nil, err
	}

	ffmpeg.cmd.Stdin = ytdlp.stdout

	return &YouTube{ytdlp, ffmpeg}, nil
}

// Implements [io.Reader]
func (youtube *YouTube) Read(p []byte) (int, error) {
	return youtube.ffmpeg.stdout.Read(p)
}
