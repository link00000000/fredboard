//go:build gui_sdl

package routines

import (
	"accidentallycoded.com/fredboard/v3/internal/syncext"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/sdlbackend"
)

func NewUIRoutine(logger *logging.Logger, name string) syncext.Routine {
	factory := func() backend.Backend[sdlbackend.SDLWindowFlags] { return sdlbackend.NewSDLBackend() }
	return newUIRoutine(logger, name, factory)
}
