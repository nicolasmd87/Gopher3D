package renderer

import (
	"Gopher3D/internal/logger"
	"fmt"
	"image"
	"image/draw"
	"os"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"go.uber.org/zap"
)

var currentTextureID uint32 = ^uint32(0) // Initialize with an invalid value
var frustum Frustum

type OpenGLRenderer struct {
	FrustumCullingEnabled bool
	FaceCullingEnabled    bool
	Debug                 bool
	modelLoc              int32
	viewProjLoc           int32
	lightPosLoc           int32
	lightColorLoc         int32
	lightIntensityLoc     int32
	diffuseColorUniform   int32
	shininessUniform      int32
	specularColorUniform  int32
	textureUniform        int32
	vertexShader          uint32
	fragmentShader        uint32
	Shader                Shader
	Models                []*Model
}

func (rend *OpenGLRenderer) Init(width, height int32, _ *glfw.Window) {
	if err := gl.Init(); err != nil {
		logger.Log.Error("OpenGL initialization failed", zap.Error(err))
		return
	}

	if rend.Debug {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	}

	rend.FrustumCullingEnabled = false
	rend.FaceCullingEnabled = false
	SetDefaultTexture(rend)
	gl.Viewport(0, 0, width, height)
	rend.Shader = InitShader()
	rend.vertexShader = genShader(rend.Shader.vertexSource, gl.VERTEX_SHADER)
	rend.fragmentShader = genShader(rend.Shader.fragmentSource, gl.FRAGMENT_SHADER)
	rend.Shader.program = genShaderProgram(rend.vertexShader, rend.fragmentShader)

	gl.UseProgram(rend.Shader.program)

	rend.modelLoc = gl.GetUniformLocation(rend.Shader.program, gl.Str("model\x00"))
	rend.viewProjLoc = gl.GetUniformLocation(rend.Shader.program, gl.Str("viewProjection\x00"))

	// Set light properties for each model
	rend.lightPosLoc = gl.GetUniformLocation(rend.Shader.program, gl.Str("light.position\x00"))
	rend.lightColorLoc = gl.GetUniformLocation(rend.Shader.program, gl.Str("light.color\x00"))
	rend.lightIntensityLoc = gl.GetUniformLocation(rend.Shader.program, gl.Str("light.intensity\x00"))
	// Set material properties for each model
	rend.diffuseColorUniform = gl.GetUniformLocation(rend.Shader.program, gl.Str("diffuseColor\x00"))
	rend.shininessUniform = gl.GetUniformLocation(rend.Shader.program, gl.Str("shininess\x00"))
	rend.specularColorUniform = gl.GetUniformLocation(rend.Shader.program, gl.Str("specularColor\x00"))
	// Set texture properties for each model
	rend.textureUniform = gl.GetUniformLocation(rend.Shader.program, gl.Str("uTexture\x00"))

	logger.Log.Info("OpenGL render initialized")
}

func (rend *OpenGLRenderer) AddModel(model *Model) {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(model.InterleavedData)*4, gl.Ptr(model.InterleavedData), gl.STATIC_DRAW)

	var ebo uint32
	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(model.Faces)*4, gl.Ptr(model.Faces), gl.STATIC_DRAW)

	stride := int32((3 + 2 + 3) * 4)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, stride, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, stride, gl.PtrOffset(5*4))
	gl.EnableVertexAttribArray(2)

	model.VAO = vao
	model.VBO = vbo
	model.EBO = ebo
	model.ModelMatrix = mgl32.Ident4()

	rend.Models = append(rend.Models, model)
}

