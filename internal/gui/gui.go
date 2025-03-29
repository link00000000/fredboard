package gui

import (
	"accidentallycoded.com/fredboard/v3/internal/syncext"
	"github.com/AllenDang/cimgui-go/imgui"
)

func Draw() error {
	if err := MainMenuBar(); err != nil {
		return err
	}

	mainMenuSize := imgui.ItemRectSize()
	viewport := imgui.MainViewport()
	imgui.SetNextWindowPos(viewport.Pos().Add(imgui.Vec2{X: 0, Y: mainMenuSize.Y}))
	imgui.SetNextWindowSize(viewport.Size().Sub(imgui.Vec2{X: 0, Y: mainMenuSize.Y}))
	if err := MainWindow(); err != nil {
		return err
	}

	return nil
}

func MainMenuBar() error {
	if imgui.BeginMainMenuBar() {
		if imgui.BeginMenu("File") {
			if imgui.MenuItemBoolPtr("Quit", "q", nil) {
				return syncext.ErrRequestTermAllRoutines
			}

			imgui.EndMenu()
		}

		imgui.EndMainMenuBar()
	}

	return nil
}

func MainWindow() error {
	if imgui.BeginV("##main-window", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoNav) {
		imgui.TextUnformatted("this is a test")

		imgui.End()
	}

	return nil
}
