package main

import (
	"log"
	"runtime"
	"time"

	"Gopher3D/renderer"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var lastX, lastY float64 = 400, 300 // Initialize to the center of the window
var firstMouse = true
var camera renderer.Camera

func main() {
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		log.Fatalf("could not initialize glfw: %v", err)
	}
	defer glfw.Terminate()

	window, err := glfw.CreateWindow(800, 600, "3D Engine", nil, nil)
	if err != nil {
		log.Fatalf("could not create glfw window: %v", err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		log.Fatalf("could not initialize OpenGL: %v", err)
	}

	renderer.Init()
	model := renderer.LoadObject()
	renderer.AddModel(model)

	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	camera = renderer.NewCamera(800, 600) // Initialize the global camera variable

	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled) // Hide and capture the cursor
	window.SetCursorPosCallback(mouseCallback)                // Set the callback function for mouse movement

	var lastTime = glfw.GetTime()
	for !window.ShouldClose() {
		currentTime := glfw.GetTime()
		deltaTime := currentTime - lastTime
		lastTime = currentTime

		camera.ProcessKeyboard(window, float32(deltaTime))
		renderer.Render(camera, deltaTime) // Pass the dereferenced camera object to Render

		window.SwapBuffers()
		glfw.PollEvents()

		time.Sleep(16 * time.Millisecond)
	}
}

func mouseCallback(w *glfw.Window, xpos, ypos float64) {
	if w.GetMouseButton(glfw.MouseButtonRight) == glfw.Press {
		xoffset := xpos - lastX
		yoffset := lastY - ypos // Reversed since y-coordinates go from bottom to top
		lastX = xpos
		lastY = ypos

		camera.ProcessMouseMovement(float32(xoffset), float32(yoffset), true)
	} else {
		// Update the last mouse position even if the right mouse button is not pressed
		// to prevent a sudden jump when the button is pressed again.
		lastX = xpos
		lastY = ypos
	}
}
