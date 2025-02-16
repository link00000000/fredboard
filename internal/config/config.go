package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"accidentallycoded.com/fredboard/v3/internal/errors"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

type OptionError struct {
	Option  string
	Message string
}

type LoggingHandlerType string

const (
	LoggingHandlerType_JSON   LoggingHandlerType = "json"
	LoggingHandlerType_Pretty LoggingHandlerType = "pretty"
)

func NewOptionError(option string, message string) OptionError {
	return OptionError{Option: option, Message: message}
}

func (e OptionError) Error() string {
	return "error while configuring option " + e.Option + ": " + e.Message
}

type AudioSettings struct {
	NumChannels  *int `json:"numChannels"`
	SampleRateHz *int `json:"sampleRateHz"`
	BitrateKbps  *int `json:"bitrateKbps"`
}

func (s *AudioSettings) init() {
	if s.NumChannels == nil {
		v := 2
		s.NumChannels = &v
	}

	if s.SampleRateHz == nil {
		v := 48000
		s.SampleRateHz = &v
	}

	if s.BitrateKbps == nil {
		v := 64
		s.BitrateKbps = &v
	}
}

// TODO: better validation that checks that the set values will work with eachother
func (s *AudioSettings) validate() []error {
	errs := errors.NewErrorList()

	if *s.NumChannels < 2 {
		errs.Add(NewOptionError("audio.numchannels", "invalid value. minimum of 2"))
	}

	if *s.SampleRateHz < 1 {
		errs.Add(NewOptionError("audio.samplerateHz", "invalid value. minimum of 1"))
	}

	if *s.BitrateKbps < 1 {
		errs.Add(NewOptionError("audio.bitrateKbps", "invalid value. minimum of 1"))
	}

	return errs.Slice()
}

type DiscordSettings struct {
	AppId     *string `json:"appId"`
	PublicKey *string `json:"publicKey"`
	Token     *string `json:"token"`
}

func (s *DiscordSettings) init() {
}

func (s *DiscordSettings) validate() []error {
	errs := errors.NewErrorList()

	if s.AppId == nil || *s.AppId == "" {
		errs.Add(NewOptionError("discord.appid", "required option is not set"))
	}

	if s.PublicKey == nil || *s.PublicKey == "" {
		errs.Add(NewOptionError("discord.publickey", "required option is not set"))
	}

	if s.Token == nil || *s.Token == "" {
		errs.Add(NewOptionError("discord.token", "required option is not set"))
	}

	return errs.Slice()
}

type LoggerSettings struct {
	Handlers []*LogHandlerSettings `json:"handlers"`
}

func (s *LoggerSettings) init() {
	if len(s.Handlers) == 0 {
		handler := LoggingHandlerType_Pretty
		level := logging.LevelInfo
		output := "stdout"

		s.Handlers = append(s.Handlers, &LogHandlerSettings{
			Type:   &handler,
			Level:  &level,
			Output: &output,
		})

		return
	}

	for _, handlerSettings := range s.Handlers {
		handlerSettings.init()
	}
}

func (s *LoggerSettings) validate() []error {
	errs := errors.NewErrorList()

	for _, loggerSettings := range s.Handlers {
		if loggerSettings == nil {
			errs.Add(NewOptionError("logging.handlers.[]", "contains a null"))
			continue
		}

		errs.Add(loggerSettings.validate()...)
	}

	return errs.Slice()
}

type LogHandlerSettings struct {
	Type   *LoggingHandlerType `json:"type"`
	Level  *logging.Level      `json:"level"`
	Output *string             `json:"output"`
}

func (s *LogHandlerSettings) init() {
}

func (s *LogHandlerSettings) validate() []error {
	errs := errors.NewErrorList()

	if s.Type == nil {
		errs.Add(NewOptionError("logging.handlers.[].type", "handler is null"))
	} else {
		switch *s.Type {
		case LoggingHandlerType_JSON, LoggingHandlerType_Pretty:
			break
		default:
			errs.Add(NewOptionError("logging.handlers.[].type", fmt.Sprintf("invalid handler type \"%s\"", string(*s.Type))))
		}
	}

	if s.Level == nil {
		errs.Add(NewOptionError("Logging.handlers.[].level", "level is null"))
	} else {
		switch *s.Level {
		case logging.LevelDebug, logging.LevelInfo, logging.LevelWarn, logging.LevelError, logging.LevelFatal, logging.LevelPanic:
			break
		default:
			errs.Add(NewOptionError("logging.handlers.[].level", fmt.Sprintf("invalid level \"%s\"", string(*s.Level))))
		}
	}

	return errs.Slice()
}

type WebSettings struct {
	Address *string `json:"address"`
}

func (s *WebSettings) init() {
	if s.Address == nil {
		v := ":8080"
		s.Address = &v
	}
}

func (s *WebSettings) validate() []error {
	errs := errors.NewErrorList()

	if s.Address == nil || *s.Address == "" {
		errs.Add(NewOptionError("web.address", "required value is not set"))
	}

	return errs.Slice()
}

type YtdlpSettings struct {
	CookiesFile *string `json:"cookiesFile"`
}

func (s *YtdlpSettings) init() {
}

func (s *YtdlpSettings) validate() []error {
	return []error{}
}

type Settings struct {
	Audio   *AudioSettings   `json:"audio"`
	Discord *DiscordSettings `json:"discord"`
	Loggers *LoggerSettings  `json:"logging"`
	Web     *WebSettings     `json:"web"`
	Ytdlp   *YtdlpSettings   `json:"ytdlp"`
}

func (s *Settings) init() {
	if s.Audio == nil {
		s.Audio = &AudioSettings{}
	}
	s.Audio.init()

	if s.Discord == nil {
		s.Discord = &DiscordSettings{}
	}
	s.Discord.init()

	if s.Loggers == nil {
		s.Loggers = &LoggerSettings{}
	}
	s.Loggers.init()

	if s.Web == nil {
		s.Web = &WebSettings{}
	}
	s.Web.init()

	if s.Ytdlp == nil {
		s.Ytdlp = &YtdlpSettings{}
	}
	s.Ytdlp.init()
}

func (s *Settings) validate() []error {
	errs := errors.NewErrorList()

	errs.Add(s.Audio.validate()...)
	errs.Add(s.Discord.validate()...)
	errs.Add(s.Loggers.validate()...)
	errs.Add(s.Web.validate()...)
	errs.Add(s.Ytdlp.validate()...)

	return errs.Slice()
}

var settings = Settings{}

var initErrs errors.ErrorList

func Init() error {
	cwd, err := os.Getwd()

	if err != nil {
		return err
	}

	configFilePath := fmt.Sprintf("%s/config.json", cwd)
	if opt, ok := os.LookupEnv("FREDBOARD_CONFIG"); ok {
		configFilePath = opt
	}

	f, err := os.Open(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to open config file at %s: %w", configFilePath, err)
	}
	defer f.Close()

	configFileContents, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read config file at %s: %w", configFilePath, err)
	}

	err = json.Unmarshal(configFileContents, &settings)
	if err != nil {
		return fmt.Errorf("failed to parse config file at %s: %w", configFilePath, err)
	}

	settings.init()

	return nil
}

func Validate() (ok bool, err []error) {
	errs := errors.NewErrorList(initErrs.Slice()...)
	errs.Add(settings.validate()...)

	return !errs.Any(), errs.Slice()
}

func Get() Settings {
	return settings
}
