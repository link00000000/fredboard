package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"accidentallycoded.com/fredboard/v3/internal/optional"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

const (
	ffmpegExecutableName = "ffmpeg"

	Format_PCMSigned16BitLittleEndian = "s16le"
	Format_Ogg                        = "ogg"
)

type Config struct {
	ExePath optional.Optional[string]
}

func Exe(config Config) (exe string, err error) {
	if config.ExePath.IsSet() {
		return config.ExePath.Get(), nil
	}

	exe, err = exec.LookPath(ffmpegExecutableName)
	if err == nil || errors.Is(err, exec.ErrDot) {
		return exe, nil
	}

	return "", err
}

func NewEncodeCmd(ctx context.Context, config Config, format string, sampleRateHz, nAudioChannels int) (cmd *exec.Cmd, err error) {
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

type transcoder struct {
	cmd    *exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc
	stdout io.ReadCloser
	stderr []byte
	err    error
}

func (t *transcoder) Read(p []byte) (n int, err error) {
	return t.stdout.Read(p)
}

func (t *transcoder) Close() (err error) {
	t.cancel()
	return nil
}

func (r *transcoder) Err() error {
	return r.err
}

func NewTranscoder(
	logger *logging.Logger,
	config Config,
	r io.Reader,
	format string,
	sampleRateHz, nAudioChannels int,
) (*transcoder, error) {
	ctx, cancel := context.WithCancel(context.Background())
	t := &transcoder{ctx: ctx, cancel: cancel}

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

	t.cmd = exec.CommandContext(ctx, exe, args...)

	t.cmd.Stdin = r

	stdout, err := t.cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to created stdout pipe: %w", err)
	}
	t.stdout = stdout

	stderr, err := t.cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	stderr1, pw := io.Pipe()
	stderr2 := io.TeeReader(stderr, pw)
	go logger.LogReader(stderr1, logging.LevelDebug, "[ffmpeg stderr]: %s")
	go func() {
		var err error
		t.stderr, err = io.ReadAll(stderr2)
		if err != nil {
			t.err = fmt.Errorf("failed to buffer all of ffmpeg stderr: %w", err)
		}
	}()

	err = t.cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg cmd: %w", err)
	}

	return t, err
}
