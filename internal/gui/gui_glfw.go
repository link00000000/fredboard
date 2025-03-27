//go:build gui

package gui

import (
	"context"
	"errors"
	"runtime"

	"accidentallycoded.com/fredboard/v3/internal/syncext"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
)

var ErrRequestExit = errors.New("exit requested")

var currentBackend backend.Backend[glfwbackend.GLFWWindowFlags]
var lastCursorPos *imgui.Vec2

func Run(ctx context.Context, logger *logging.Logger) error {
	errs := syncext.NewSyncData(make([]error, 0))
	done := make(chan struct{})

	go func() {
		runtime.LockOSThread()

		var err error
		currentBackend, err = backend.CreateBackend(glfwbackend.NewGLFWBackend())

		if err != nil {
			errs.Do(func(errs *[]error) {
				*errs = append(*errs, err)
			})

			close(done)
		}

		go func() {
			<-ctx.Done()
			currentBackend.SetShouldClose(true)
		}()

		//currentBackend.SetWindowFlags(glfwbackend.GLFWWindowFlagsDecorated, 0)
		currentBackend.SetBgColor(imgui.NewVec4(0, 0, 0, 1.0))
		currentBackend.CreateWindow("FredBoard", 1200, 900)
		currentBackend.SetBeforeDestroyContextHook(func() { close(done) })

		currentBackend.Run(func() {
			err := mainWindow()

			if err != nil {
				if err != ErrRequestExit {
					errs.Do(func(errs *[]error) {
						*errs = append(*errs, err)
					})
				}

				currentBackend.SetShouldClose(true)
			}
		})
	}()

	<-done
	return nil
}

func mainWindow() (err error) {
	if imgui.BeginMainMenuBar() {
		if imgui.BeginMenu("File") {

			if imgui.MenuItemBoolPtr("Quit", "q", nil) {
				err = ErrRequestExit
			}

			imgui.EndMenu()
		}

		imgui.EndMainMenuBar()
	}

	menuSize := imgui.ItemRectSize()

	viewport := imgui.MainViewport()
	imgui.SetNextWindowPos(viewport.Pos().Add(imgui.Vec2{X: 0, Y: menuSize.Y}))
	imgui.SetNextWindowSize(viewport.Size().Sub(imgui.Vec2{X: 0, Y: menuSize.Y}))

	if imgui.BeginV("##main-window", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoNav) {
		imgui.TextUnformatted("this is a test")

		imgui.End()
	}

	return
}
