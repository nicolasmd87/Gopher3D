package renderer

import (
	"Gopher3D/internal/logger"
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"log"
	"unsafe"

	as "github.com/vulkan-go/asche"
	vk "github.com/vulkan-go/vulkan"
	lin "github.com/xlab/linmath"
	"go.uber.org/zap"
)

func NewScene(spinAngle float32) *Scene {
	scene := &Scene{
		spinAngle: spinAngle,
	}
	scene.modelMatrix.Identity()
	return scene
}

type Scene struct {
	as.BaseVulkanApp
	Debug bool

	width      uint32
	height     uint32
	format     vk.Format
	colorSpace vk.ColorSpace

	textures          []*Texture
	depth             *Depth
	useStagingBuffers bool

	descPool vk.DescriptorPool

	pipelineLayout vk.PipelineLayout
	descLayout     vk.DescriptorSetLayout
	pipelineCache  vk.PipelineCache
	renderPass     vk.RenderPass
	pipeline       vk.Pipeline

	projectionMatrix lin.Mat4x4
	viewMatrix       lin.Mat4x4
	modelMatrix      lin.Mat4x4

	spinAngle float32
}

func (s *Scene) prepareDepth() {
	dev := s.Context().Device()
	depthFormat := vk.FormatD16Unorm
	s.depth = &Depth{
		format: depthFormat,
	}
	ret := vk.CreateImage(dev, &vk.ImageCreateInfo{
		SType:     vk.StructureTypeImageCreateInfo,
		ImageType: vk.ImageType2d,
		Format:    depthFormat,
		Extent: vk.Extent3D{
			Width:  s.width,
			Height: s.height,
			Depth:  1,
		},
		MipLevels:   1,
		ArrayLayers: 1,
		Samples:     vk.SampleCount1Bit,
		Tiling:      vk.ImageTilingOptimal,
		Usage:       vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit),
	}, nil, &s.depth.image)

	if ret != vk.Success {
		logger.Log.Error("Failed to create depth image")
		return
	}

	logger.Log.Info("Creating depth image")

	var memReqs vk.MemoryRequirements
	vk.GetImageMemoryRequirements(dev, s.depth.image, &memReqs)
	memReqs.Deref()

	memProps := s.Context().Platform().MemoryProperties()
	memTypeIndex, _ := as.FindRequiredMemoryTypeFallback(memProps,
		vk.MemoryPropertyFlagBits(memReqs.MemoryTypeBits), vk.MemoryPropertyDeviceLocalBit)
	s.depth.memAlloc = &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: memTypeIndex,
	}

	var mem vk.DeviceMemory
	ret = vk.AllocateMemory(dev, s.depth.memAlloc, nil, &mem)

	if ret != vk.Success {
		logger.Log.Error("Failed to allocate memory for depth image")
	}

	s.depth.mem = mem

	ret = vk.BindImageMemory(dev, s.depth.image, s.depth.mem, 0)

	if ret != vk.Success {
		logger.Log.Error("Failed to bind image memory")
	}

	var view vk.ImageView
	ret = vk.CreateImageView(dev, &vk.ImageViewCreateInfo{
		SType:  vk.StructureTypeImageViewCreateInfo,
		Format: depthFormat,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectDepthBit),
			LevelCount: 1,
			LayerCount: 1,
		},
		ViewType: vk.ImageViewType2d,
		Image:    s.depth.image,
	}, nil, &view)

	if ret != vk.Success {
		logger.Log.Error("Failed to create image view")
	}

	logger.Log.Info("Creating image view")
	s.depth.view = view
}

var texEnabled = []string{
	"textures/gopher.png",
}

