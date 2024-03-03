package renderer

import (
	"Gopher3D/internal/logger"
	"fmt"
	"image"
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
	"go.uber.org/zap"
)

type VulkanRenderer struct {
	FrustumCullingEnabled bool
	FaceCullingEnabled    bool
	Debug                 bool
	modelLoc              int32
	viewProjLoc           int32
	lightPosLoc           int32
	lightColorLoc         int32
	lightIntensityLoc     int32
	vertexShader          uint32
	fragmentShader        uint32
	shaderProgram         uint32
	Shader                Shader
	window                *glfw.Window
	instance              vk.Instance
	physicalDevice        vk.PhysicalDevice
	surface               vk.Surface
	device                vk.Device
	graphicsQueue         vk.Queue
	swapChain             vk.Swapchain
	Models                []*Model
}

type QueueFamilyIndices struct {
	GraphicsFamily int32
	PresentFamily  int32
	IsComplete     func() bool
}

type SwapChainSupportDetails struct {
	Capabilities vk.SurfaceCapabilities
	Formats      []vk.SurfaceFormat
	PresentModes []vk.PresentMode
}

func (rend *VulkanRenderer) Init(width, height int32, window *glfw.Window) {
	vk.SetGetInstanceProcAddr(glfw.GetVulkanGetInstanceProcAddress())
	rend.window = window
	// Initialize the Vulkan library
	if err := vk.Init(); err != nil {
		logger.Log.Error("Failed to initialize Vulkan: %v", zap.Error(err))
	}
	// Query the required extensions for GLFW
	glfwExtensions := window.GetRequiredInstanceExtensions()

	// Create Vulkan instance
	instanceCreateInfo := &vk.InstanceCreateInfo{
		SType: vk.StructureTypeInstanceCreateInfo,
		PApplicationInfo: &vk.ApplicationInfo{
			SType:              vk.StructureTypeApplicationInfo,
			PApplicationName:   "Vulkan Application\x00",
			ApplicationVersion: vk.MakeVersion(1, 0, 0),
			PEngineName:        "No Engine\x00",
			EngineVersion:      vk.MakeVersion(1, 0, 0),
			ApiVersion:         vk.ApiVersion10,
		},
		EnabledExtensionCount:   uint32(len(glfwExtensions)),
		PpEnabledExtensionNames: glfwExtensions,
	}

	var instance vk.Instance
	if vk.CreateInstance(instanceCreateInfo, nil, &instance) != vk.Success {
		logger.Log.Error("Failed to create instance")
	}
	rend.instance = instance

	if err := glfw.VulkanSupported(); !err {
		logger.Log.Error("Vulkan not supported")
	} else {
		surface, err := window.CreateWindowSurface(instance, nil)
		if err != nil {
			logger.Log.Error("Failed to create window surface", zap.Error(err))
		}
		rend.surface = vk.SurfaceFromPointer(surface)
	}

	// Select a physical device
	var deviceCount uint32
	vk.EnumeratePhysicalDevices(instance, &deviceCount, nil)
	if deviceCount == 0 {
		logger.Log.Error("Failed to find GPUs with Vulkan support")
	}

	var physicalDevices = make([]vk.PhysicalDevice, deviceCount)
	vk.EnumeratePhysicalDevices(instance, &deviceCount, physicalDevices)

	// Choose the first device for the moment being
	for _, pd := range physicalDevices {
		if isDeviceExtensionSupported(pd, "VK_KHR_swapchain") {
			rend.physicalDevice = pd
			logger.Log.Info("Selected device supports VK_KHR_swapchain")
			break
		}
	}

	if rend.physicalDevice == nil {
		logger.Log.Fatal("No suitable device found that supports VK_KHR_swapchain")
		return
	}
	var deviceCreateInfo vk.DeviceCreateInfo
	// Create logical device and queues
	queueFamilyIndex := findQueueFamilies(rend.physicalDevice, rend.surface)

	deviceQueueCreateInfo := vk.DeviceQueueCreateInfo{
		SType:            vk.StructureTypeDeviceQueueCreateInfo,
		QueueFamilyIndex: uint32(queueFamilyIndex.GraphicsFamily),
		QueueCount:       1,
		PQueuePriorities: []float32{1.0},
	}
	// Ensure the physical device can support the VK_KHR_swapchain extension
	if !isDeviceExtensionSupported(rend.physicalDevice, "VK_KHR_swapchain") {
		logger.Log.Fatal("Device does not support VK_KHR_swapchain")
		deviceCreateInfo = vk.DeviceCreateInfo{
			SType:                vk.StructureTypeDeviceCreateInfo,
			QueueCreateInfoCount: 1,
			PQueueCreateInfos:    []vk.DeviceQueueCreateInfo{deviceQueueCreateInfo},
		}
	} else {
		logger.Log.Info("Device supports VK_KHR_swapchain")
		deviceExtensions := []string{"VK_KHR_swapchain"}
		deviceCreateInfo = vk.DeviceCreateInfo{
			SType:                   vk.StructureTypeDeviceCreateInfo,
			QueueCreateInfoCount:    1,
			PQueueCreateInfos:       []vk.DeviceQueueCreateInfo{deviceQueueCreateInfo},
			EnabledExtensionCount:   uint32(len(deviceExtensions)),
			PpEnabledExtensionNames: deviceExtensions,
		}

	}

	var device vk.Device
	if vk.CreateDevice(rend.physicalDevice, &deviceCreateInfo, nil, &device) != vk.Success {
		logger.Log.Error("Failed to create logical device")
	}

	rend.device = device
	var queue vk.Queue

	if rend.device != nil {
		vk.GetDeviceQueue(rend.device, uint32(queueFamilyIndex.GraphicsFamily), 0, &rend.graphicsQueue)
	} else {
		logger.Log.Error("Attempted to get a device queue from an invalid device")
		return // Handle the error appropriately
	}
	rend.graphicsQueue = queue

	// Create swap chain
	swapChainSupport := querySwapChainSupport(rend.physicalDevice, rend.surface)

	surfaceFormat := chooseSwapSurfaceFormat(swapChainSupport.Formats)
	presentMode := chooseSwapPresentMode(swapChainSupport.PresentModes)
	extent := chooseSwapExtent(swapChainSupport.Capabilities, width, height)

	swapChainCreateInfo := vk.SwapchainCreateInfo{
		SType:            vk.StructureTypeSwapchainCreateInfo,
		Surface:          rend.surface,
		MinImageCount:    swapChainSupport.Capabilities.MinImageCount + 1,
		ImageFormat:      surfaceFormat.Format,
		ImageColorSpace:  surfaceFormat.ColorSpace,
		ImageExtent:      extent,
		ImageArrayLayers: 1,
		ImageUsage:       vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit),
		PresentMode:      presentMode, // Use the chosen present mode here
		// Additional fields as needed, like preTransform, compositeAlpha, clipped, oldSwapchain
		// queueFamilyIndexCount, pQueueFamilyIndices, preTransform, compositeAlpha, clipped, and oldSwapchain.
		ImageSharingMode: vk.SharingModeExclusive,
		// The preTransform and compositeAlpha fields depend on the capabilities of the physical device.
		PreTransform:   swapChainSupport.Capabilities.CurrentTransform,
		CompositeAlpha: vk.CompositeAlphaOpaqueBit,
		Clipped:        vk.True,          // Enable clipping to improve performance by not rendering obscured regions.
		OldSwapchain:   vk.NullSwapchain, // Set to VK_NULL_HANDLE unless you are recreating an existing swapchain.
	}

	var swapChain vk.Swapchain
	if vk.CreateSwapchain(device, &swapChainCreateInfo, nil, &swapChain) != vk.Success {
		logger.Log.Error("Failed to create swap chain")
	}

	rend.swapChain = swapChain // Ensure your struct has a swapChain field

	// Setup remaining components like image views, render pass, graphics pipeline, etc.I.

	logger.Log.Info("Vulkan Renderer initialized")
}

