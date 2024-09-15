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

// TODO: Separate window into an abstract class with width and height as fields
type Gopher struct {
	Width          int32
	Height         int32
	ModelChan      chan *renderer.Model
	ModelBatchChan chan []*renderer.Model
	Light          *renderer.Light
	rendererAPI    renderer.Render
	window         *glfw.Window
	Camera         *renderer.Camera
	frameTrackId   int
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
		//TODO: We need to be able to set width and height of the window
		rendererAPI:    rendAPI,
		Width:          1024,
		Height:         768,
		ModelChan:      make(chan *renderer.Model, 1000000),
		ModelBatchChan: make(chan []*renderer.Model, 1000000),
		frameTrackId:   0,
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

	gopher.window, err = glfw.CreateWindow(int(gopher.Width), int(gopher.Height), "Gopher3D", nil, nil)

	if err != nil {
		logger.Log.Error("Could not create glfw window: %v", zap.Error(err))
	}

	if _, ok := gopher.rendererAPI.(*renderer.OpenGLRenderer); ok {
		gopher.window.MakeContextCurrent()
		if err := gl.Init(); err != nil {
			logger.Log.Error("Could not initialize OpenGL: %v", zap.Error(err))
			return
		}
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	}

	gopher.window.SetPos(x, y)

	gopher.rendererAPI.Init(gopher.Width, gopher.Height, gopher.window)

	// Fixed camera in each scene for now
	gopher.Camera = renderer.NewDefaultCamera(gopher.Width, gopher.Height)

	//window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled) // Hide and capture the cursor
	gopher.window.SetInputMode(glfw.CursorMode, glfw.CursorNormal) // Set cursor to normal mode initially

	gopher.window.SetCursorPosCallback(gopher.mouseCallback) // Set the callback function for mouse movement

	gopher.RenderLoop()
}

func (gopher *Gopher) RenderLoop() {
	var lastTime = glfw.GetTime()
	for !gopher.window.ShouldClose() {
		currentTime := glfw.GetTime()
		deltaTime := currentTime - lastTime
		lastTime = currentTime

		gopher.Camera.ProcessKeyboard(gopher.window, float32(deltaTime))

		//TODO: Rignt now it's fixed but maybe in the future we can make it confgigurable?
		if gopher.frameTrackId >= 2 {
			behaviour.GlobalBehaviourManager.UpdateAllFixed()
			gopher.frameTrackId = 0
		}
		behaviour.GlobalBehaviourManager.UpdateAll()
		gopher.rendererAPI.Render(*gopher.Camera, gopher.Light)

		switch gopher.rendererAPI.(type) {
		case *renderer.OpenGLRenderer:
			gopher.window.SwapBuffers()
		}
		gopher.frameTrackId++
		glfw.PollEvents()
	}
	gopher.rendererAPI.Cleanup()
}

func (gopher *Gopher) SetDebugMode(debug bool) {
	renderer.Debug = debug
}

func (gopher *Gopher) SetFrustumCulling(enabled bool) {
	renderer.FrustumCullingEnabled = enabled
}

func (gopher *Gopher) SetFaceCulling(enabled bool) {
	renderer.FaceCullingEnabled = enabled
}

func (gopher *Gopher) AddModel(model *renderer.Model) {
	gopher.rendererAPI.AddModel(model)
}

func (gopher *Gopher) RemoveModel(model *renderer.Model) {
	gopher.rendererAPI.RemoveModel(model)
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
func (gopher *Gopher) mouseCallback(w *glfw.Window, xpos, ypos float64) {
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

		gopher.Camera.ProcessMouseMovement(float32(xoffset), float32(yoffset), true)
	} else {
		firstMouse = true
	}
}
