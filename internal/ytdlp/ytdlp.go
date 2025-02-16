package ytdlp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"

	"accidentallycoded.com/fredboard/v3/internal/events"
	"accidentallycoded.com/fredboard/v3/internal/io/cancellablereader"
	"accidentallycoded.com/fredboard/v3/internal/optional"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

const ytdlpExecutableName = "yt-dlp.exe"

const (
	MetadataType_Playlist = "playlist"
	MetadataType_Video    = "video"
)

var defaultConfig Config = Config{
	ExePath:     optional.Empty[string](),
	CookiesPath: optional.Empty[string](),
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

type instance struct {
	logger         *logging.Logger
	url            string
	cmd            *exec.Cmd
	ctx            context.Context
	stdout, stderr io.ReadCloser
	stderrBuffer   []byte
	onDone         *events.EventEmitter[error]
}

type Config struct {
	ExePath     optional.Optional[string]
	CookiesPath optional.Optional[string]
}

func GetMetadata(logger *logging.Logger, url string, config *Config, ctx context.Context) (metadata *Metadata, err error) {
	inst, err := newInstance(logger, url, config, []string{"--dump-single-json"}, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create new yt-dlp instance: %w", err)
	}

	done := make(chan error)
	inst.onDone.AddChan(done)

	err = inst.cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start yt-dlp subprocess: %w", err)
	}

	go func() {
		decoder := json.NewDecoder(inst.stdout)
		err = decoder.Decode(&metadata)
		if err != nil {
			inst.fail(err)
			return
		}
	}()

	go func() {
		buf := make([]byte, 1024)

		for {
			n, err := inst.stderr.Read(buf)
			buf = buf[:n]

			inst.stderrBuffer = append(inst.stderrBuffer, buf...)

			if err == io.EOF {
				break
			}

			if err != nil {
				inst.fail(err)
				return
			}
		}
	}()

	go func() {
		err = inst.cmd.Wait()
		if err != nil {
			inst.fail(err)
			return
		}

		inst.success()
	}()

	inst.logger.Info("downloading metadata", "url", inst.url)
	err = <-done

	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func Open(logger *logging.Logger, url string, config *Config, ctx context.Context) (reader io.Reader, err error) {
	inst, err := newInstance(logger, url, config, []string{"-o", "-"}, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create new yt-dlp instance: %w", err)
	}

	done := make(chan error)
	inst.onDone.AddChan(done)

	err = inst.cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start yt-dlp subprocess: %w", err)
	}

	go func() {
		buf := make([]byte, 1024)

		for {
			n, err := inst.stderr.Read(buf)
			buf = buf[:n]

			inst.stderrBuffer = append(inst.stderrBuffer, buf...)

			if err == io.EOF {
				break
			}

			if err != nil {
				inst.fail(err)
				return
			}
		}
	}()

	go func() {
		err = inst.cmd.Wait()
		if err != nil {
			inst.fail(err)
			return
		}

		inst.success()
	}()

	return cancellablereader.New(ctx, inst.stdout), nil
}

func newInstance(logger *logging.Logger, url string, config *Config, additionalFlags []string, ctx context.Context) (inst *instance, err error) {
	if config == nil {
		config = &defaultConfig
	}

	var exePath string
	if config.ExePath.IsSet() {
		exePath = config.ExePath.Get()
	} else {
		logger.Debug("searching for yt-dlp in path", slog.String("executable", ytdlpExecutableName))

		exePath, err = exec.LookPath(ytdlpExecutableName)
		if err != nil && !errors.Is(err, exec.ErrDot) {
			return nil, err
		}
	}

	args := []string{
		url,
		"--quiet",
		"--verbose",            // continue to log but log to stderr instead of stdout
		"--restrict-filenames", // restrict filenames to only ASCII characters
		"--abort-on-error",     // do not continue to download if there is an error
	}

	if config.CookiesPath.IsSet() {
		args = append(args, "--cookies", config.CookiesPath.Get())
	}

	args = append(args, additionalFlags...)

	inst = &instance{
		logger: logger,
		url:    url,
		cmd:    exec.CommandContext(ctx, exePath, args...),
		ctx:    ctx,

		onDone: events.NewEventEmitter[error](),
	}

	logger.Debug("created yt-dlp child process", "executable", exePath, "args", args)

	logger.Debug("creating stdout pipe")
	inst.stdout, err = inst.cmd.StdoutPipe()
	if err != nil {
		logger.Error("failed to create stdout pipe", "error", err)
		return nil, err
	}

	logger.Debug("creating stderr pipe")
	inst.stderr, err = inst.cmd.StderrPipe()
	if err != nil {
		logger.Error("failed to create stderr pipe", "error", err)
		return nil, err
	}

	return inst, nil
}

func (inst *instance) fail(err error) {
	logArgs := make([]any, 0)

	switch v := err.(type) {
	case *exec.ExitError:
		logArgs = append(logArgs, "error", err, "stderr", string(inst.stderrBuffer), "exitCode", v.ExitCode())
	default:
		logArgs = append(logArgs, "error", err, "stderr", string(inst.stderrBuffer))
	}

	inst.logger.Error("yt-dlp instance failed", logArgs...)
	inst.onDone.Broadcast(err)
}

func (inst *instance) success() {
	inst.logger.Debug("yt-dlp instance completed successfully")
	inst.onDone.Broadcast(nil)
}
