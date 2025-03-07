package ytdlp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"accidentallycoded.com/fredboard/v3/internal/optional"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

type YtdlpAudioQuality string

const (
	ytdlpExecutableName = "yt-dlp"

	MetadataType_Playlist = "playlist"
	MetadataType_Video    = "video"

	YtdlpAudioQuality_WorstAudio YtdlpAudioQuality = "worstaudio"
	YtdlpAudioQuality_BestAudio                    = "bestaudio"
)

type Config struct {
	ExePath     optional.Optional[string]
	CookiesPath optional.Optional[string]
}

type Metadata struct {
	Type        string `json:"_type"`
	Title       string `json:"title"`
	Description string `json:"Description"`
	Thumbnails  []struct {
		Url    string `json:"url"`
		Height int    `json:"height"`
		Width  int    `json:"width"`
	} `json:"thumbnails"`

	// only if Type == "playlist"
	Entries []*struct {
		Id           string `json:"id"`
		Title        string `json:"title"`
		ThumbnailUrl string `json:"thumbnail"`
		Url          string `json:"webpage_url"`
	} `json:"entries"`
}

func Exe(config Config) (exe string, err error) {
	if config.ExePath.IsSet() {
		return config.ExePath.Get(), nil
	}

	exe, err = exec.LookPath(ytdlpExecutableName)
	if err == nil || errors.Is(err, exec.ErrDot) {
		return exe, nil
	}

	return "", err
}

func NewMetadataCmd(ctx context.Context, config Config, url string) (cmd *exec.Cmd, err error) {
	args := []string{
		url,
		"--quiet", "--verbose", // continue to log but log to stderr instead of stdout
		"--restrict-filenames", // restrict filenames to only ASCII characters
		"--abort-on-error",     // do not continue to download if there is an error
		"--dump-single-json",   // write metadata to stdout as JSON
	}

	if config.CookiesPath.IsSet() {
		args = append(args, "--cookies", config.CookiesPath.Get())
	}

	exe, err := Exe(config)
	if err != nil {
		return nil, fmt.Errorf("error while resolving yt-dlp executable path: %w", err)
	}

	return exec.CommandContext(ctx, exe, args...), nil
}

func NewVideoCmd(ctx context.Context, config Config, url string, quality YtdlpAudioQuality) (cmd *exec.Cmd, err error) {
	args := []string{
		url,
		"--quiet", "--verbose", // continue to log but log to stderr instead of stdout
		"--restrict-filenames", // restrict filenames to only ASCII characters
		"--abort-on-error",     // do not continue to download if there is an error
		"--extract-audio",
		"--audio-format", "wav", // force output to wav for further processing
		"--format", string(quality),
		"-o", "-", // output to stdout
	}

	if config.CookiesPath.IsSet() {
		args = append(args, "--cookies", config.CookiesPath.Get())
	}

	exe, err := Exe(config)
	if err != nil {
		return nil, err
	}

	return exec.CommandContext(ctx, exe, args...), nil
}

type videoReader struct {
	cmd    *exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc
	stdout io.ReadCloser
	stderr []byte
	err    error
}

func (r *videoReader) Read(p []byte) (n int, err error) {
	return r.stdout.Read(p)
}

func (r *videoReader) Close() error {
	r.cancel()
	return nil
}

func (r *videoReader) Err() error {
	return r.err
}

func NewVideoReader(logger *logging.Logger, config Config, url string, quality YtdlpAudioQuality) (*videoReader, error) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &videoReader{ctx: ctx, cancel: cancel, stderr: make([]byte, 0)}

	cmd, err := NewVideoCmd(ctx, config, url, quality)
	if err != nil {
		return nil, fmt.Errorf("failed to create VideoCmd: %w", err)
	}
	r.cmd = cmd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to created stdout pipe: %w", err)
	}
	r.stdout = stdout

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	stderr1, pw := io.Pipe()
	stderr2 := io.TeeReader(stderr, pw)
	go logger.LogReader(stderr1, logging.LevelDebug, "[ytdlp stderr]: %s")
	go func() {
		var err error
		r.stderr, err = io.ReadAll(stderr2)
		if err != nil {
			r.err = fmt.Errorf("failed to buffer all of ytdlp stderr: %w", err)
		}
	}()

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start ytdlp cmd: %w", err)
	}

	return r, err
}