func (rend *VulkanRenderer) Render(camera Camera, light *Light) {
	// Acquire an image from the swap chain
	//rend.acquireNextImage()

	// Execute the command buffer
	//rend.executeCommandBuffer()

	// Present the image
	//rend.presentImage()
}

func (rend *VulkanRenderer) AddModel(model *Model) {
	rend.Models = append(rend.Models, model)
}

func (rend *VulkanRenderer) LoadTexture(path string) (uint32, error) {
	return 0, nil
}

func (rend *VulkanRenderer) CreateTextureFromImage(img image.Image) (uint32, error) {
	return 0, nil
}

func (rend *VulkanRenderer) Cleanup() {
	// Cleanup synchronization objects
	//cleanupSyncObjects(rend.device, rend.syncObjects)

	// Cleanup command buffers
	//cleanupCommandBuffers(rend.device, rend.commandPool, rend.commandBuffers)

	// Cleanup command pool
	//cleanupCommandPool(rend.device, rend.commandPool)

	// Cleanup framebuffers
	//cleanupFramebuffers(rend.device, rend.framebuffers)

	// Cleanup graphics pipeline
	//cleanupGraphicsPipeline(rend.device, rend.pipeline)

	// Cleanup render pass
	//cleanupRenderPass(rend.device, rend.renderPass)

	// Cleanup image views
	//cleanupImageViews(rend.device, rend.imageViews)

	// Cleanup swap chain
	//cleanupSwapChain(rend.device, rend.swapChain)

	// Cleanup logical device
	//cleanupLogicalDevice(rend.device)

	// Cleanup surface
	//cleanupSurface(rend.instance, rend.surface)

	// Cleanup debug messenger
	//cleanupDebugMessenger(rend.instance, rend.debugMessenger)

	// Cleanup Vulkan instance
	//cleanupVulkanInstance(rend.instance)
}

func (rend *VulkanRenderer) SetDebugMode(debug bool) {
	rend.Debug = debug
}

