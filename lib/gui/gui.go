package main

import "C"
import "github.com/AllenDang/cimgui-go/imgui"

//export Render
func Render() {
	if imgui.BeginMainMenuBar() {
		if imgui.BeginMenu("File") {
			if imgui.MenuItemBoolPtr("Quit", "q", nil) {
				// TODO
				return
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
}

func main() {
}
