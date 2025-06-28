package config

import (
	"fmt"

	"github.com/link00000000/fredboard/v3/internal/optional"
)

type unvalidatedAudioConfig struct {
	NumChannels  optional.Optional[int]
	SampleRateHz optional.Optional[int]
	BitrateKbps  optional.Optional[int]
}

type unvalidatedDiscordConfig struct {
	AppId     optional.Optional[string]
	PublicKey optional.Optional[string]
	Token     optional.Optional[string]
}

type unvalidatedLoggingHandlerConfig struct {
	Type   optional.Optional[LoggingHandlerType]
	Level  optional.Optional[LoggingHandlerLevel]
	Output optional.Optional[string]
}

type unvalidatedLoggingConfig struct {
	Handlers []unvalidatedLoggingHandlerConfig
}

type unvalidatedWebConfig struct {
	Address optional.Optional[string]
}

type unvalidatedYtdlpConfig struct {
	CookiesFile optional.Optional[string]
	ExePath     optional.Optional[string]
}

type unvalidatedFfmpegConfig struct {
	ExePath optional.Optional[string]
}

type unvalidatedConfig struct {
	Audio   optional.Optional[unvalidatedAudioConfig]
	Discord optional.Optional[unvalidatedDiscordConfig]
	Logging optional.Optional[unvalidatedLoggingConfig]
	Web     optional.Optional[unvalidatedWebConfig]
	Ytdlp   optional.Optional[unvalidatedYtdlpConfig]
	Ffmpeg  optional.Optional[unvalidatedFfmpegConfig]
}

type ConfigurationValidationError struct {
	Option  string
	Message string
}

func NewConfigurationValidationError(option string, message string) ConfigurationValidationError {
	return ConfigurationValidationError{Option: option, Message: message}
}

func (e ConfigurationValidationError) Error() string {
	return "invalid configuration for option " + e.Option + ": " + e.Message
}

// TODO: better validation that checks that the set values will work with eachother
func (c unvalidatedAudioConfig) validate() (cfg AudioConfig, errs []ConfigurationValidationError) {
	errs = make([]ConfigurationValidationError, 0)

	switch {
	case !c.NumChannels.IsSet():
		errs = append(errs, NewConfigurationValidationError("audio.numChannels", "required option is not set"))
	case c.NumChannels.Get() < 2:
		errs = append(errs, NewConfigurationValidationError("audio.numchannels", "invalid value. minimum of 2"))
	default:
		cfg.NumChannels = c.NumChannels.Get()
	}

	switch {
	case !c.SampleRateHz.IsSet():
		errs = append(errs, NewConfigurationValidationError("audio.sampleRateHz", "required option is not set"))
	case c.SampleRateHz.Get() <= 0:
		errs = append(errs, NewConfigurationValidationError("audio.sampleRateHz", "invalid value (must be greater than 0)"))
	default:
		cfg.SampleRateHz = c.SampleRateHz.Get()
	}

	switch {
	case !c.BitrateKbps.IsSet():
		errs = append(errs, NewConfigurationValidationError("audio.bitrateKbps", "required option is not set"))
	case c.BitrateKbps.Get() <= 0:
		errs = append(errs, NewConfigurationValidationError("audio.bitrateKbps", "invalid value (must be greater than 0)"))
	default:
		cfg.BitrateKbps = c.BitrateKbps.Get()
	}

	return cfg, errs
}

func (c unvalidatedDiscordConfig) validate() (cfg DiscordConfig, errs []ConfigurationValidationError) {
	errs = make([]ConfigurationValidationError, 0)

	switch {
	case !c.AppId.IsSet(), c.AppId.Get() == "":
		errs = append(errs, NewConfigurationValidationError("discord.appId", "required option is not set"))
	default:
		cfg.AppId = c.AppId.Get()
	}

	switch {
	case !c.PublicKey.IsSet(), c.PublicKey.Get() == "":
		errs = append(errs, NewConfigurationValidationError("discord.publicKey", "required option is not set"))
	default:
		cfg.PublicKey = c.PublicKey.Get()
	}

	switch {
	case !c.Token.IsSet(), c.Token.Get() == "":
		errs = append(errs, NewConfigurationValidationError("discord.Token", "required option is not set"))
	default:
		cfg.Token = c.Token.Get()
	}

	return cfg, errs
}

