//go build: gui_glfw || gui_sdl

package routines

import (
	"errors"
	"runtime"

	"accidentallycoded.com/fredboard/v3/internal/events"
	"accidentallycoded.com/fredboard/v3/internal/syncext"
	"accidentallycoded.com/fredboard/v3/internal/telemetry/logging"
	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/imgui"
)

type UIRoutine[TBackendFlags ~int] struct {
	id             syncext.RoutineId
	name           string
	logger         *logging.Logger
	term           chan bool
	backend        backend.Backend[TBackendFlags]
	errs           *syncext.SyncData[[]error]
	doneEvent      *events.EventEmitter[struct{}]
	backendFactory func() backend.Backend[TBackendFlags]
}

func (r UIRoutine[TBackendFlags]) Id() syncext.RoutineId {
	return r.id
}

func (r *UIRoutine[TBackendFlags]) SetId(id syncext.RoutineId) {
	r.id = id
}

func (r UIRoutine[TBackendFlags]) Name() string {
	return r.name
}

func (r UIRoutine[TBackendFlags]) Status() string {
	return "TODO"
}

func (r *UIRoutine[TBackendFlags]) Run() error {
	defer r.doneEvent.Broadcast(struct{}{})

	var err error
	r.backend, err = backend.CreateBackend(r.backendFactory())
	r.logger.Debug("created GLFW backend", "backend", r.backend)

	if err != nil {
		return err
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
			return errors.Join(r.getErrors()...)
		}
	}
}

func (r *UIRoutine[TBackend]) Wait() {
	done := make(chan struct{})
	handle := r.doneEvent.AddChan(done)
	<-done
	r.doneEvent.RemoveDelegate(handle)
}

func (r *UIRoutine[TBackend]) destroyUI() {
	r.logger.Debug("destroying UI")
	r.backend.SetShouldClose(true)
}

func (r *UIRoutine[TBackend]) renderLoop() {
	err := mainWindow()

	if err != nil {
		r.addError(err)
		r.destroyUI()
	}
}

func (r *UIRoutine[TBackend]) addError(err error) {
	r.errs.Lock()
	defer r.errs.Unlock()

	r.errs.Data = append(r.errs.Data, err)
}

func (r *UIRoutine[TBackend]) getErrors() []error {
	r.errs.Lock()
	defer r.errs.Unlock()

	return r.errs.Data
}

func (r *UIRoutine[TBackend]) Terminate(force bool, requestedBy syncext.Routine) {
	r.term <- force
}

func newUIRoutine[TBackend ~int](logger *logging.Logger, name string, backendFactory func() backend.Backend[TBackend]) syncext.Routine {
	return &UIRoutine[TBackend]{
		id:             0,
		name:           name,
		logger:         logger,
		term:           make(chan bool, 1),
		backend:        nil,
		errs:           syncext.NewSyncData(make([]error, 0)),
		doneEvent:      events.NewEventEmitter[struct{}](),
		backendFactory: backendFactory,
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
