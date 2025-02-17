package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"accidentallycoded.com/fredboard/v3/internal/optional"
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
