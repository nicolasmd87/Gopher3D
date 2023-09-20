package main

import (
	"fmt"
	"runtime"
	"time"

	"Gopher3D/renderer"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func main() {
	runtime.LockOSThread()

	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		panic(fmt.Errorf("could not initialize glfw: %v", err))
	}
	defer glfw.Terminate()

	// Create a new window
	window, err := glfw.CreateWindow(800, 600, "Rotating Triangle", nil, nil)
	if err != nil {
		panic(fmt.Errorf("could not create glfw window: %v", err))
	}
	window.MakeContextCurrent()

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		panic(fmt.Errorf("could not initialize OpenGL: %v", err))
	}

	// Initialize the renderer
	renderer.Init()

	// Set the clear color to black
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	// Main loop
	var lastTime = glfw.GetTime()

	for !window.ShouldClose() {
		// Calculate deltaTime
		currentTime := glfw.GetTime()
		deltaTime := currentTime - lastTime
		lastTime = currentTime

		// Call the renderer's Render function
		renderer.Render(deltaTime)

		// Swap the buffers
		window.SwapBuffers()

		// Poll for events
		glfw.PollEvents()

		// Sleep to avoid maxing out CPU
		time.Sleep(16 * time.Millisecond)
	}
}