func (s *Scene) prepareTextureImage(path string, tiling vk.ImageTiling,
	usage vk.ImageUsageFlagBits, memoryProps vk.MemoryPropertyFlagBits) *Texture {

	dev := s.Context().Device()
	texFormat := vk.FormatR8g8b8a8Unorm
	_, width, height, err := loadTextureData(path, 0)
	if err != nil {
		logger.Log.Error("Failed to load texture data", zap.Error(err))
	}
	tex := &Texture{
		texWidth:    int32(width),
		texHeight:   int32(height),
		imageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
	}

	var image vk.Image
	ret := vk.CreateImage(dev, &vk.ImageCreateInfo{
		SType:     vk.StructureTypeImageCreateInfo,
		ImageType: vk.ImageType2d,
		Format:    texFormat,
		Extent: vk.Extent3D{
			Width:  uint32(width),
			Height: uint32(height),
			Depth:  1,
		},
		MipLevels:     1,
		ArrayLayers:   1,
		Samples:       vk.SampleCount1Bit,
		Tiling:        tiling,
		Usage:         vk.ImageUsageFlags(usage),
		InitialLayout: vk.ImageLayoutPreinitialized,
	}, nil, &image)

	if ret != vk.Success {
		logger.Log.Error("Failed to create image")
	}

	tex.image = image

	var memReqs vk.MemoryRequirements
	vk.GetImageMemoryRequirements(dev, tex.image, &memReqs)
	memReqs.Deref()

	memProps := s.Context().Platform().MemoryProperties()
	memTypeIndex, _ := as.FindRequiredMemoryTypeFallback(memProps,
		vk.MemoryPropertyFlagBits(memReqs.MemoryTypeBits), memoryProps)
	tex.memAlloc = &vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReqs.Size,
		MemoryTypeIndex: memTypeIndex,
	}
	var mem vk.DeviceMemory
	ret = vk.AllocateMemory(dev, tex.memAlloc, nil, &mem)
	if ret != vk.Success {
		logger.Log.Error("Failed to allocate memory for texture")
	}

	logger.Log.Info("Allocating memory for texture")
	tex.mem = mem
	ret = vk.BindImageMemory(dev, tex.image, tex.mem, 0)
	if ret != vk.Success {
		logger.Log.Error("Failed to bind image memory")
	}
	logger.Log.Info("Binding image memory for texture")
	hostVisible := memoryProps&vk.MemoryPropertyHostVisibleBit != 0
	if hostVisible {
		var layout vk.SubresourceLayout
		vk.GetImageSubresourceLayout(dev, tex.image, &vk.ImageSubresource{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
		}, &layout)
		layout.Deref()

		data, _, _, err := loadTextureData(path, int(layout.RowPitch))
		if err != nil {
			logger.Log.Error("Failed to load texture data", zap.Error(err))
		}
		if len(data) > 0 {
			var pData unsafe.Pointer
			ret = vk.MapMemory(dev, tex.mem, 0, vk.DeviceSize(len(data)), 0, &pData)
			if ret != vk.Success {
				log.Printf("vulkan warning: failed to map device memory for data (len=%d)", len(data))
				return tex
			}
			n := vk.Memcopy(pData, data)
			if n != len(data) {
				log.Printf("vulkan warning: failed to copy data, %d != %d", n, len(data))
			}
			vk.UnmapMemory(dev, tex.mem)
		}
	}
	return tex
}

func (s *Scene) setImageLayout(image vk.Image, aspectMask vk.ImageAspectFlagBits,
	oldImageLayout, newImageLayout vk.ImageLayout,
	srcAccessMask vk.AccessFlagBits,
	srcStages, dstStages vk.PipelineStageFlagBits) {

	cmd := s.Context().CommandBuffer()
	if cmd == nil {
		logger.Log.Error("vulkan: command buffer not initialized")
	}

	imageMemoryBarrier := vk.ImageMemoryBarrier{
		SType:         vk.StructureTypeImageMemoryBarrier,
		SrcAccessMask: vk.AccessFlags(srcAccessMask),
		DstAccessMask: 0,
		OldLayout:     oldImageLayout,
		NewLayout:     newImageLayout,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask: vk.ImageAspectFlags(aspectMask),
			LayerCount: 1,
			LevelCount: 1,
		},
		Image: image,
	}
	switch newImageLayout {
	case vk.ImageLayoutTransferDstOptimal:
		// make sure anything that was copying from this image has completed
		imageMemoryBarrier.DstAccessMask = vk.AccessFlags(vk.AccessTransferWriteBit)
	case vk.ImageLayoutColorAttachmentOptimal:
		imageMemoryBarrier.DstAccessMask = vk.AccessFlags(vk.AccessColorAttachmentWriteBit)
	case vk.ImageLayoutDepthStencilAttachmentOptimal:
		imageMemoryBarrier.DstAccessMask = vk.AccessFlags(vk.AccessDepthStencilAttachmentWriteBit)
	case vk.ImageLayoutShaderReadOnlyOptimal:
		imageMemoryBarrier.DstAccessMask =
			vk.AccessFlags(vk.AccessShaderReadBit) | vk.AccessFlags(vk.AccessInputAttachmentReadBit)
	case vk.ImageLayoutTransferSrcOptimal:
		imageMemoryBarrier.DstAccessMask = vk.AccessFlags(vk.AccessTransferReadBit)
	case vk.ImageLayoutPresentSrc:
		imageMemoryBarrier.DstAccessMask = vk.AccessFlags(vk.AccessMemoryReadBit)
	default:
		imageMemoryBarrier.DstAccessMask = 0
	}

	vk.CmdPipelineBarrier(cmd,
		vk.PipelineStageFlags(srcStages), vk.PipelineStageFlags(dstStages),
		0, 0, nil, 0, nil, 1, []vk.ImageMemoryBarrier{imageMemoryBarrier})
}

