package config

import (
	"github.com/link00000000/fredboard/v3/internal/optional"
)

func applyDefaults(cfg *unvalidatedConfig) {
	if !cfg.Audio.IsSet() {
		cfg.Audio = optional.Make(unvalidatedAudioConfig{})
	}

	if !cfg.Audio.Get().NumChannels.IsSet() {
		cfg.Audio.GetMut().NumChannels.Set(2)
	}

	if !cfg.Audio.Get().SampleRateHz.IsSet() {
		cfg.Audio.GetMut().SampleRateHz.Set(48000)
	}

	if !cfg.Audio.Get().BitrateKbps.IsSet() {
		cfg.Audio.GetMut().BitrateKbps.Set(64)
	}

	if !cfg.Logging.IsSet() {
		cfg.Logging.Set(unvalidatedLoggingConfig{})
	}

	if len(cfg.Logging.Get().Handlers) == 0 {
		hCfg := unvalidatedLoggingHandlerConfig{
			Type:   optional.Make(LoggingHandlerType_Pretty),
			Level:  optional.Make(LoggingHandlerLevel_Info),
			Output: optional.Make("stderr"),
		}

		cfg.Logging.GetMut().Handlers = append(cfg.Logging.Get().Handlers, hCfg)
	}

	if !cfg.Web.IsSet() {
		cfg.Web.Set(unvalidatedWebConfig{})
	}

	if !cfg.Web.Get().Address.IsSet() {
		cfg.Web.GetMut().Address.Set(":8080")
	}
}
