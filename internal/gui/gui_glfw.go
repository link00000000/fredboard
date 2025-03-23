//go:build gui

package gui

import (
	"context"
	"runtime"

	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
)

func Run(ctx context.Context, logger *logging.Logger) error {
	runtime.LockOSThread()

	currentBackend, err := backend.CreateBackend(glfwbackend.NewGLFWBackend())
	if err != nil {
		return err
	}

	currentBackend.SetBgColor(imgui.NewVec4(0.45, 0.55, 0.6, 1.0))
	currentBackend.CreateWindow("FredBoard", 1200, 900)

	cerr := make(chan error)
	go func() {
		currentBackend.Run(func() {
			err := render()
			if err != nil {
				cerr <- err
			}
		})
	}()

	select {
	case err := <-cerr:
		currentBackend.SetShouldClose(true)
		return err
	case <-ctx.Done():
		currentBackend.SetShouldClose(true)
		return nil
	}
}

func render() error {
	imgui.Begin("main window")
	imgui.TextUnformatted("TESTING")
	imgui.End()

	return nil
}
