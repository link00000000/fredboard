package config

import (
	"encoding/json"

	"accidentallycoded.com/fredboard/v3/internal/optional"
)

type jsonAudioConfig struct {
	NumChannels  optional.Optional[int] `json:"numChannels"`
	SampleRateHz optional.Optional[int] `json:"sampleRateHz"`
	BitrateKbps  optional.Optional[int] `json:"bitrateKbps"`
}

func (c jsonAudioConfig) merge(cfg unvalidatedAudioConfig) unvalidatedAudioConfig {
	if !cfg.NumChannels.IsSet() && c.NumChannels.IsSet() {
		cfg.NumChannels.Set(c.NumChannels.Get())
	}

	if !cfg.SampleRateHz.IsSet() && c.SampleRateHz.IsSet() {
		cfg.SampleRateHz.Set(c.SampleRateHz.Get())
	}

	if !cfg.BitrateKbps.IsSet() && c.BitrateKbps.IsSet() {
		cfg.BitrateKbps.Set(c.BitrateKbps.Get())
	}

	return cfg
}

type jsonDiscordConfig struct {
	AppId     optional.Optional[string] `json:"appId"`
	PublicKey optional.Optional[string] `json:"publicKey"`
	Token     optional.Optional[string] `json:"token"`
}

func (c jsonDiscordConfig) merge(cfg unvalidatedDiscordConfig) unvalidatedDiscordConfig {
	if !cfg.AppId.IsSet() && c.AppId.IsSet() {
		cfg.AppId.Set(c.AppId.Get())
	}

	if !cfg.PublicKey.IsSet() && c.PublicKey.IsSet() {
		cfg.PublicKey.Set(c.PublicKey.Get())
	}

	if !cfg.Token.IsSet() && c.Token.IsSet() {
		cfg.Token.Set(c.Token.Get())
	}

	return cfg
}

type jsonInputLoggingHandlerConfig struct {
	Type   optional.Optional[LoggingHandlerType]  `json:"type"`
	Level  optional.Optional[LoggingHandlerLevel] `json:"level"`
	Output optional.Optional[string]              `json:"output"`
}

type jsonLoggingConfig struct {
	Handlers []jsonInputLoggingHandlerConfig `json:"handlers"`
}

func (c jsonLoggingConfig) merge(cfg unvalidatedLoggingConfig) unvalidatedLoggingConfig {
	for _, handler := range c.Handlers {
		cfg.Handlers = append(cfg.Handlers, unvalidatedLoggingHandlerConfig{Type: handler.Type, Level: handler.Level, Output: handler.Output})
	}

	return cfg
}

type jsonWebConfig struct {
	Address optional.Optional[string] `json:"address"`
}

func (c jsonWebConfig) merge(cfg unvalidatedWebConfig) unvalidatedWebConfig {
	if !cfg.Address.IsSet() && c.Address.IsSet() {
		cfg.Address.Set(c.Address.Get())
	}

	return cfg
}

type jsonYtdlpConfig struct {
	CookiesFile optional.Optional[string] `json:"cookiesFile"`
	ExePath     optional.Optional[string] `json:"exePath"`
}

func (c jsonYtdlpConfig) merge(cfg unvalidatedYtdlpConfig) unvalidatedYtdlpConfig {
	if !cfg.CookiesFile.IsSet() && c.CookiesFile.IsSet() {
		cfg.CookiesFile.Set(c.CookiesFile.Get())
	}

	if !cfg.ExePath.IsSet() && c.ExePath.IsSet() {
		cfg.ExePath.Set(c.ExePath.Get())
	}

	return cfg
}

type jsonFfmpegConfig struct {
	ExePath optional.Optional[string] `json:"exePath"`
}

func (c jsonFfmpegConfig) merge(cfg unvalidatedFfmpegConfig) unvalidatedFfmpegConfig {
	if !cfg.ExePath.IsSet() && c.ExePath.IsSet() {
		cfg.ExePath.Set(c.ExePath.Get())
	}

	return cfg
}

type jsonConfig struct {
	Audio   optional.Optional[jsonAudioConfig]   `json:"audio"`
	Discord optional.Optional[jsonDiscordConfig] `json:"discord"`
	Logging optional.Optional[jsonLoggingConfig] `json:"logging"`
	Web     optional.Optional[jsonWebConfig]     `json:"web"`
	Ytdlp   optional.Optional[jsonYtdlpConfig]   `json:"ytdlp"`
	Ffmpeg  optional.Optional[jsonFfmpegConfig]  `json:"ffmpeg"`
}

func fromJson(data []byte) (cfg unvalidatedConfig, err error) {
	var v jsonConfig

	err = json.Unmarshal(data, &v)
	if err != nil {
		return unvalidatedConfig{}, err
	}

	if v.Audio.IsSet() {
		if !cfg.Audio.IsSet() {
			cfg.Audio = optional.Make(unvalidatedAudioConfig{})
		}
		cfg.Audio.Set(v.Audio.Get().merge(cfg.Audio.Get()))
	}

	if v.Discord.IsSet() {
		if !cfg.Discord.IsSet() {
			cfg.Discord = optional.Make(unvalidatedDiscordConfig{})
		}
		cfg.Discord.Set(v.Discord.Get().merge(cfg.Discord.Get()))
	}

	if v.Logging.IsSet() {
		if !cfg.Logging.IsSet() {
			cfg.Logging = optional.Make(unvalidatedLoggingConfig{})
		}
		cfg.Logging.Set(v.Logging.Get().merge(cfg.Logging.Get()))
	}

	if v.Web.IsSet() {
		if !cfg.Web.IsSet() {
			cfg.Web = optional.Make(unvalidatedWebConfig{})
		}
		cfg.Web.Set(v.Web.Get().merge(cfg.Web.Get()))
	}

	if v.Ytdlp.IsSet() {
		if !cfg.Ytdlp.IsSet() {
			cfg.Ytdlp = optional.Make(unvalidatedYtdlpConfig{})
		}
		cfg.Ytdlp.Set(v.Ytdlp.Get().merge(cfg.Ytdlp.Get()))
	}

	if v.Ffmpeg.IsSet() {
		if !cfg.Ffmpeg.IsSet() {
			cfg.Ffmpeg = optional.Make(unvalidatedFfmpegConfig{})
		}
		cfg.Ffmpeg.Set(v.Ffmpeg.Get().merge(cfg.Ffmpeg.Get()))
	}

	return cfg, nil
}
