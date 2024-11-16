package config

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type OptionError struct {
	Option  string
	Message string
}

func NewOptionError(option string, message string) OptionError {
	return OptionError{Option: option, Message: message}
}

func (e OptionError) Error() string {
	return "error while configuring option " + e.Option + ": " + e.Message
}

var initErrors []error

var Config struct {
	Audio struct {
		NumChannels  int
		SampleRateHz int
		BitrateKbps  int
	}
	Discord struct {
		AppId     string
		PublicKey string
		Token     string
	}
	Logging struct {
		Level slog.Level
	}
}

func Init() {
	initErrors = make([]error, 0)

	Config.Audio.NumChannels = 2
	if opt, ok := os.LookupEnv("FREDBOARD_AUDIO_NUM_CHANNELS"); ok {
		if i, err := strconv.Atoi(opt); err != nil {
			initErrors = append(initErrors, NewOptionError("Audio.NumChannels", err.Error()))
		} else {
			Config.Audio.NumChannels = i
		}
	}

	Config.Audio.SampleRateHz = 48000
	if opt, ok := os.LookupEnv("FREDBOARD_AUDIO_SAMPLE_RATE_HZ"); ok {
		if i, err := strconv.Atoi(opt); err != nil {
			initErrors = append(initErrors, NewOptionError("Audio.SampleRateHz", err.Error()))
		} else {
			Config.Audio.SampleRateHz = i
		}
	}

	Config.Audio.BitrateKbps = 64
	if opt, ok := os.LookupEnv("FREDBOARD_AUDIO_BITRATE_KBPS"); ok {
		if i, err := strconv.Atoi(opt); err != nil {
			initErrors = append(initErrors, NewOptionError("Audio.BitrateKbps", err.Error()))
		} else {
			Config.Audio.BitrateKbps = i
		}
	}

	Config.Discord.AppId, _ = os.LookupEnv("FREDBOARD_DISCORD_APP_ID")
	Config.Discord.PublicKey, _ = os.LookupEnv("FREDBOARD_DISCORD_PUBLIC_KEY")
	Config.Discord.Token, _ = os.LookupEnv("FREDBOARD_DISCORD_TOKEN")

	Config.Logging.Level = slog.LevelInfo
	if opt, ok := os.LookupEnv("FREDBOARD_LOG_LEVEL"); ok {
		switch strings.ToUpper(opt) {
		case "ERROR":
			Config.Logging.Level = slog.LevelInfo.Level()
		case "WARN":
			Config.Logging.Level = slog.LevelWarn.Level()
		case "INFO":
			Config.Logging.Level = slog.LevelInfo.Level()
		case "DEBUG":
			Config.Logging.Level = slog.LevelDebug.Level()
		default:
			initErrors = append(initErrors, NewOptionError("Logging.Level", "invalid option value, allowed values are ERROR, WARN, INFO, DEBUG"))
		}
	}
}

func IsValid() (bool, error) {
	errs := initErrors[:]

	if len(Config.Discord.AppId) == 0 {
		errs = append(errs, NewOptionError("Discord.AppId", "required option not set"))
	}

	if len(Config.Discord.PublicKey) == 0 {
		errs = append(errs, NewOptionError("Discord.PublicKey", "required option not set"))
	}

	if len(Config.Discord.Token) == 0 {
		errs = append(errs, NewOptionError("Discord.Token", "required option not set"))
	}

	if len(errs) > 0 {
		return false, errors.Join(errs...)
	}

	return true, nil
}
