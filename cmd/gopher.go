package main

import (
	"log"
	"runtime"
	"time"

	"Gopher3D/renderer"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	imgui "github.com/inkyblackness/imgui-go"
)

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

	// Begin GUI
	// After creating the ImGui context
	imgui.CreateContext(nil)

	// Retrieve font texture data
	io := imgui.CurrentIO()
	image := io.Fonts().TextureDataRGBA32()
	io.SetDisplaySize(imgui.Vec2{X: float32(width), Y: float32(height)})

	// Upload the texture data to OpenGL
	var fontTexture uint32
	gl.GenTextures(1, &fontTexture)
	gl.BindTexture(gl.TEXTURE_2D, fontTexture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(image.Width), int32(image.Height), 0, gl.RGBA, gl.UNSIGNED_BYTE, image.Pixels)

	// Set the texture ID for ImGui
	io.Fonts().SetTextureID(imgui.TextureID(fontTexture))

	// End GUI
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		log.Fatalf("Could not initialize OpenGL: %v", err)
	}

	renderer.Init(width, height)
	model := renderer.LoadObject()
	renderer.AddModel(model)

	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	camera = renderer.NewCamera(width, height)

	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	window.SetCursorPosCallback(mouseCallback)

	var lastTime = glfw.GetTime()
	for !window.ShouldClose() {
		// Start the Dear ImGui frame
		imgui.NewFrame()

		// Create a top menu bar
		if imgui.BeginMainMenuBar() {
			if imgui.BeginMenu("Model") {
				if imgui.MenuItem("Load") {
					// Handle the loading logic here
				}
				imgui.EndMenu()
			}
			imgui.EndMainMenuBar()
		}

		currentTime := glfw.GetTime()
		deltaTime := currentTime - lastTime
		lastTime = currentTime
		camera.ProcessKeyboard(window, float32(deltaTime))
		renderer.Render(camera, deltaTime)

		// Render ImGui's output
		imgui.Render()
		// Here, you would integrate the rendering logic to draw ImGui's data using OpenGL.

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