func (s *Scene) prepareTextures() {
	dev := s.Context().Device()
	texFormat := vk.FormatR8g8b8a8Unorm
	var props vk.FormatProperties
	gpu := s.Context().Platform().PhysicalDevice()
	vk.GetPhysicalDeviceFormatProperties(gpu, texFormat, &props)
	props.Deref()

	prepareTex := func(path string) *Texture {
		var tex *Texture

		if (props.LinearTilingFeatures&vk.FormatFeatureFlags(vk.FormatFeatureSampledImageBit) != 0) &&
			!s.useStagingBuffers {
			// -> device can texture using linear textures

			tex = s.prepareTextureImage(path, vk.ImageTilingLinear, vk.ImageUsageSampledBit,
				vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit)

			// Nothing in the pipeline needs to be complete to start, and don't allow fragment
			// shader to run until layout transition completes
			s.setImageLayout(tex.image, vk.ImageAspectColorBit,
				vk.ImageLayoutPreinitialized, tex.imageLayout,
				vk.AccessHostWriteBit,
				vk.PipelineStageTopOfPipeBit, vk.PipelineStageFragmentShaderBit)

		} else if props.OptimalTilingFeatures&vk.FormatFeatureFlags(vk.FormatFeatureSampledImageBit) != 0 {
			//  Must use staging buffer to copy linear texture to optimized
			log.Println("vulkan warn: using staging buffers")

			staging := s.prepareTextureImage(path, vk.ImageTilingLinear, vk.ImageUsageTransferSrcBit,
				vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit)
			tex = s.prepareTextureImage(path, vk.ImageTilingOptimal,
				vk.ImageUsageTransferDstBit|vk.ImageUsageSampledBit, vk.MemoryPropertyDeviceLocalBit)

			s.setImageLayout(staging.image, vk.ImageAspectColorBit,
				vk.ImageLayoutPreinitialized, vk.ImageLayoutTransferSrcOptimal,
				vk.AccessHostWriteBit,
				vk.PipelineStageTopOfPipeBit, vk.PipelineStageTransferBit)

			s.setImageLayout(tex.image, vk.ImageAspectColorBit,
				vk.ImageLayoutPreinitialized, vk.ImageLayoutTransferDstOptimal,
				vk.AccessHostWriteBit,
				vk.PipelineStageTopOfPipeBit, vk.PipelineStageTransferBit)

			cmd := s.Context().CommandBuffer()
			if cmd == nil {
				logger.Log.Error("vulkan: command buffer not initialized")
			}
			vk.CmdCopyImage(cmd, staging.image, vk.ImageLayoutTransferSrcOptimal,
				tex.image, vk.ImageLayoutTransferDstOptimal,
				1, []vk.ImageCopy{{
					SrcSubresource: vk.ImageSubresourceLayers{
						AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
						LayerCount: 1,
					},
					SrcOffset: vk.Offset3D{
						X: 0, Y: 0, Z: 0,
					},
					DstSubresource: vk.ImageSubresourceLayers{
						AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
						LayerCount: 1,
					},
					DstOffset: vk.Offset3D{
						X: 0, Y: 0, Z: 0,
					},
					Extent: vk.Extent3D{
						Width:  uint32(staging.texWidth),
						Height: uint32(staging.texHeight),
						Depth:  1,
					},
				}})
			s.setImageLayout(tex.image, vk.ImageAspectColorBit,
				vk.ImageLayoutTransferDstOptimal, tex.imageLayout,
				vk.AccessTransferWriteBit,
				vk.PipelineStageTransferBit, vk.PipelineStageFragmentShaderBit)
			// cannot destroy until cmd is submitted.. must keep a list somewhere
			// staging.DestroyImage(dev)
		} else {
			logger.Log.Error("No support for R8g8b8a8Unorm as texture image format")
		}

		var sampler vk.Sampler
		ret := vk.CreateSampler(dev, &vk.SamplerCreateInfo{
			SType:                   vk.StructureTypeSamplerCreateInfo,
			MagFilter:               vk.FilterNearest,
			MinFilter:               vk.FilterNearest,
			MipmapMode:              vk.SamplerMipmapModeNearest,
			AddressModeU:            vk.SamplerAddressModeClampToEdge,
			AddressModeV:            vk.SamplerAddressModeClampToEdge,
			AddressModeW:            vk.SamplerAddressModeClampToEdge,
			AnisotropyEnable:        vk.False,
			MaxAnisotropy:           1,
			CompareOp:               vk.CompareOpNever,
			BorderColor:             vk.BorderColorFloatOpaqueWhite,
			UnnormalizedCoordinates: vk.False,
		}, nil, &sampler)
		tex.sampler = sampler

		if ret != vk.Success {
			logger.Log.Error("Failed to create sampler")
		}

		logger.Log.Info("Creating sampler")
		var view vk.ImageView
		ret = vk.CreateImageView(dev, &vk.ImageViewCreateInfo{
			SType:    vk.StructureTypeImageViewCreateInfo,
			Image:    tex.image,
			ViewType: vk.ImageViewType2d,
			Format:   texFormat,
			Components: vk.ComponentMapping{
				R: vk.ComponentSwizzleR,
				G: vk.ComponentSwizzleG,
				B: vk.ComponentSwizzleB,
				A: vk.ComponentSwizzleA,
			},
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
				LevelCount: 1,
				LayerCount: 1,
			},
		}, nil, &view)

		if ret != vk.Success {
			logger.Log.Error("Failed to create image view")
		}

		tex.view = view

		return tex
	}

	s.textures = make([]*Texture, 0, len(texEnabled))
	for _, texFile := range texEnabled {
		s.textures = append(s.textures, prepareTex(texFile))
	}
}

