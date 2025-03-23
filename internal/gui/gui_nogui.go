//go:build !gui

package gui

import (
	"context"

	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

func Run(ctx context.Context, logger *logging.Logger) error {
	logger.Debug("program not compiled to include gui. compile with '-tags=gui'")
	return nil
}
