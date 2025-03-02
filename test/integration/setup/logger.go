package setup

import (
	"testing"

	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

func SetupLogger(t *testing.T) *logging.Logger {
	return logging.NewLogger()
}
