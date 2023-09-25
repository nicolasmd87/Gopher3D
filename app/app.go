package main

import (
	"log"
	"runtime"
	"time"

	"Gopher3D/renderer"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

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

	var lastTime = glfw.GetTime()
	for !window.ShouldClose() {
		currentTime := glfw.GetTime()
		deltaTime := currentTime - lastTime
		lastTime = currentTime

		renderer.Render(deltaTime)
		window.SwapBuffers()
		glfw.PollEvents()

		time.Sleep(16 * time.Millisecond)
	}
}
