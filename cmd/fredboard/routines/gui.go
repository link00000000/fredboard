//go:build gui

package routines

import (
	"errors"
	"fmt"
	"runtime"

	"accidentallycoded.com/fredboard/v3/internal/events"
	"accidentallycoded.com/fredboard/v3/internal/syncext"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
)

type UIRoutine struct {
	id        syncext.RoutineId
	name      string
	logger    *logging.Logger
	term      chan bool
	backend   backend.Backend[glfwbackend.GLFWWindowFlags]
	errs      *syncext.SyncData[[]error]
	doneEvent *events.EventEmitter[struct{}]
}

func (r UIRoutine) Id() syncext.RoutineId {
	return r.id
}

func (r *UIRoutine) SetId(id syncext.RoutineId) {
	r.id = id
}

func (r UIRoutine) Name() string {
	return r.name
}

func (r UIRoutine) Status() string {
	return "TODO"
}

func (r *UIRoutine) Run() error {
	defer r.doneEvent.Broadcast(struct{}{})

	var err error
	r.backend, err = backend.CreateBackend(glfwbackend.NewGLFWBackend())
	r.logger.Debug("created GLFW backend", "backend", r.backend)

	if err != nil {
		return fmt.Errorf("failed to create gui backend: %w", err)
	}

	err = initGui()
	if err != nil {
		return fmt.Errorf("failed to init gui: %w", err)
	}

	uiWindowDestroyed := make(chan struct{})
	go func() {
		runtime.LockOSThread()

		r.backend.SetBgColor(imgui.NewVec4(0, 0, 0, 1.0))
		r.backend.CreateWindow("FredBoard", 1200, 900)

		r.backend.SetCloseCallback(func() {
			r.logger.Debug("UI closed, UIRoutine is requesting to terminate all routines")
			r.addError(syncext.ErrRequestTermAllRoutines)
		})

		r.backend.SetBeforeDestroyContextHook(func() {
			r.logger.Debug("UI destroyed")
			close(uiWindowDestroyed)
		})

		r.logger.Debug("starting UI rendering loop")
		r.backend.Run(r.renderLoop)
	}()

	for {
		select {
		case force := <-r.term:
			r.logger.Debug("UIRoutine recieved term request", "force", force)
			if force {
				r.logger.Debug("UIRoutine forcefully terminating")
				return errors.Join(r.getErrors()...)
			}
			r.destroyUI()
		case <-uiWindowDestroyed:
			err = deinitGui()
			if err != nil {
				r.addError(fmt.Errorf("failed to deinit gui: %w", err))
			}
			return errors.Join(r.getErrors()...)
		}
	}
}

func (r *UIRoutine) Wait() {
	done := make(chan struct{})
	handle := r.doneEvent.AddChan(done)
	<-done
	r.doneEvent.RemoveDelegate(handle)
}

func (r *UIRoutine) destroyUI() {
	r.logger.Debug("destroying UI")
	r.backend.SetShouldClose(true)
}

func (r *UIRoutine) renderLoop() {
	err := renderGui()

	if err != nil {
		r.addError(err)
		r.destroyUI()
	}
}

func (r *UIRoutine) addError(err error) {
	r.errs.Lock()
	defer r.errs.Unlock()

	r.errs.Data = append(r.errs.Data, err)
}

func (r *UIRoutine) getErrors() []error {
	r.errs.Lock()
	defer r.errs.Unlock()

	return r.errs.Data
}

func (r *UIRoutine) Terminate(force bool, requestedBy syncext.Routine) {
	r.term <- force
}

func NewUIRoutine(logger *logging.Logger, name string) syncext.Routine {
	return &UIRoutine{
		id:        0,
		name:      name,
		logger:    logger,
		term:      make(chan bool, 1),
		backend:   nil,
		errs:      syncext.NewSyncData(make([]error, 0)),
		doneEvent: events.NewEventEmitter[struct{}](),
	}
}

// TODO: Remove
func mainWindow() (err error) {
	if imgui.BeginMainMenuBar() {
		if imgui.BeginMenu("File") {
			if imgui.MenuItemBoolPtr("Quit", "q", nil) {
				err = syncext.ErrRequestTermAllRoutines
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
