package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"accidentallycoded.com/fredboard/v3/internal/optional"
	"accidentallycoded.com/fredboard/v3/internal/syncext"
	"accidentallycoded.com/fredboard/v3/internal/telemetry"
)

const (
	ffmpegExecutableName = "ffmpeg"

	Format_PCMSigned16BitLittleEndian = "s16le"
	Format_Ogg                        = "ogg"
)

type Config struct {
	ExePath optional.Optional[string]
}

func exe(config Config) (exe string, err error) {
	if config.ExePath.IsSet() {
		return config.ExePath.Get(), nil
	}

	exe, err = exec.LookPath(ffmpegExecutableName)
	if err == nil || errors.Is(err, exec.ErrDot) {
		return exe, nil
	}

	return "", err
}

type transcoder struct {
	cancel context.CancelFunc
	stdout io.ReadCloser
	err    syncext.SyncData[error]
}

func (t *transcoder) Read(p []byte) (n int, err error) {
	t.err.Lock()
	defer t.err.Unlock()

	if t.err.Data != nil {
		return 0, t.err.Data
	}

	return t.stdout.Read(p)
}

func (t *transcoder) Close() (err error) {
	t.cancel()
	return nil
}

func NewTranscoder(
	ctx context.Context,
	config Config,
	r io.Reader,
	format string,
	sampleRateHz, nAudioChannels int,
) (*transcoder, error, <-chan *exec.ExitError) {
	ctx, cancel := context.WithCancel(context.Background())
	t := &transcoder{cancel: cancel}

	args := []string{
		"-hide_banner", // supress the copyright and build information
		"-i", "pipe:0", // read from stdin
		"-f", format,
		"-ar", fmt.Sprintf("%d", sampleRateHz), // set the sample rate
		"-ac", fmt.Sprintf("%d", nAudioChannels), // set the number of audio channels
		"-y", // if outputting to a file and it exists, overrwite it
		"pipe:1",
	}

	exe, err := exe(config)
	if err != nil {
		return nil, fmt.Errorf("error while resolving ffmpeg executable path: %w", err), nil
	}

	cmd := exec.CommandContext(ctx, exe, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to created stdin pipe: %w", err), nil
	}

	go func() {
		n, err := io.Copy(stdin, r)
		telemetry.Logger.DebugContext(ctx, "copied bytes from reader to ffmpeg stdin", "n", n, "error", err)

		if err != nil {
			t.err.Lock()
			t.err.Data = errors.Join(t.err.Data, err)
			t.err.Unlock()
		}

		stdin.Close()
	}()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to created stdout pipe: %w", err), nil
	}
	t.stdout = stdout

	// TODO: Log stderr
	/*
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stderr pipe: %w", err), nil
		}

		stderrBytes := syncext.SyncData[[]byte]{Data: make([]byte, 0)}

		stderr1, pw := io.Pipe()
		stderr2 := io.TeeReader(stderr, pw)
		go logger.LogReader(stderr1, logging.LevelDebug, "[ffmpeg stderr]: %s")
		go func() {
			var err error
			stderrBytes.Lock()
			stderrBytes.Data, err = io.ReadAll(stderr2)
			stderrBytes.Unlock()

			if err != nil {
				t.err.Lock()
				t.err.Data = errors.Join(t.err.Data, fmt.Errorf("failed to buffer all of ffmpeg stderr: %w", err))
				t.err.Unlock()
			}
		}()
	*/

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg cmd: %w", err), nil
	}

	exit := make(chan *exec.ExitError, 1)

	go func() {
		defer close(exit)

		err := cmd.Wait()

		if err != nil {
			switch err := err.(type) {
			case *exec.ExitError:
				// TODO: return stderr
				/*
					stderrBytes.Lock()
					exit <- &exec.ExitError{ProcessState: err.ProcessState, Stderr: stderrBytes.Data}
					stderrBytes.Unlock()
				*/
				exit <- &exec.ExitError{}
			default:
				panic(err)
			}
		}
	}()

	return t, err, exit
}
