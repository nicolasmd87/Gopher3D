package engine

import (
	behaviour "Gopher3D/internal/Behaviour"
	"Gopher3D/internal/logger"
	"Gopher3D/internal/renderer"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"go.uber.org/zap"
)

var COLOR_ACTIVECAPTION int32 = 2

// Initialize to the center of the window
var lastX, lastY float64
var firstMouse bool = true
var camera renderer.Camera
var refreshRate time.Duration = 1000 / 144 // 144 FPS

// TODO: Separate window into an abtact class with width and height as fields
type Gopher struct {
	ModelChan      chan *renderer.Model
	ModelBatchChan chan []*renderer.Model
	window         *glfw.Window
	Light          *renderer.Light
	Width          int32
	Height         int32
}

func NewGopher() *Gopher {
	logger.Init()
	logger.Log.Info("Gopher3D initializing...")
	return &Gopher{
		Width:          1024,
		Height:         768,
		ModelChan:      make(chan *renderer.Model, 1000000),
		ModelBatchChan: make(chan []*renderer.Model, 1000000),
	}
}

func (gopher *Gopher) Render(x, y int) {
	lastX, lastY = float64(gopher.Width/2), float64(gopher.Width/2)
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		logger.Log.Error("Could not initialize glfw: %v", zap.Error(err))
	}
	defer glfw.Terminate()

	// Set GLFW window hints here
	glfw.WindowHint(glfw.Decorated, glfw.True)
	glfw.WindowHint(glfw.Resizable, glfw.True)
	window, err := glfw.CreateWindow(int(gopher.Width), int(gopher.Height), "Gopher3D", nil, nil)

	if err != nil {
		logger.Log.Error("Could not create glfw window: %v", zap.Error(err))
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		logger.Log.Error("Could not initialize OpenGL: %v", zap.Error(err))
	}

	// Set GLFW window position here using the passed-in position
	window.SetPos(x, y)

	renderer.Init(gopher.Width, gopher.Height)

	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	camera = renderer.NewCamera(gopher.Width, gopher.Height) // Initialize the global camera variable

	//window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled) // Hide and capture the cursor
	window.SetInputMode(glfw.CursorMode, glfw.CursorNormal) // Set cursor to normal mode initially

	window.SetCursorPosCallback(mouseCallback) // Set the callback function for mouse movement

	// Debug mode
	renderer.Debug = false

	var lastTime = glfw.GetTime()

	// TODO: Frame limiter - timer.sleep(remaining time to complete frame)
	for !window.ShouldClose() {
		currentTime := glfw.GetTime()
		deltaTime := currentTime - lastTime
		lastTime = currentTime
		camera.ProcessKeyboard(window, float32(deltaTime))
		behaviour.GlobalBehaviourManager.UpdateAll() // Update all behaviors
		renderer.Render(camera, deltaTime, gopher.Light)

		window.SwapBuffers()
		glfw.PollEvents()

		select {
		case model := <-gopher.ModelChan:
			renderer.AddModel(model)
			//renderer.SetTexture("../examples/textures/2k_mars.jpg", model)
		case modelBatch := <-gopher.ModelBatchChan:
			AddModelBatch(modelBatch)
			continue
		case <-time.After(refreshRate):
			continue
		}

	}
}

func mouseCallback(w *glfw.Window, xpos, ypos float64) {
	// Check if the window is focused and the right mouse button is pressed
	if w.GetAttrib(glfw.Focused) == glfw.True && w.GetMouseButton(glfw.MouseButtonRight) == glfw.Press {
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

// TODO: Fix ?? Probably an issue with pointers
func AddModelBatch(models []*renderer.Model) {
	for _, model := range models {
		if model != nil {
			renderer.AddModel(model)
		}
	}
}