func (s *Scene) drawBuildCommandBuffer(res *as.SwapchainImageResources, cmd vk.CommandBuffer) {
	ret := vk.BeginCommandBuffer(cmd, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageSimultaneousUseBit),
	})
	logger.Log.Info("Begin command buffer")
	clearValues := make([]vk.ClearValue, 2)
	clearValues[1].SetDepthStencil(1, 0)
	clearValues[0].SetColor([]float32{
		0.2, 0.2, 0.2, 0.2,
	})
	vk.CmdBeginRenderPass(cmd, &vk.RenderPassBeginInfo{
		SType:       vk.StructureTypeRenderPassBeginInfo,
		RenderPass:  s.renderPass,
		Framebuffer: res.Framebuffer(),
		RenderArea: vk.Rect2D{
			Offset: vk.Offset2D{
				X: 0, Y: 0,
			},
			Extent: vk.Extent2D{
				Width:  s.width,
				Height: s.height,
			},
		},
		ClearValueCount: 2,
		PClearValues:    clearValues,
	}, vk.SubpassContentsInline)

	vk.CmdBindPipeline(cmd, vk.PipelineBindPointGraphics, s.pipeline)
	vk.CmdBindDescriptorSets(cmd, vk.PipelineBindPointGraphics, s.pipelineLayout,
		0, 1, []vk.DescriptorSet{res.DescriptorSet()}, 0, nil)
	vk.CmdSetViewport(cmd, 0, 1, []vk.Viewport{{
		Width:    float32(s.width),
		Height:   float32(s.height),
		MinDepth: 0.0,
		MaxDepth: 1.0,
	}})
	vk.CmdSetScissor(cmd, 0, 1, []vk.Rect2D{{
		Offset: vk.Offset2D{
			X: 0, Y: 0,
		},
		Extent: vk.Extent2D{
			Width:  s.width,
			Height: s.height,
		},
	}})

	vk.CmdDraw(cmd, 12*3, 1, 0, 0)
	// Note that ending the renderpass changes the image's layout from
	// vk.ImageLayoutColorAttachmentOptimal to vk.ImageLayoutPresentSrc
	vk.CmdEndRenderPass(cmd)

	graphicsQueueIndex := s.Context().Platform().GraphicsQueueFamilyIndex()
	presentQueueIndex := s.Context().Platform().PresentQueueFamilyIndex()
	if graphicsQueueIndex != presentQueueIndex {
		// Separate Present Queue Case
		//
		// We have to transfer ownership from the graphics queue family to the
		// present queue family to be able to present.  Note that we don't have
		// to transfer from present queue family back to graphics queue family at
		// the start of the next frame because we don't care about the image's
		// contents at that point.
		vk.CmdPipelineBarrier(cmd,
			vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
			vk.PipelineStageFlags(vk.PipelineStageBottomOfPipeBit),
			0, 0, nil, 0, nil, 1, []vk.ImageMemoryBarrier{{
				SType:               vk.StructureTypeImageMemoryBarrier,
				SrcAccessMask:       0,
				DstAccessMask:       vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
				OldLayout:           vk.ImageLayoutPresentSrc,
				NewLayout:           vk.ImageLayoutPresentSrc,
				SrcQueueFamilyIndex: graphicsQueueIndex,
				DstQueueFamilyIndex: presentQueueIndex,
				SubresourceRange: vk.ImageSubresourceRange{
					AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
					LayerCount: 1,
					LevelCount: 1,
				},
				Image: res.Image(),
			}})
	}
	ret = vk.EndCommandBuffer(cmd)

	if ret != vk.Success {
		logger.Log.Error("Failed to end command buffer")
	}

}

