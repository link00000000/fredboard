package setup

import (
	"testing"

	"github.com/link00000000/fredboard/v3/internal/telemetry/logging"
)

func SetupLogger(t *testing.T) *logging.Logger {
	return logging.NewLogger()
}
