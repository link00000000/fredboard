package main

import (
	"fmt"
	"runtime"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
	_ "github.com/AllenDang/cimgui-go/immarkdown"
	_ "github.com/AllenDang/cimgui-go/imnodes"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	currentBackend, err := backend.CreateBackend(glfwbackend.NewGLFWBackend())
	if err != nil {
		panic(err)
	}

	currentBackend.SetBgColor(imgui.NewVec4(0.45, 0.55, 0.6, 1.0))

	//currentBackend.SetWindowFlags(glfwbackend.GLFWWindowFlagsDecorated, 0)
	currentBackend.CreateWindow("Hello from cimgui-go", 1200, 900)

	currentBackend.SetDropCallback(func(p []string) {
		fmt.Printf("drop triggered: %v", p)
	})

	currentBackend.SetCloseCallback(func() {
		fmt.Println("window is closing")
	})

	isInitialized := false
	currentBackend.Run(func() {
		imgui.ClearSizeCallbackPool()

		var (
			dockspaceId imgui.ID
			leftDockId  imgui.ID
			rightDockId imgui.ID
		)

		// initialize dockspace
		dockspaceId = imgui.DockSpaceOverViewportV(0, imgui.MainViewport(), imgui.DockNodeFlagsPassthruCentralNode, imgui.NewWindowClass())
		if !isInitialized {
			// clear existing layout
			imgui.InternalDockBuilderRemoveNode(dockspaceId)
			imgui.InternalDockBuilderAddNodeV(dockspaceId, imgui.DockNodeFlagsNone)

			// split workspace
			imgui.InternalDockBuilderSplitNode(dockspaceId, imgui.DirLeft, 0.25, &leftDockId, &rightDockId)

			// add windows to docks
			imgui.InternalDockBuilderDockWindow("Left 1", leftDockId)
			imgui.InternalDockBuilderDockWindow("Left 2", leftDockId)
			imgui.InternalDockBuilderDockWindow("Right 1", rightDockId)
			imgui.InternalDockBuilderDockWindow("Right 2", rightDockId)

			isInitialized = true
		}

		imgui.BeginMainMenuBar()
		if imgui.BeginMenu("testing") {
			imgui.EndMenu()
		}
		imgui.EndMainMenuBar()

		imgui.Begin("Left 1")
		imgui.Text("this is the left1 window")
		imgui.End()

		imgui.Begin("Left 2")
		imgui.Text("this is the left2 window")
		imgui.End()

		imgui.Begin("Right 1")
		imgui.Text("this is the right1 window")
		imgui.End()

		imgui.Begin("Right 2")
		imgui.Text("this is the right2 window")
		imgui.End()
	})
}
