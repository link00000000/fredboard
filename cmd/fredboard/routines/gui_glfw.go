//go:build gui_glfw

package routines

import (
	"accidentallycoded.com/fredboard/v3/internal/syncext"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
)

func NewUIRoutine(logger *logging.Logger, name string) syncext.Routine {
	factory := func() backend.Backend[glfwbackend.GLFWWindowFlags] { return glfwbackend.NewGLFWBackend() }
	return newUIRoutine(logger, name, factory)
}