func (s *Scene) prepareCubeDataBuffers() {
	dev := s.Context().Device()

	var VP lin.Mat4x4
	var MVP lin.Mat4x4
	VP.Mult(&s.projectionMatrix, &s.viewMatrix)
	MVP.Mult(&VP, &s.modelMatrix)

	data := vkTexCubeUniform{
		mvp: MVP,
	}
	for i := 0; i < 12*3; i++ {
		data.position[i][0] = gVertexBufferData[i*3]
		data.position[i][1] = gVertexBufferData[i*3+1]
		data.position[i][2] = gVertexBufferData[i*3+2]
		data.position[i][3] = 1.0
		data.attr[i][0] = gUVBufferData[2*i]
		data.attr[i][1] = gUVBufferData[2*i+1]
		data.attr[i][2] = 0
		data.attr[i][3] = 0
	}

	dataRaw := data.Data()
	memProps := s.Context().Platform().MemoryProperties()
	swapchainImageResources := s.Context().SwapchainImageResources()
	for _, res := range swapchainImageResources {
		buf := as.CreateBuffer(dev, memProps, dataRaw, vk.BufferUsageUniformBufferBit)
		res.SetUniformBuffer(buf.Buffer, buf.Memory)
	}
}

func (s *Scene) prepareDescriptorLayout() {
	dev := s.Context().Device()

	var descLayout vk.DescriptorSetLayout
	ret := vk.CreateDescriptorSetLayout(dev, &vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		BindingCount: 2,
		PBindings: []vk.DescriptorSetLayoutBinding{
			{
				Binding:         0,
				DescriptorType:  vk.DescriptorTypeUniformBuffer,
				DescriptorCount: 1,
				StageFlags:      vk.ShaderStageFlags(vk.ShaderStageVertexBit),
			}, {
				Binding:         1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				DescriptorCount: uint32(len(texEnabled)),
				StageFlags:      vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
			}},
	}, nil, &descLayout)

	if ret != vk.Success {
		logger.Log.Error("Failed to create descriptor set layout")
		return
	}
	s.descLayout = descLayout

	var pipelineLayout vk.PipelineLayout
	ret = vk.CreatePipelineLayout(dev, &vk.PipelineLayoutCreateInfo{
		SType:          vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount: 1,
		PSetLayouts: []vk.DescriptorSetLayout{
			s.descLayout,
		},
	}, nil, &pipelineLayout)

	if ret != vk.Success {
		logger.Log.Error("Failed to create pipeline layout")
		return
	}

	logger.Log.Info("Creating pipeline layout")
	s.pipelineLayout = pipelineLayout
}

func (s *Scene) prepareRenderPass() {
	dev := s.Context().Device()
	// The initial layout for the color and depth attachments will be vk.LayoutUndefined
	// because at the start of the renderpass, we don't care about their contents.
	// At the start of the subpass, the color attachment's layout will be transitioned
	// to vk.LayoutColorAttachmentOptimal and the depth stencil attachment's layout
	// will be transitioned to vk.LayoutDepthStencilAttachmentOptimal.  At the end of
	// the renderpass, the color attachment's layout will be transitioned to
	// vk.LayoutPresentSrc to be ready to present.  This is all done as part of
	// the renderpass, no barriers are necessary.
	var renderPass vk.RenderPass
	ret := vk.CreateRenderPass(dev, &vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 2,
		PAttachments: []vk.AttachmentDescription{{
			Format:         s.Context().SwapchainDimensions().Format,
			Samples:        vk.SampleCount1Bit,
			LoadOp:         vk.AttachmentLoadOpClear,
			StoreOp:        vk.AttachmentStoreOpStore,
			StencilLoadOp:  vk.AttachmentLoadOpDontCare,
			StencilStoreOp: vk.AttachmentStoreOpDontCare,
			InitialLayout:  vk.ImageLayoutUndefined,
			FinalLayout:    vk.ImageLayoutPresentSrc,
		}, {
			Format:         s.depth.format,
			Samples:        vk.SampleCount1Bit,
			LoadOp:         vk.AttachmentLoadOpClear,
			StoreOp:        vk.AttachmentStoreOpDontCare,
			StencilLoadOp:  vk.AttachmentLoadOpDontCare,
			StencilStoreOp: vk.AttachmentStoreOpDontCare,
			InitialLayout:  vk.ImageLayoutUndefined,
			FinalLayout:    vk.ImageLayoutDepthStencilAttachmentOptimal,
		}},
		SubpassCount: 1,
		PSubpasses: []vk.SubpassDescription{{
			PipelineBindPoint:    vk.PipelineBindPointGraphics,
			ColorAttachmentCount: 1,
			PColorAttachments: []vk.AttachmentReference{{
				Attachment: 0,
				Layout:     vk.ImageLayoutColorAttachmentOptimal,
			}},
			PDepthStencilAttachment: &vk.AttachmentReference{
				Attachment: 1,
				Layout:     vk.ImageLayoutDepthStencilAttachmentOptimal,
			},
		}},
	}, nil, &renderPass)
	if ret != vk.Success {
		logger.Log.Error("Failed to create render pass")
		return
	}
	logger.Log.Info("Creating render pass")
	s.renderPass = renderPass
}

