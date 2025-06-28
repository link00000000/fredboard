package ytdlp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/link00000000/fredboard/v3/internal/optional"
	"github.com/link00000000/fredboard/v3/internal/syncext"
	"github.com/link00000000/telemetry/logging"
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

func exe(config Config) (exe string, err error) {
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

	exe, err := exe(config)
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

	exe, err := exe(config)
	if err != nil {
		return nil, err
	}

	return exec.CommandContext(ctx, exe, args...), nil
}

type videoReader struct {
	cancel context.CancelFunc
	stdout io.ReadCloser
	err    syncext.SyncData[error]
}

func (r *videoReader) Read(p []byte) (int, error) {
	r.err.Lock()
	defer r.err.Unlock()

	if r.err.Data != nil {
		return 0, r.err.Data
	}

	return r.stdout.Read(p)
}

func (r *videoReader) Close() error {
	r.cancel()
	return nil
}

func NewVideoReader(logger *logging.Logger, config Config, url string, quality YtdlpAudioQuality) (*videoReader, error, <-chan *exec.ExitError) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &videoReader{cancel: cancel}

	cmd, err := NewVideoCmd(ctx, config, url, quality)
	if err != nil {
		return nil, fmt.Errorf("failed to create VideoCmd: %w", err), nil
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to created stdout pipe: %w", err), nil
	}
	r.stdout = stdout

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err), nil
	}

	stderrBytes := syncext.SyncData[[]byte]{Data: make([]byte, 0)}

	stderr1, pw := io.Pipe()
	stderr2 := io.TeeReader(stderr, pw)
	go logger.LogReader(stderr1, logging.LevelDebug, "[ytdlp stderr]: %s")
	go func() {
		stderrBytes.Lock()
		stderrBytes.Data, err = io.ReadAll(stderr2)
		stderrBytes.Unlock()

		if err != nil {
			r.err.Lock()
			r.err.Data = errors.Join(r.err.Data, fmt.Errorf("failed to buffer all of ytdlp stderr: %w", err))
			r.err.Unlock()
		}
	}()

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start ytdlp cmd: %w", err), nil
	}

	exit := make(chan *exec.ExitError, 1)

	go func() {
		defer close(exit)

		err := cmd.Wait()

		if err != nil {
			switch err := err.(type) {
			case *exec.ExitError:
				stderrBytes.Lock()
				exit <- &exec.ExitError{ProcessState: err.ProcessState, Stderr: stderrBytes.Data}
				stderrBytes.Unlock()
			default:
				panic(err)
			}
		}
	}()

	return r, err, exit
}
