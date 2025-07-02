package setup

import (
	"testing"

	"github.com/link00000000/go-telemetry/logging"
)

func SetupLogger(t *testing.T) *logging.Logger {
	return logging.NewLogger()
}
