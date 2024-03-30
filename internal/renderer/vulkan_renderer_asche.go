package renderer

import (
	"Gopher3D/internal/logger"
	"image"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/vulkan-go/asche"
	as "github.com/vulkan-go/asche"
	vk "github.com/vulkan-go/vulkan"
	"go.uber.org/zap"
)

type VulkanRendererAsche struct {
	VulkanApp             *Application
	FrustumCullingEnabled bool
	FaceCullingEnabled    bool
	Debug                 bool
	Shader                Shader
	Models                []*Model
}
type Application struct {
	app      *as.BaseVulkanApp
	ctx      as.Context
	platform as.Platform
	window   *glfw.Window
}

func NewVulkanApp(win *glfw.Window) *Application {
	if win == nil {
		logger.Log.Error("NewVulkanApp: Window is nil")
		return nil
	}
	return &Application{window: win}
}

func (app *Application) VulkanInit(newCtx asche.Context) error {
	logger.Log.Info("Initializing Vulkan Context")
	if newCtx == nil {
		logger.Log.Error("VulkanInit: Context is nil")
		return vk.Error(vk.ErrorInitializationFailed)
	}
	logger.Log.Info("VulkanInit: Context initialized")
	app.ctx = newCtx
	return nil
}

func (app *Application) VulkanSurface(instance vk.Instance) (surface vk.Surface) {
	logger.Log.Info("Creating Vulkan Surface")
	surfUint64, err := app.window.CreateWindowSurface(instance, nil)
	if err != nil {
		logger.Log.Error("Failed to create window surface", zap.Error(err))
		return vk.NullSurface
	}
	surface = vk.SurfaceFromPointer(surfUint64)
	if err != nil {
		logger.Log.Error("Failed to create window surface", zap.Error(err))
		return vk.NullSurface
	}
	return surface
}

func (a *Application) VulkanLayers() []string {
	return []string{
		//		"VK_LAYER_GOOGLE_threading",
		// 		"VK_LAYER_LUNARG_parameter_validation",
		//		"VK_LAYER_LUNARG_object_tracker",
		//		"VK_LAYER_LUNARG_core_validation",
		//		"VK_LAYER_LUNARG_api_dump",
		//		"VK_LAYER_LUNARG_swapchain",
		///		"VK_LAYER_GOOGLE_unique_objects",
	}
}

func (app *Application) VulkanDebug() bool {
	return true
}

func (app *Application) VulkanAppName() string {
	return "Gopher3D"
}

func (app *Application) VulkanAppVersion() vk.Version {
	return vk.Version(vk.Version11)
}

func (app *Application) VulkanAPIVersion() vk.Version {
	return vk.Version(vk.ApiVersion11)
}

func (app *Application) VulkanInstanceExtensions() []string {
	return app.window.GetRequiredInstanceExtensions()
}

func (app *Application) VulkanSwapchainDimensions() *as.SwapchainDimensions {
	return &as.SwapchainDimensions{
		Width: uint32(app.window.GetMonitor().GetVideoMode().Height), Height: uint32(app.window.GetMonitor().GetVideoMode().Width), Format: vk.FormatB8g8r8a8Unorm,
	}
}

func (app *Application) VulkanMode() as.VulkanMode {
	return as.DefaultVulkanMode
}

func (app *Application) VulkanDeviceExtensions() []string {
	return []string{
		"VK_KHR_swapchain",
	}
}

func (rend *VulkanRendererAsche) Init(width, height int32, window *glfw.Window) {
	vk.SetGetInstanceProcAddr(glfw.GetVulkanGetInstanceProcAddress())
	rend.VulkanApp = NewVulkanApp(window)

	logger.Log.Info("Initializing Vulkan Renderer")

	// Initialize Vulkan
	if err := vk.Init(); err != nil {
		logger.Log.Error("Failed to initialize Vulkan", zap.Error(err))
		return
	}

	// Create a new Asche platform
	logger.Log.Info("Vulkan APP", zap.Any("app", rend.VulkanApp))
	logger.Log.Info("Vulkan APP window", zap.Any("win", rend.VulkanApp.window))

	platform, err := as.NewPlatform(rend.VulkanApp.app)

	logger.Log.Info("Creating Asche platform")
	if err != nil {
		logger.Log.Error("Failed to create Asche platform", zap.Error(err))
		return
	}

	rend.VulkanApp.platform = platform

	logger.Log.Info("Vulkan Renderer initialized successfully")
}

func (rend *VulkanRendererAsche) Render(camera Camera, light *Light) {
	if rend.VulkanApp.ctx == nil {
		return
	}
	imageIdx, outdated, err := rend.VulkanApp.ctx.AcquireNextImage()
	if err != nil {
		logger.Log.Error("Failed to acquire next image", zap.Error(err))
		return
	}
	if outdated {
		logger.Log.Info("Swapchain outdated")
		return
	}
	logger.Log.Info("Rendering frame", zap.Int("Image index", imageIdx))
	rend.VulkanApp.ctx.PresentImage(imageIdx)
}

func (rend *VulkanRendererAsche) AddModel(model *Model) {
	rend.Models = append(rend.Models, model)
}

func (rend *VulkanRendererAsche) LoadTexture(path string) (uint32, error) {
	return 0, nil
}

func (rend *VulkanRendererAsche) CreateTextureFromImage(img image.Image) (uint32, error) {
	return 0, nil
}

func (rend *VulkanRendererAsche) Cleanup() {

}

func (rend *VulkanRendererAsche) SetDebugMode(debug bool) {
	rend.Debug = debug
}

func (rend *VulkanRendererAsche) SetFrustumCulling(enabled bool) {
	rend.FrustumCullingEnabled = enabled
}

func (rend *VulkanRendererAsche) SetFaceCulling(enabled bool) {
	rend.FaceCullingEnabled = enabled
}
