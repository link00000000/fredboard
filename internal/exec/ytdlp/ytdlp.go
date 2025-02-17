package ytdlp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	internal_errors "accidentallycoded.com/fredboard/v3/internal/errors"
	"accidentallycoded.com/fredboard/v3/internal/optional"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

const ytdlpExecutableName = "yt-dlp.exe"

const (
	MetadataType_Playlist = "playlist"
	MetadataType_Video    = "video"
)

type YtdlpAudioQuality string

const (
	YtdlpAudioQuality_WorstAudio YtdlpAudioQuality = "worstaudio"
	YtdlpAudioQuality_BestAudio                    = "bestaudio"
)

type Config struct {
	ExePath     optional.Optional[string]
	CookiesPath optional.Optional[string]
}

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

/*
type instance struct {
	logger         *logging.Logger
	url            string
	cmd            *exec.Cmd
	ctx            context.Context
	stdout, stderr io.ReadCloser
	stderrBuffer   []byte
	onDone         *events.EventEmitter[error]
}
*/

type instance struct {
	logger *logging.Logger
	cmd    *exec.Cmd
	ctx    context.Context
	errs   internal_errors.ErrorList
	stderr []string

	Stdin  io.WriteCloser
	Stdout io.ReadCloser
}

/*
func GetMetadata(ctx context.Context, logger *logging.Logger, config *Config, url string) (metadata *Metadata, err error) {
	inst, err := newInstance(ctx, logger, config, url, []string{"--dump-single-json"})
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
*/

func Exe(config *Config) (exe string, err error) {
	if config.ExePath.IsSet() {
		return config.ExePath.Get(), nil
	}

	exe, err = exec.LookPath(ytdlpExecutableName)
	if err == nil || errors.Is(err, exec.ErrDot) {
		return exe, nil
	}

	return "", err
}

func NewMetadataCmd(ctx context.Context, config *Config, url string) (cmd *exec.Cmd, err error) {
	if config == nil {
		config = &defaultConfig
	}

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

func NewVideoCmd(ctx context.Context, config *Config, url string) (cmd *exec.Cmd, err error) {
	if config == nil {
		config = &defaultConfig
	}

	args := []string{
		url,
		"--quiet", "--verbose", // continue to log but log to stderr instead of stdout
		"--restrict-filenames", // restrict filenames to only ASCII characters
		"--abort-on-error",     // do not continue to download if there is an error
		"-o", "-",              // output to stdout
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

func VideoInst(ctx context.Context, logger *logging.Logger, config *Config, url string, quality YtdlpAudioQuality) (inst *instance, err error) {
	args := []string{
		url,
		"--quiet", "--verbose", // continue to log but log to stderr instead of stdout
		"--restrict-filenames", // restrict filenames to only ASCII characters
		"--abort-on-error",     // do not continue to download if there is an error
		"-o", "-",              // output to stdout
		"--format", string(quality),
	}

	if config.CookiesPath.IsSet() {
		args = append(args, "--cookies", config.CookiesPath.Get())
	}

	exe, err := Exe(config)
	if err != nil {
		return nil, err
	}

	inst = &instance{
		logger: logger,
		cmd:    exec.CommandContext(ctx, exe, args...),
		ctx:    ctx,
		stderr: make([]string, 0),
	}

	stderrPipe, err := inst.cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe for yt-dlp subprocess: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			if scanner.Err() != nil {
				inst.errs.AddThreadSafe(scanner.Err())
				break
			}

			line := scanner.Text()
			inst.logger.Debug("yt-dlp stderr output", "message", line)
			inst.stderr = append(inst.stderr, line)
		}
	}()

	inst.Stdin, err = inst.cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe for yt-dlp subprocess: %w", err)
	}

	inst.Stdout, err = inst.cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe for yt-dlp subprocess: %w", err)
	}

	return inst, nil
}

/*
func Open(ctx context.Context, logger *logging.Logger, config *Config, url string) (r io.Reader, err error) {
	inst, err := newInstance(ctx, logger, config, url, []string{"-o", "-"})
	if err != nil {
		return nil, fmt.Errorf("failed to create new yt-dlp instance: %w", err)
	}

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
*/

func (inst *instance) Run() error {
	return inst.cmd.Run()
}

func (inst *instance) Err() error {
	return inst.errs.Join()
}

func (inst *instance) Stderr() []string {
	return inst.stderr
}

/*
func newInstance(ctx context.Context, logger *logging.Logger, config *Config, url string, additionalFlags []string) (inst *instance, err error) {
	if config == nil {
		config = &defaultConfig
	}

	var exePath string
	if config.ExePath.IsSet() {
		exePath = config.ExePath.Get()
	} else {
		logger.Debug("searching for yt-dlp in path", "executable", ytdlpExecutableName)

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

	logger.Debug("created yt-dlp subprocess", "executable", exePath, "args", args)

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
*/
