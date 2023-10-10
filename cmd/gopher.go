package main

import (
	"log"
	"runtime"
	"time"

	"Gopher3D/renderer"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	imgui "github.com/AllenDang/cimgui-go"
)

const glsl_version string = "#version 410"

var width, height int32 = 800, 600                                 // Initialize to the center of the window
var lastX, lastY float64 = float64(width / 2), float64(height / 2) // Initialize to the center of the window
var firstMouse bool = true
var camera renderer.Camera
var refreshRate time.Duration = 1000 / 144 // 144 FPS

func main() {
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		log.Fatalf("Could not initialize glfw: %v", err)
	}
	defer glfw.Terminate()

	window, err := glfw.CreateWindow(int(width), int(height), "Gopher 3D", nil, nil)
	if err != nil {
		log.Fatalf("Could not create glfw window: %v", err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		log.Fatalf("Could not initialize OpenGL: %v", err)
	}

	renderer.Init(width, height)
	model := renderer.LoadObject()
	renderer.AddModel(model)

	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	// Initialize ImGui context
	context := imgui.CreateContext()
	defer context.Destroy()

	if err != nil {
		log.Fatalf("Failed to initialize GLFW //platform  // TODO: Adjust this reference: %v", err)
	}

	camera = renderer.NewCamera(width, height) // Initialize the global camera variable

	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled) // Hide and capture the cursor
	window.SetCursorPosCallback(mouseCallback)                // Set the callback function for mouse movement

	var lastTime = glfw.GetTime()
	for !window.ShouldClose() {
		currentTime := glfw.GetTime()
		deltaTime := currentTime - lastTime
		lastTime = currentTime
		camera.ProcessKeyboard(window, float32(deltaTime))

		// Start ImGui frame
		// TODO: Adjust this line if necessary
		//imgui.NewFrame()

		// TODO: Add ImGui UI elements here or call a separate function

		// Render ImGui
		imgui.Render()
		// TODO: Correct Rendering Call and use renderer.Render
		renderer.Render(camera, deltaTime) // Pass the dereferenced camera object to Render

		window.SwapBuffers()
		glfw.PollEvents()

		time.Sleep(refreshRate * time.Millisecond)
	}
}
func mouseCallback(w *glfw.Window, xpos, ypos float64) {
	if w.GetMouseButton(glfw.MouseButtonRight) == glfw.Press {
		if firstMouse {
			lastX = xpos
			lastY = ypos
			firstMouse = false
			return
		}

		xoffset := xpos - lastX
		yoffset := lastY - ypos // Reversed since y-coordinates go from bottom to top
		lastX = xpos
		lastY = ypos

		camera.ProcessMouseMovement(float32(xoffset), float32(yoffset), true)
	} else {
		firstMouse = true
	}
}
