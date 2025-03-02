package config

import (
	"fmt"
	"io"
	"os"

	"accidentallycoded.com/fredboard/v3/internal/optional"
)

type (
	LoggingHandlerType  string
	LoggingHandlerLevel string
)

const (
	LoggingHandlerType_JSON   LoggingHandlerType = "json"
	LoggingHandlerType_Pretty LoggingHandlerType = "pretty"

	LoggingHandlerLevel_Debug LoggingHandlerLevel = "debug"
	LoggingHandlerLevel_Info  LoggingHandlerLevel = "info"
	LoggingHandlerLevel_Warn  LoggingHandlerLevel = "warn"
	LoggingHandlerLevel_Error LoggingHandlerLevel = "error"
	LoggingHandlerLevel_Fatal LoggingHandlerLevel = "fatal"
	LoggingHandlerLevel_Panic LoggingHandlerLevel = "panic"
)

var validatedConfig optional.Optional[Config]

type AudioConfig struct {
	NumChannels  int
	SampleRateHz int
	BitrateKbps  int
}

type DiscordConfig struct {
	AppId     string
	PublicKey string
	Token     string
}

type LoggingHandlerConfig struct {
	Type   LoggingHandlerType
	Level  LoggingHandlerLevel
	Output string
}

type LoggingConfig struct {
	Handlers []LoggingHandlerConfig
}

type WebConfig struct {
	Address string
}

type YtdlpConfig struct {
	CookiesFile optional.Optional[string]
	ExePath     optional.Optional[string]
}

type FfmpegConfig struct {
	ExePath optional.Optional[string]
}

type Config struct {
	Audio   AudioConfig
	Discord DiscordConfig
	Logging LoggingConfig
	Web     WebConfig
	Ytdlp   YtdlpConfig
	Ffmpeg  FfmpegConfig
}

type ConfigInitOptions struct {
	Files []string
}

func Init(initOptions ConfigInitOptions) (verrs []ConfigurationValidationError, err error) {
	// TODO: merge multiple configs from multiple sources instead of just using the first file
	configFile := initOptions.Files[0]

	f, err := os.Open(configFile)
	if err != nil {
		return []ConfigurationValidationError{}, fmt.Errorf("failed to open config file at %s: %w", configFile, err)
	}
	defer f.Close()

	configBytes, err := io.ReadAll(f)
	if err != nil {
		return []ConfigurationValidationError{}, fmt.Errorf("failed to read config file at %s: %w", configFile, err)
	}

	cfg, err := fromJson(configBytes)
	if err != nil {
		return []ConfigurationValidationError{}, fmt.Errorf("failed to parse json of config file at %s: %w", configFile, err)
	}

	applyDefaults(&cfg)

	vCfg, verrs := validate(cfg)
	if len(verrs) != 0 {
		return verrs, nil
	}

	validatedConfig.Set(vCfg)

	return []ConfigurationValidationError{}, nil
}

func Get() Config {
	if !validatedConfig.IsSet() {
		panic("tried to get config before it was initialized")
	}

	return validatedConfig.Get()
}