func (c unvalidatedLoggingHandlerConfig) validate(idx int) (cfg LoggingHandlerConfig, errs []ConfigurationValidationError) {
	errs = make([]ConfigurationValidationError, 0)

	switch {
	case !c.Type.IsSet():
		errs = append(errs, NewConfigurationValidationError(fmt.Sprintf("logging.handlers[%d].type", idx), "required option is not set"))
	case
		c.Type.Get() == LoggingHandlerType_JSON,
		c.Type.Get() == LoggingHandlerType_Pretty:
		cfg.Type = c.Type.Get()
	default:
		errs = append(errs, NewConfigurationValidationError(fmt.Sprintf("logging.handlers[%d].type", idx), "invalid handler type"))
	}

	switch {
	case !c.Level.IsSet():
		errs = append(errs, NewConfigurationValidationError(fmt.Sprintf("logging.handlers[%d].level", idx), "required option is not set"))
	case
		c.Level.Get() == LoggingHandlerLevel_Debug,
		c.Level.Get() == LoggingHandlerLevel_Info,
		c.Level.Get() == LoggingHandlerLevel_Warn,
		c.Level.Get() == LoggingHandlerLevel_Error,
		c.Level.Get() == LoggingHandlerLevel_Fatal,
		c.Level.Get() == LoggingHandlerLevel_Panic:
		cfg.Level = c.Level.Get()
	default:
		errs = append(errs, NewConfigurationValidationError(fmt.Sprintf("logging.handlers[%d].level", idx), "invalid level"))
	}

	switch {
	case !c.Output.IsSet(), c.Output.Get() == "":
		errs = append(errs, NewConfigurationValidationError(fmt.Sprintf("logging.handlers[%d].output", idx), "required option is not set"))
	default:
		cfg.Output = c.Output.Get()
	}

	return cfg, errs
}

func (c unvalidatedLoggingConfig) validate() (cfg LoggingConfig, errs []ConfigurationValidationError) {
	errs = make([]ConfigurationValidationError, 0)

	for i, h := range c.Handlers {
		hCfg, err := h.validate(i)
		cfg.Handlers = append(cfg.Handlers, hCfg)

		if err != nil {
			errs = append(errs, err...)
		}
	}

	return cfg, errs
}

func (c unvalidatedWebConfig) validate() (cfg WebConfig, errs []ConfigurationValidationError) {
	errs = make([]ConfigurationValidationError, 0)

	switch {
	case !c.Address.IsSet() || c.Address.Get() == "":
		errs = append(errs, NewConfigurationValidationError("web.address", "required option is not set"))
	default:
		cfg.Address = c.Address.Get()
	}

	return cfg, errs
}

func (c unvalidatedYtdlpConfig) validate() (cfg YtdlpConfig, errs []ConfigurationValidationError) {
	errs = make([]ConfigurationValidationError, 0)

	cfg.CookiesFile = c.CookiesFile
	cfg.ExePath = c.ExePath

	return cfg, errs
}

func (c unvalidatedFfmpegConfig) validate() (cfg FfmpegConfig, errs []ConfigurationValidationError) {
	errs = make([]ConfigurationValidationError, 0)

	cfg.ExePath = c.ExePath

	return cfg, errs
}

func validate(uCfg unvalidatedConfig) (cfg Config, errs []ConfigurationValidationError) {
	var verrs []ConfigurationValidationError

	if !uCfg.Audio.IsSet() {
		uCfg.Audio = optional.Make(unvalidatedAudioConfig{})
	}

	if cfg.Audio, verrs = uCfg.Audio.Get().validate(); len(verrs) > 0 {
		errs = append(errs, verrs...)
	}

	if !uCfg.Discord.IsSet() {
		uCfg.Discord = optional.Make(unvalidatedDiscordConfig{})
	}

	if cfg.Discord, verrs = uCfg.Discord.Get().validate(); len(verrs) > 0 {
		errs = append(errs, verrs...)
	}

	if !uCfg.Logging.IsSet() {
		uCfg.Logging = optional.Make(unvalidatedLoggingConfig{})
	}

	if cfg.Logging, verrs = uCfg.Logging.Get().validate(); len(verrs) > 0 {
		errs = append(errs, verrs...)
	}

	if !uCfg.Web.IsSet() {
		uCfg.Web = optional.Make(unvalidatedWebConfig{})
	}

	if cfg.Web, verrs = uCfg.Web.Get().validate(); len(verrs) > 0 {
		errs = append(errs, verrs...)
	}

	if !uCfg.Ytdlp.IsSet() {
		uCfg.Ytdlp = optional.Make(unvalidatedYtdlpConfig{})
	}

	if cfg.Ytdlp, verrs = uCfg.Ytdlp.Get().validate(); len(verrs) > 0 {
		errs = append(errs, verrs...)
	}

	if !uCfg.Ffmpeg.IsSet() {
		uCfg.Ffmpeg = optional.Make(unvalidatedFfmpegConfig{})
	}

	if cfg.Ffmpeg, verrs = uCfg.Ffmpeg.Get().validate(); len(verrs) > 0 {
		errs = append(errs, verrs...)
	}

	if len(errs) > 0 {
		return Config{}, errs
	}

	return cfg, []ConfigurationValidationError{}
}
