package ytdlp

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"accidentallycoded.com/fredboard/v3/internal/optional"
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
