package ffmpeg

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

const (
	ffmpegExecutableName = "ffmpeg"

	Format_PCMSigned16BitLittleEndian = "s16le"
)

type Config struct {
	ExePath optional.Optional[string]
}

var defaultConfig Config = Config{
	ExePath: optional.Empty[string](),
}

/*
type instance struct {
	logger         *logging.Logger
	cmd            *exec.Cmd
	ctx            context.Context
	stdin          io.WriteCloser
	stdout, stderr io.ReadCloser
	onDone         *events.EventEmitter[error]
}
*/

type instance struct {
	logger *logging.Logger
	cmd    *exec.Cmd
	ctx    context.Context
	errs   internal_errors.ErrorList
	stderr []string
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

type TransformResult struct {
	Stderr []byte
	Err    error
}

/*
func Transform(ctx context.Context, logger *logging.Logger, config *Config, sampleRateHz, nAudioChannels int) (w io.WriteCloser, r io.ReadCloser, done chan TransformResult, err error) {
	inst, err := newInstance(ctx, logger, config, r, sampleRateHz, nAudioChannels)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create new ffmpeg instance: %w", err)
	}

	err = inst.cmd.Start()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to start ffmpeg subprocess: %w", err)
	}

	go func() {
		var wg sync.WaitGroup
		errs := internal_errors.NewErrorList()

		// stdin
		go func() {
			wg.Add(1)
			defer wg.Done()

			_, err := io.Copy(inst.stdin, r)
			errs.AddThreadSafe(err)
		}()

		var stderr []byte

		// err
		go func() {
			wg.Add(1)
			defer wg.Done()

			stderr, err = io.ReadAll(inst.stderr)
			errs.AddThreadSafe(err)
		}()

		wg.Wait()

		err = inst.cmd.Wait()
		errs.AddThreadSafe(err)

		done <- TransformResult{stderr, errs.Join()}
	}()

	return inst.stdin, cancellablereader.New(ctx, inst.stdout), done, nil
}
*/

func Exe(config *Config) (exe string, err error) {
	if config.ExePath.IsSet() {
		return config.ExePath.Get(), nil
	}

	exe, err = exec.LookPath(ffmpegExecutableName)
	if err == nil || errors.Is(err, exec.ErrDot) {
		return exe, nil
	}

	return "", err
}

func NewEncodeCmd(ctx context.Context, config *Config, format string, sampleRateHz, nAudioChannels int) (cmd *exec.Cmd, err error) {
	if config == nil {
		config = &defaultConfig
	}

	args := []string{
		"-hide_banner", // supress the copyright and build information
		"-i", "pipe:0", // read from stdin
		"-f", format,
		"-ar", fmt.Sprintf("%d", sampleRateHz), // set the sample rate
		"-ac", fmt.Sprintf("%d", nAudioChannels), // set the number of audio channels
		"-y", // if outputting to a file and it exists, overrwite it
		"pipe:1",
	}

	exe, err := Exe(config)
	if err != nil {
		return nil, fmt.Errorf("error while resolving ffmpeg executable path: %w", err)
	}

	return exec.CommandContext(ctx, exe, args...), nil
}

func Encode(ctx context.Context, logger *logging.Logger, config *Config, sampleRateHz, nAudioChannels int) (inst *instance, err error) {
	if config == nil {
		config = &defaultConfig
	}

	var exePath string
	if config.ExePath.IsSet() {
		exePath = config.ExePath.Get()
	} else {
		logger.Debug("searching for ffmpeg in path", "executable", ffmpegExecutableName)

		exePath, err = exec.LookPath(ffmpegExecutableName)
		if err != nil && !errors.Is(err, exec.ErrDot) {
			return nil, err
		}
	}

	args := []string{
		// supress the copyright and build information
		"-hide_banner",

		// read from stdin
		"-i",
		"pipe:0",

		// use signed 16-bit PCM as the output format
		"-f",
		"s16le",

		// set the sample rate
		"-ar",
		fmt.Sprintf("%d", sampleRateHz),

		// set the number of audio channels
		"-ac",
		fmt.Sprintf("%d", nAudioChannels),

		// if outputting to a file and it exists, overrwite it
		"-y",

		// output to stdout
		"audio.pcms16le",
		//"pipe:1",
	}

	inst = &instance{
		logger: logger,
		cmd:    exec.CommandContext(ctx, exePath, args...),
		ctx:    ctx,
		stderr: make([]string, 0),
	}

	stderrPipe, err := inst.cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe for ffmpeg subprocess: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			if scanner.Err() != nil {
				inst.errs.AddThreadSafe(scanner.Err())
				break
			}

			line := scanner.Text()
			inst.logger.Debug("ffmpeg stderr output", "message", line)
			inst.stderr = append(inst.stderr, line)
		}
	}()

	inst.stdin, err = inst.cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe for ffmpeg subprocess: %w", err)
	}

	inst.stdout, err = inst.cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe for ffmpeg subprocess: %w", err)
	}

	return inst, nil
}