func (rend *OpenGLRenderer) Render(camera Camera, light *Light) {

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	viewProjection := camera.GetViewProjection()
	gl.UseProgram(rend.Shader.program)
	gl.UniformMatrix4fv(rend.viewProjLoc, 1, false, &viewProjection[0])

	if light != nil && light.Mode == "static" && !light.Calculated {
		rend.calculateLights(light)
		light.Calculated = true
	} else if light != nil && !light.Calculated {
		rend.calculateLights(light)
	}

	gl.Enable(gl.DEPTH_TEST)

	// Culling : https://learnopengl.com/Advanced-OpenGL/Face-culling
	if rend.FaceCullingEnabled {
		gl.Enable(gl.CULL_FACE)
		// IF FACES OF THE MODEL ARE RENDERED IN THE WRONG ORDER, TRY SWITCHING THE FOLLOWING LINE TO gl.CCW or we need to make sure the winding of each model is consistent
		// CCW = Counter ClockWise
		gl.CullFace(gl.FRONT)
		gl.FrontFace(gl.CW)
	}

	// Calculate frustum

	// TODO: Add check to see if camera is dirty(moved)
	if rend.FrustumCullingEnabled {
		frustum = camera.CalculateFrustum()
	}

	modLen := len(rend.Models)
	for i := 0; i < modLen; i++ {
		// Skip rendering if the model is outside the frustum
		if rend.FrustumCullingEnabled && !frustum.IntersectsSphere(rend.Models[i].BoundingSphereCenter, rend.Models[i].BoundingSphereRadius) {
			continue
		}

		if rend.Models[i].IsDirty {
			// Recalculate the model matrix only if necessary
			rend.Models[i].ModelMatrix = CalculateModelMatrix(*rend.Models[i])
			rend.Models[i].IsDirty = false
		}
		// Upload the model matrix to the GPU
		gl.UniformMatrix4fv(rend.modelLoc, 1, false, &rend.Models[i].ModelMatrix[0])

		// Bind material's texture if available
		if rend.Models[i].Material != nil {
			if rend.Models[i].Material.TextureID != currentTextureID {
				gl.BindTexture(gl.TEXTURE_2D, rend.Models[i].Material.TextureID)
				currentTextureID = rend.Models[i].Material.TextureID
			}
			gl.Uniform3fv(rend.diffuseColorUniform, 1, &rend.Models[i].Material.DiffuseColor[0])
			gl.Uniform3fv(rend.specularColorUniform, 1, &rend.Models[i].Material.SpecularColor[0])
			gl.Uniform1f(rend.shininessUniform, rend.Models[i].Material.Shininess)
		}

		// Set the sampler to the first texture unit
		gl.Uniform1i(rend.textureUniform, 0)
		gl.BindVertexArray(rend.Models[i].VAO)

		gl.DrawElements(gl.TRIANGLES, int32(len(rend.Models[i].Faces)), gl.UNSIGNED_INT, nil)
		gl.BindVertexArray(0)
	}
	gl.Disable(gl.DEPTH_TEST)
	// Disable culling after rendering
	gl.Disable(gl.CULL_FACE)
}

func (rend *OpenGLRenderer) calculateLights(light *Light) {
	gl.Uniform3f(rend.lightPosLoc, light.Position[0], light.Position[1], light.Position[2])
	gl.Uniform3f(rend.lightColorLoc, light.Color[0], light.Color[1], light.Color[2])
	gl.Uniform1f(rend.lightIntensityLoc, light.Intensity)
}

func (rend *OpenGLRenderer) Cleanup() {
	for _, model := range rend.Models {
		gl.DeleteVertexArrays(1, &model.VAO)
		gl.DeleteBuffers(1, &model.VBO)
		gl.DeleteBuffers(1, &model.EBO)
	}
}

func (rend *OpenGLRenderer) LoadTexture(filePath string) (uint32, error) { // TODO: Consider specifying image format or handling different formats properly

	imgFile, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0, err
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var textureID uint32
	// Like any of the previous objects in OpenGL, textures are referenced with an ID; let's create one:
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D, textureID)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(rgba.Rect.Size().X), int32(rgba.Rect.Size().Y), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))

	// Set texture parameters (optional)
	// gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.REPEAT)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	// GL_NEAREST results in blocked patterns where we can clearly see the pixels that form the texture while GL_LINEAR produces a smoother pattern where the individual pixels are less visible.
	// GL_LINEAR produces a more realistic output, but some developers prefer a more 8-bit look and as a result pick the GL_NEAREST option
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	return textureID, nil
}

func (rend *OpenGLRenderer) CreateTextureFromImage(img image.Image) (uint32, error) {
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D, textureID)

	rgba, ok := img.(*image.RGBA)
	if !ok {
		// Convert to *image.RGBA if necessary
		b := img.Bounds()
		rgba = image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(rgba, rgba.Bounds(), img, b.Min, draw.Src)
	}

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(rgba.Rect.Size().X), int32(rgba.Rect.Size().Y), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))

	// Set texture parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	return textureID, nil
}

func genShader(source string, shaderType uint32) uint32 {
	shader := gl.CreateShader(shaderType)
	cSources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, cSources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		logger.Log.Error("Failed to compile", zap.Uint32("shader type:", shaderType), zap.String("log", log))
	}

	return shader
}

func genShaderProgram(vertexShader, fragmentShader uint32) uint32 {
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		logger.Log.Error("Failed to link program", zap.String("log", log))
	}
	gl.DetachShader(program, vertexShader)
	gl.DeleteShader(vertexShader)
	gl.DetachShader(program, fragmentShader)
	gl.DeleteShader(fragmentShader)
	return program
}

// CalculateModelMatrix calculates the transformation matrix for a model
func CalculateModelMatrix(model Model) mgl32.Mat4 {
	// Start with an identity matrix
	modelMatrix := mgl32.Ident4()

	// Apply scale, rotation, and translation
	modelMatrix = modelMatrix.Mul4(mgl32.Scale3D(model.Scale.X(), model.Scale.Y(), model.Scale.Z()))
	modelMatrix = modelMatrix.Mul4(mgl32.Translate3D(model.Position.X(), model.Position.Y(), model.Position.Z()))
	modelMatrix = modelMatrix.Mul4(model.Rotation.Mat4())

	return modelMatrix
}

func CreateLight() *Light {
	return &Light{
		Position:  mgl32.Vec3{0.0, 300.0, 0.0}, // Example position
		Color:     mgl32.Vec3{1.0, 1.0, 1.0},   // White light
		Intensity: 1.0,                         // Full intensity
	}
}
