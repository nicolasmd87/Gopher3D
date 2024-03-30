package renderer

import (
	"Gopher3D/internal/logger"
	"image"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
	as "github.com/vulkan-go/asche"
	vk "github.com/vulkan-go/vulkan"
	"go.uber.org/zap"
)

type VulkanRendererAsche struct {
	VulkanApp             *Application
	platform              as.Platform
	FrustumCullingEnabled bool
	FaceCullingEnabled    bool
	Debug                 bool
	Shader                Shader
	Models                []*Model
}
type Application struct {
	*BaseAPP
	debugEnabled bool
	window       *glfw.Window
}

type BaseAPP struct {
	as.BaseVulkanApp
}

func NewVulkanApp() *Application {
	return &Application{debugEnabled: true, BaseAPP: &BaseAPP{}}
}

func (app *Application) VulkanInit(newCtx as.Context) error {
	logger.Log.Info("Initializing Vulkan Context")
	if newCtx == nil {
		logger.Log.Error("VulkanInit: Context is nil")
		return vk.Error(vk.ErrorInitializationFailed)
	}
	logger.Log.Info("VulkanInit: Context initialized")
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
	return false
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
		Width: uint32(1024), Height: uint32(768), Format: vk.FormatB8g8r8a8Unorm,
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
	runtime.LockOSThread()
	vk.SetGetInstanceProcAddr(glfw.GetVulkanGetInstanceProcAddress())
	rend.VulkanApp = NewVulkanApp()
	rend.VulkanApp.window = window

	logger.Log.Info("Initializing Vulkan Renderer")

	// Initialize Vulkan
	if err := vk.Init(); err != nil {
		logger.Log.Error("Failed to initialize Vulkan", zap.Error(err))
		return
	}

	// Create a new Asche platform
	logger.Log.Info("Vulkan APP", zap.Any("app", rend.VulkanApp))

	logger.Log.Info("Creating Asche platform")

	platform, err := as.NewPlatform(rend.VulkanApp)

	if err != nil {
		logger.Log.Error("Failed to create Asche platform", zap.Error(err))
		return
	}

	/*dim := rend.VulkanApp.Context().SwapchainDimensions()
	logger.Log.Info("Initialized %s with %+v swapchain", zap.String("VulkanApp:", rend.VulkanApp.VulkanAppName()), zap.Any("Swapchain dimensions:", dim))*/

	rend.platform = platform

	logger.Log.Info("Vulkan Renderer initialized successfully")
}

func (rend *VulkanRendererAsche) Render(camera Camera, light *Light) {
	/*	if rend.VulkanApp.ctx == nil {
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
		rend.VulkanApp.ctx.PresentImage(imageIdx)*/
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
