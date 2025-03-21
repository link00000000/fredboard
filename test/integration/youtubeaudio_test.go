//go:build integration

package integration_test

import (
	"testing"

	cfg "accidentallycoded.com/fredboard/v3/internal/config"
	"accidentallycoded.com/fredboard/v3/internal/exec/ytdlp"
	"accidentallycoded.com/fredboard/v3/internal/optional"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	test_setup "accidentallycoded.com/fredboard/v3/test/integration/setup"
)

var (
	config cfg.Settings
	logger *logging.Logger
)

func setup(t *testing.T) {
	config = test_setup.SetupConfig(t)
	logger = test_setup.SetupLogger(t)
}

func TestYoutubeAudioWriteToFile(t *testing.T) {
	setup(t)

	t.Logf("%", config)

	ytdlpConfig := ytdlp.Config{
		ExePath:     optional.Make(*config.Ytdlp.ExePath),
		CookiesPath: optional.Make(*config.Ytdlp.CookiesFile),
	}

	r, err := ytdlp.NewVideoReader(logger, &ytdlpConfig, "https://www.youtube.com/watch?v=jNQXAC9IVRw", ytdlp.YtdlpAudioQuality_BestAudio)
	if err != nil {
		t.Fatal("failed to create new video reader", err)
	}

	r.Close()
}