func (rend *VulkanRenderer) SetFrustumCulling(enabled bool) {
	rend.FrustumCullingEnabled = enabled
}

func (rend *VulkanRenderer) SetFaceCulling(enabled bool) {
	rend.FaceCullingEnabled = enabled
}

func findQueueFamilies(device vk.PhysicalDevice, surface vk.Surface) QueueFamilyIndices {
	var indices QueueFamilyIndices
	indices.GraphicsFamily = -1
	indices.PresentFamily = -1

	indices.IsComplete = func() bool {
		return indices.GraphicsFamily >= 0 && indices.PresentFamily >= 0
	}

	var queueFamilyCount uint32
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &queueFamilyCount, nil)

	queueFamilies := make([]vk.QueueFamilyProperties, queueFamilyCount)
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &queueFamilyCount, queueFamilies)

	for i, queueFamily := range queueFamilies {
		if queueFamily.QueueFlags&vk.QueueFlags(vk.QueueGraphicsBit) != 0 {
			indices.GraphicsFamily = int32(i)
		}

		var presentSupport vk.Bool32
		vk.GetPhysicalDeviceSurfaceSupport(device, uint32(i), surface, &presentSupport)

		if presentSupport == vk.True {
			indices.PresentFamily = int32(i)
		}

		if indices.IsComplete() {
			break
		}
	}

	return indices
}

func querySwapChainSupport(device vk.PhysicalDevice, surface vk.Surface) SwapChainSupportDetails {
	var details SwapChainSupportDetails

	vk.GetPhysicalDeviceSurfaceCapabilities(device, surface, &details.Capabilities)

	var formatCount uint32
	vk.GetPhysicalDeviceSurfaceFormats(device, surface, &formatCount, nil)
	if formatCount != 0 {
		details.Formats = make([]vk.SurfaceFormat, formatCount)
		vk.GetPhysicalDeviceSurfaceFormats(device, surface, &formatCount, details.Formats)
	}

	var presentModeCount uint32
	vk.GetPhysicalDeviceSurfacePresentModes(device, surface, &presentModeCount, nil)
	if presentModeCount != 0 {
		details.PresentModes = make([]vk.PresentMode, presentModeCount)
		vk.GetPhysicalDeviceSurfacePresentModes(device, surface, &presentModeCount, details.PresentModes)
	}

	return details
}

func chooseSwapPresentMode(availablePresentModes []vk.PresentMode) vk.PresentMode {
	for _, availablePresentMode := range availablePresentModes {
		if availablePresentMode == vk.PresentModeMailbox {
			return availablePresentMode // Triple buffering
		}
	}
	return vk.PresentModeFifo // VSync
}

func chooseSwapSurfaceFormat(availableFormats []vk.SurfaceFormat) vk.SurfaceFormat {
	for _, availableFormat := range availableFormats {
		if availableFormat.Format == vk.FormatB8g8r8a8Unorm && availableFormat.ColorSpace == vk.ColorSpaceSrgbNonlinear {
			return availableFormat
		}
	}
	return availableFormats[0]
}

func chooseSwapExtent(capabilities vk.SurfaceCapabilities, width, height int32) vk.Extent2D {
	if capabilities.CurrentExtent.Width != uint32(math.MaxUint32) {
		return capabilities.CurrentExtent
	} else {
		actualExtent := vk.Extent2D{Width: uint32(width), Height: uint32(height)}
		actualExtent.Width = max(capabilities.MinImageExtent.Width, min(capabilities.MaxImageExtent.Width, actualExtent.Width))
		actualExtent.Height = max(capabilities.MinImageExtent.Height, min(capabilities.MaxImageExtent.Height, actualExtent.Height))
		return actualExtent
	}
}

func isDeviceExtensionSupported(device vk.PhysicalDevice, extensionName string) bool {

	var extensionCount uint32
	vk.EnumerateDeviceExtensionProperties(device, "", &extensionCount, nil)
	if extensionCount == 0 {
		logger.Log.Info("No extensions found for device")
		return false
	}

	extensions := make([]vk.ExtensionProperties, extensionCount)
	result := vk.EnumerateDeviceExtensionProperties(device, "", &extensionCount, extensions)
	if result != vk.Success {
		// Handle the error (e.g., log it and exit or return an error from your function)
		logger.Log.Error("Failed to enumerate device extension properties")
		return false
	}
	for _, ext := range extensions {
		fmt.Printf("Raw extension name data: %v\n", ext.ExtensionName[:])
		extName := vk.ToString(ext.ExtensionName[:]) // Manually convert C string to Go string
		logger.Log.Info("Device extension found", zap.String("extensionName", extName))
		if extName == extensionName {
			return true
		}
	}

	logger.Log.Info("Extension not supported", zap.String("searchedExtension", extensionName))
	return false
}

func min(x, y uint32) uint32 {
	if x < y {
		return x
	}
	return y
}

func max(x, y uint32) uint32 {
	if x > y {
		return x
	}
	return y
}

// Move to utils
