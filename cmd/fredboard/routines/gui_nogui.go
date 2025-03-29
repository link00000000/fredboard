//go:build !gui_glfw && !gui_sdl

package routines

import (
	"accidentallycoded.com/fredboard/v3/internal/syncext"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
)

func NewUIRoutine(logger *logging.Logger, name string) syncext.Routine {
	return syncext.NewBasicRoutine(name, func(term <-chan bool) error {
		logger.Error("attempted to start gui but this program was not built with gui enabled. recompile with -tags=gui")
		return nil
	})
}