func (s *Scene) preparePipeline() {
	dev := s.Context().Device()

	vs, err := as.LoadShaderModule(dev, MustAsset("shaders/cube.vert.spv"))

	if err != nil {
		logger.Log.Error("Failed to load vertex shader", zap.Error(err))
		return
	}

	fs, err := as.LoadShaderModule(dev, MustAsset("shaders/cube.frag.spv"))

	if err != nil {
		logger.Log.Error("Failed to load fragment shader", zap.Error(err))
		return
	}

	var pipelineCache vk.PipelineCache
	ret := vk.CreatePipelineCache(dev, &vk.PipelineCacheCreateInfo{
		SType: vk.StructureTypePipelineCacheCreateInfo,
	}, nil, &pipelineCache)

	logger.Log.Info("Creating pipeline cache")

	s.pipelineCache = pipelineCache
	fillMode := vk.PolygonModeFill
	logger.Log.Info("PolygonMode", zap.Any("Debug:", s.Debug))
	if s.Debug {
		fillMode = vk.PolygonModeLine
	}
	pipelineCreateInfos := []vk.GraphicsPipelineCreateInfo{{
		SType:      vk.StructureTypeGraphicsPipelineCreateInfo,
		Layout:     s.pipelineLayout,
		RenderPass: s.renderPass,

		PDynamicState: &vk.PipelineDynamicStateCreateInfo{
			SType:             vk.StructureTypePipelineDynamicStateCreateInfo,
			DynamicStateCount: 2,
			PDynamicStates: []vk.DynamicState{
				vk.DynamicStateScissor,
				vk.DynamicStateViewport,
			},
		},
		PVertexInputState: &vk.PipelineVertexInputStateCreateInfo{
			SType: vk.StructureTypePipelineVertexInputStateCreateInfo,
		},
		PInputAssemblyState: &vk.PipelineInputAssemblyStateCreateInfo{
			SType:    vk.StructureTypePipelineInputAssemblyStateCreateInfo,
			Topology: vk.PrimitiveTopologyTriangleList,
		},

		PRasterizationState: &vk.PipelineRasterizationStateCreateInfo{
			SType:       vk.StructureTypePipelineRasterizationStateCreateInfo,
			PolygonMode: fillMode,
			CullMode:    vk.CullModeFlags(vk.CullModeBackBit),
			FrontFace:   vk.FrontFaceCounterClockwise,
			LineWidth:   1.0,
		},
		PColorBlendState: &vk.PipelineColorBlendStateCreateInfo{
			SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
			AttachmentCount: 1,
			PAttachments: []vk.PipelineColorBlendAttachmentState{{
				ColorWriteMask: 0xF,
				BlendEnable:    vk.False,
			}},
		},
		PMultisampleState: &vk.PipelineMultisampleStateCreateInfo{
			SType:                vk.StructureTypePipelineMultisampleStateCreateInfo,
			RasterizationSamples: vk.SampleCount1Bit,
		},
		PViewportState: &vk.PipelineViewportStateCreateInfo{
			SType:         vk.StructureTypePipelineViewportStateCreateInfo,
			ScissorCount:  1,
			ViewportCount: 1,
		},
		PDepthStencilState: &vk.PipelineDepthStencilStateCreateInfo{
			SType:                 vk.StructureTypePipelineDepthStencilStateCreateInfo,
			DepthTestEnable:       vk.True,
			DepthWriteEnable:      vk.True,
			DepthCompareOp:        vk.CompareOpLessOrEqual,
			DepthBoundsTestEnable: vk.False,
			Back: vk.StencilOpState{
				FailOp:    vk.StencilOpKeep,
				PassOp:    vk.StencilOpKeep,
				CompareOp: vk.CompareOpAlways,
			},
			StencilTestEnable: vk.False,
			Front: vk.StencilOpState{
				FailOp:    vk.StencilOpKeep,
				PassOp:    vk.StencilOpKeep,
				CompareOp: vk.CompareOpAlways,
			},
		},
		StageCount: 2,
		PStages: []vk.PipelineShaderStageCreateInfo{{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageVertexBit,
			Module: vs,
			PName:  "main\x00",
		}, {
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageFragmentBit,
			Module: fs,
			PName:  "main\x00",
		}},
	}}
	pipeline := make([]vk.Pipeline, 1)
	ret = vk.CreateGraphicsPipelines(dev, s.pipelineCache, 1, pipelineCreateInfos, nil, pipeline)

	if ret != vk.Success {
		logger.Log.Error("Failed to create graphics pipeline")
		return
	}

	if ret != vk.Success {
		logger.Log.Error("Failed to create graphics pipeline")
		return
	}

	s.pipeline = pipeline[0]

	vk.DestroyShaderModule(dev, vs, nil)
	vk.DestroyShaderModule(dev, fs, nil)
}

