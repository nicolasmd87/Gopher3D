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

// Enum for rendererAPIs Vulkan and OpenGL
type rendAPI int

const (
	OPENGL rendAPI = iota
	VULKAN
)

// TODO: Separate window into an abtact class with width and height as fields
type Gopher struct {
	Width          int32
	Height         int32
	ModelChan      chan *renderer.Model
	ModelBatchChan chan []*renderer.Model
	Light          *renderer.Light
	rendererAPI    renderer.Render
	window         *glfw.Window
}

func NewGopher(rendererAPI rendAPI) *Gopher {
	logger.Init()
	logger.Log.Info("Gopher3D initializing...")
	//Default renderer is OpenGL until we get Vulkan working
	var rendAPI renderer.Render
	if rendererAPI == OPENGL {
		rendAPI = &renderer.OpenGLRenderer{}
	} else {
		rendAPI = &renderer.VulkanRenderer{}
	}
	return &Gopher{
		//TODO: We need to be able to switch through renderers here. that's why we are building the interface
		rendererAPI:    rendAPI,
		Width:          1024,
		Height:         768,
		ModelChan:      make(chan *renderer.Model, 1000000),
		ModelBatchChan: make(chan []*renderer.Model, 1000000),
	}
}

// Gopher API
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

	var window *glfw.Window
	var err error

	switch gopher.rendererAPI.(type) {
	case *renderer.VulkanRenderer:
		// Set GLFW to not create an OpenGL context
		glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	case *renderer.OpenGLRenderer:
		glfw.WindowHint(glfw.ContextVersionMajor, 4)
		glfw.WindowHint(glfw.ContextVersionMinor, 1)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	default:
		logger.Log.Error("Unknown renderer type", zap.String("fun", "Render"))
		return
	}

	window, err = glfw.CreateWindow(int(gopher.Width), int(gopher.Height), "Gopher3D", nil, nil)

	if err != nil {
		logger.Log.Error("Could not create glfw window: %v", zap.Error(err))
	}

	if _, ok := gopher.rendererAPI.(*renderer.OpenGLRenderer); ok {
		window.MakeContextCurrent()
		if err := gl.Init(); err != nil {
			logger.Log.Error("Could not initialize OpenGL: %v", zap.Error(err))
			return
		}
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	}

	window.SetPos(x, y)

	gopher.rendererAPI.Init(gopher.Width, gopher.Height, window)

	camera = renderer.NewCamera(gopher.Width, gopher.Height) // Initialize the global camera variable

	//window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled) // Hide and capture the cursor
	window.SetInputMode(glfw.CursorMode, glfw.CursorNormal) // Set cursor to normal mode initially

	window.SetCursorPosCallback(mouseCallback) // Set the callback function for mouse movement

	gopher.SetDebugMode(false)

	gopher.RenderLoop(window)
}

func (gopher *Gopher) RenderLoop(window *glfw.Window) {
	camera = renderer.NewCamera(gopher.Width, gopher.Height)

	var lastTime = glfw.GetTime()
	for !window.ShouldClose() {
		currentTime := glfw.GetTime()
		deltaTime := currentTime - lastTime
		lastTime = currentTime

		camera.ProcessKeyboard(window, float32(deltaTime))
		behaviour.GlobalBehaviourManager.UpdateAll()
		gopher.rendererAPI.Render(camera, gopher.Light)

		switch gopher.rendererAPI.(type) {
		case *renderer.OpenGLRenderer:
			window.SwapBuffers()
		}
		glfw.PollEvents()

		select {
		case model := <-gopher.ModelChan:
			gopher.rendererAPI.AddModel(model)
		case modelBatch := <-gopher.ModelBatchChan:
			gopher.AddModelBatch(modelBatch)
			continue
		case <-time.After(refreshRate):
			continue
		}
	}
}

func (gopher *Gopher) SetDebugMode(debug bool) {
	switch renderer := gopher.rendererAPI.(type) {
	case *renderer.OpenGLRenderer:
		renderer.Debug = debug
	case *renderer.VulkanRenderer:
		renderer.Debug = debug
	default:
		logger.Log.Error("Unknown renderer type")
	}
}

func (gopher *Gopher) SetFrustumCulling(enabled bool) {
	switch renderer := gopher.rendererAPI.(type) {
	case *renderer.OpenGLRenderer:
		renderer.FrustumCullingEnabled = enabled
	case *renderer.VulkanRenderer:
		renderer.FrustumCullingEnabled = enabled
	default:
		logger.Log.Error("Unknown renderer type")
	}
}
func (gopher *Gopher) SetFaceCulling(enabled bool) {
	switch renderer := gopher.rendererAPI.(type) {
	case *renderer.OpenGLRenderer:
		renderer.FaceCullingEnabled = enabled
	case *renderer.VulkanRenderer:
		renderer.FaceCullingEnabled = enabled
	default:
		logger.Log.Error("Unknown renderer type")
	}
}

func (gopher *Gopher) AddModel(model *renderer.Model) {
	gopher.rendererAPI.AddModel(model)
}

// TODO: Fix ?? Probably an issue with pointers
func (gopher *Gopher) AddModelBatch(models []*renderer.Model) {
	for _, model := range models {
		if model != nil {
			gopher.rendererAPI.AddModel(model)
		}
	}
}

// Mouse callback function
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