func (inst *instance) Run() error {
	return inst.cmd.Run()
}

func (inst *instance) Err() error {
	return inst.errs.Join()
}

func (inst *instance) Stderr() []string {
	return inst.stderr
}

func (inst *instance) Kill() error {
	return inst.cmd.Process.Kill()
}

/*
func newInstance(ctx context.Context, logger *logging.Logger, config *Config, r io.Reader, sampleRateHz, nAudioChannels int) (inst *instance, err error) {
	if config == nil {
		config = &defaultConfig
	}

	var exePath string
	if config.ExePath.IsSet() {
		exePath = config.ExePath.Get()
	} else {
		logger.Debug("searching for ffmpeg in path", "executable", ffmpegExecutableName)

		exePath, err = exec.LookPath(ffmpegExecutableName)
		if err != nil && !errors.Is(err, exec.ErrDot) {
			return nil, err
		}
	}

	args := []string{
		// supress the copyright and build information
		"-hide_banner",

		// read from stdin
		"-i",
		"pipe:0",

		// use signed 16-bit PCM as the output format
		"-f",
		"s16le",

		// set the sample rate
		"-ar",
		fmt.Sprintf("%d", sampleRateHz),

		// set the number of audio channels
		"-ac",
		fmt.Sprintf("%d", nAudioChannels),

		// output to stdout
		"pipe:1",
	}

	inst = &instance{
		logger: logger,
		cmd:    exec.CommandContext(ctx, exePath, args...),
		ctx:    ctx,

		onDone: events.NewEventEmitter[error](),
	}

	logger.Debug("created ffmpeg subprocess", "executable", exePath, "args", args)

	logger.Debug("creating stdin pipe")
	inst.stdin, err = inst.cmd.StdinPipe()
	if err != nil {
		logger.Error("failed to create stdin pipe", "error", err)
	}

	logger.Debug("creating stdout pipe")
	inst.stdout, err = inst.cmd.StdoutPipe()
	if err != nil {
		logger.Error("failed to create stdout pipe", "error", err)
	}

	logger.Debug("creating stderr pipe")
	inst.stderr, err = inst.cmd.StderrPipe()
	if err != nil {
		logger.Error("failed to create stderr pipe", "error", err)
	}

	return inst, nil
}

func (inst *instance) fail(err error) {
	logArgs := make([]any, 0)

	switch v := err.(type) {
	case *exec.ExitError:
		logArgs = append(logArgs, "error", err, "exitCode", v.ExitCode())
	default:
		logArgs = append(logArgs, "error", err)
	}

	inst.logger.Error("ffmpeg instance failed", logArgs...)
	inst.onDone.Broadcast(err)
}

func (inst *instance) success() {
	inst.logger.Debug("ffmpeg instance completed successfully")
	inst.onDone.Broadcast(nil)
}
*/