func (s *Scene) prepareDescriptorPool() {
	dev := s.Context().Device()
	swapchainImageResources := s.Context().SwapchainImageResources()
	var descPool vk.DescriptorPool
	ret := vk.CreateDescriptorPool(dev, &vk.DescriptorPoolCreateInfo{
		SType:         vk.StructureTypeDescriptorPoolCreateInfo,
		MaxSets:       uint32(len(swapchainImageResources)),
		PoolSizeCount: 2,
		PPoolSizes: []vk.DescriptorPoolSize{{
			Type:            vk.DescriptorTypeUniformBuffer,
			DescriptorCount: uint32(len(swapchainImageResources)),
		}, {
			Type:            vk.DescriptorTypeCombinedImageSampler,
			DescriptorCount: uint32(len(swapchainImageResources) * len(texEnabled)),
		}},
	}, nil, &descPool)
	if ret != vk.Success {
		logger.Log.Error("Failed to create descriptor pool")
		return
	}
	logger.Log.Info("Creating descriptor pool", zap.Any("descPool", descPool))
	s.descPool = descPool
}

func (s *Scene) prepareDescriptorSet() {
	dev := s.Context().Device()
	swapchainImageResources := s.Context().SwapchainImageResources()

	texInfos := make([]vk.DescriptorImageInfo, 0, len(s.textures))
	for _, tex := range s.textures {
		texInfos = append(texInfos, vk.DescriptorImageInfo{
			Sampler:     tex.sampler,
			ImageView:   tex.view,
			ImageLayout: vk.ImageLayoutGeneral,
		})
	}

	for _, res := range swapchainImageResources {
		var set vk.DescriptorSet
		ret := vk.AllocateDescriptorSets(dev, &vk.DescriptorSetAllocateInfo{
			SType:              vk.StructureTypeDescriptorSetAllocateInfo,
			DescriptorPool:     s.descPool,
			DescriptorSetCount: 1,
			PSetLayouts:        []vk.DescriptorSetLayout{s.descLayout},
		}, &set)
		if ret != vk.Success {
			logger.Log.Error("Failed to allocate descriptor set")
			return
		}
		logger.Log.Info("Allocating descriptor set")
		res.SetDescriptorSet(set)

		vk.UpdateDescriptorSets(dev, 2, []vk.WriteDescriptorSet{{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          set,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeUniformBuffer,
			PBufferInfo: []vk.DescriptorBufferInfo{{
				Offset: 0,
				Range:  vk.DeviceSize(vkTexCubeUniformSize),
				Buffer: res.UniformBuffer(),
			}},
		}, {
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstBinding:      1,
			DstSet:          set,
			DescriptorCount: uint32(len(texEnabled)),
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			PImageInfo:      texInfos,
		}}, 0, nil)
	}
}

func (s *Scene) prepareFramebuffers() {
	dev := s.Context().Device()
	swapchainImageResources := s.Context().SwapchainImageResources()

	for _, res := range swapchainImageResources {
		var fb vk.Framebuffer

		ret := vk.CreateFramebuffer(dev, &vk.FramebufferCreateInfo{
			SType:           vk.StructureTypeFramebufferCreateInfo,
			RenderPass:      s.renderPass,
			AttachmentCount: 2,
			PAttachments: []vk.ImageView{
				res.View(),
				s.depth.view,
			},
			Width:  s.width,
			Height: s.height,
			Layers: 1,
		}, nil, &fb)
		if ret != vk.Success {
			logger.Log.Error("Failed to create framebuffer")
			return
		}
		logger.Log.Info("Creating framebuffer")
		res.SetFramebuffer(fb)
	}
}

func (s *Scene) VulkanContextPrepare() error {
	dim := s.Context().SwapchainDimensions()
	s.height = dim.Height
	s.width = dim.Width

	s.prepareDepth()
	s.prepareTextures()
	s.prepareCubeDataBuffers()
	s.prepareDescriptorLayout()
	s.prepareRenderPass()
	s.preparePipeline()
	s.prepareDescriptorPool()
	s.prepareDescriptorSet()
	s.prepareFramebuffers()

	swapchainImageResources := s.Context().SwapchainImageResources()
	for _, res := range swapchainImageResources {
		s.drawBuildCommandBuffer(res, res.CommandBuffer())
	}
	return nil
}

func (s *Scene) VulkanContextCleanup() error {
	dev := s.Context().Device()
	vk.DestroyDescriptorPool(dev, s.descPool, nil)
	vk.DestroyPipeline(dev, s.pipeline, nil)
	vk.DestroyPipelineCache(dev, s.pipelineCache, nil)
	vk.DestroyRenderPass(dev, s.renderPass, nil)
	vk.DestroyPipelineLayout(dev, s.pipelineLayout, nil)
	vk.DestroyDescriptorSetLayout(dev, s.descLayout, nil)

	for i := 0; i < len(s.textures); i++ {
		s.textures[i].Destroy(dev)
	}
	s.depth.Destroy(dev)
	return nil
}

// Next Frame is called on every frame
func (s *Scene) NextFrame() {
	var Model lin.Mat4x4
	Model.Dup(&s.modelMatrix)
	// Rotate around the Y axis
	s.modelMatrix.Rotate(&Model, 0.0, 1.0, 0.0, lin.DegreesToRadians(s.spinAngle))
}

// Called on every frame???
func (s *Scene) VulkanContextInvalidate(imageIdx int) error {
	dev := s.Context().Device()
	res := s.Context().SwapchainImageResources()[imageIdx]

	var MVP, VP lin.Mat4x4
	VP.Mult(&s.projectionMatrix, &s.viewMatrix)
	MVP.Mult(&VP, &s.modelMatrix)

	data := MVP.Data()
	var pData unsafe.Pointer
	ret := vk.MapMemory(dev, res.UniformMemory(), 0, vk.DeviceSize(len(data)), 0, &pData)
	if ret != vk.Success {
		logger.Log.Error("Failed to map memory")
		return nil
	}
	n := vk.Memcopy(pData, data)
	if n != len(data) {
		logger.Log.Error("Failed to copy memory")
		return nil
	}
	vk.UnmapMemory(dev, res.UniformMemory())
	return nil
}

func (s *Scene) Destroy() {}

type Texture struct {
	sampler vk.Sampler

	image       vk.Image
	imageLayout vk.ImageLayout

	memAlloc *vk.MemoryAllocateInfo
	mem      vk.DeviceMemory
	view     vk.ImageView

	texWidth  int32
	texHeight int32
}

func (t *Texture) Destroy(dev vk.Device) {
	vk.DestroyImageView(dev, t.view, nil)
	vk.FreeMemory(dev, t.mem, nil)
	vk.DestroyImage(dev, t.image, nil)
	vk.DestroySampler(dev, t.sampler, nil)
}

func (t *Texture) DestroyImage(dev vk.Device) {
	vk.FreeMemory(dev, t.mem, nil)
	vk.DestroyImage(dev, t.image, nil)
}

type Depth struct {
	format   vk.Format
	image    vk.Image
	memAlloc *vk.MemoryAllocateInfo
	mem      vk.DeviceMemory
	view     vk.ImageView
}

func (d *Depth) Destroy(dev vk.Device) {
	vk.DestroyImageView(dev, d.view, nil)
	vk.DestroyImage(dev, d.image, nil)
	vk.FreeMemory(dev, d.mem, nil)
}

func loadTextureData(name string, rowPitch int) ([]byte, int, int, error) {
	data := MustAsset(name)
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, 0, 0, err
	}
	newImg := image.NewRGBA(img.Bounds())
	if rowPitch <= 4*img.Bounds().Dy() {
		// apply the proposed row pitch only if supported,
		// as we're using only optimal textures.
		newImg.Stride = rowPitch
	}
	draw.Draw(newImg, newImg.Bounds(), img, image.ZP, draw.Src)
	size := newImg.Bounds().Size()
	return []byte(newImg.Pix), size.X, size.Y, nil
}
