package renderer

import (
	"Gopher3D/internal/logger"
	"fmt"
	"image"
	"image/draw"
	"math"
	"os"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"go.uber.org/zap"
)

type Model struct {
	Id                   int
	Position             mgl32.Vec3
	Scale                mgl32.Vec3
	Rotation             mgl32.Quat
	Vertices             []float32
	Normals              []float32
	Faces                []int32
	TextureCoords        []float32
	InterleavedData      []float32
	TextureID            uint32
	VAO                  uint32 // Vertex Array Object
	VBO                  uint32 // Vertex Buffer Object
	EBO                  uint32 // Element Buffer Object
	ModelMatrix          mgl32.Mat4
	BoundingSphereCenter mgl32.Vec3
	BoundingSphereRadius float32
	IsDirty              bool
	IsBatched            bool
}

var (
	Models            []*Model
	shaderProgram     uint32
	modelLoc          int32
	viewProjLoc       int32
	lightPosLoc       int32
	lightColorLoc     int32
	lightIntensityLoc int32
	Debug             bool = false
)

// =============================================================
//
//	Shaders
//
// =============================================================
var vertexShaderSource = `#version 330 core
layout(location = 0) in vec3 inPosition; // Vertex position
layout(location = 1) in vec2 inTexCoord; // Texture Coordinate
layout(location = 2) in vec3 inNormal;   // Vertex normal



uniform mat4 model;
uniform mat4 viewProjection;
out vec2 fragTexCoord;   // Pass to fragment shader
out vec3 Normal;         // Pass normal to fragment shader
out vec3 FragPos;        // Pass position to fragment shader

void main() {
    FragPos = vec3(model * vec4(inPosition, 1.0));
	// Vertex Shader
	Normal = mat3(model) * inNormal; // Use this if the model matrix has no non-uniform scaling
    fragTexCoord = inTexCoord;
    gl_Position = viewProjection * model * vec4(inPosition, 1.0);
}
` + "\x00"

var fragmentShaderSource = `
// Fragment Shader
#version 330 core
in vec2 fragTexCoord;
in vec3 Normal;
in vec3 FragPos;

uniform sampler2D textureSampler;
uniform struct Light {
    vec3 position;
    vec3 color;
    float intensity;
} light;
uniform vec3 viewPos;
out vec4 FragColor;

void main() {
    vec4 texColor = texture(textureSampler, fragTexCoord);
    float ambientStrength = 0.1;
    vec3 ambient = ambientStrength * light.color;
    vec3 norm = normalize(Normal);
    vec3 lightDir = normalize(light.position - FragPos);
    float diff = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = diff * light.color;
    vec3 result = (ambient + diffuse) * light.intensity;
    FragColor = vec4(result, 1.0) * texColor;
}
` + "\x00"

type LightType int

const (
	STATIC_LIGHT LightType = iota
	DYNAMIC_LIGHT
)

type Light struct {
	Position   mgl32.Vec3
	Color      mgl32.Vec3
	Intensity  float32
	Type       LightType // "dynamic", "static
	Mode       string    // "directional", "point", "spot"
	Calculated bool
}

func Init(width, height int32) {
	if err := gl.Init(); err != nil {
		logger.Log.Error("OpenGL initialization failed", zap.Error(err))
		return
	}

	gl.Viewport(0, 0, width, height)
	initOpenGL()
}

func initOpenGL() {
	vertexShader := genShader(vertexShaderSource, gl.VERTEX_SHADER)
	fragmentShader := genShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	shaderProgram = genShaderProgram(vertexShader, fragmentShader)

	gl.UseProgram(shaderProgram)

	modelLoc = gl.GetUniformLocation(shaderProgram, gl.Str("model\x00"))
	viewProjLoc = gl.GetUniformLocation(shaderProgram, gl.Str("viewProjection\x00"))

	// Set light properties
	lightPosLoc = gl.GetUniformLocation(shaderProgram, gl.Str("light.position\x00"))
	lightColorLoc = gl.GetUniformLocation(shaderProgram, gl.Str("light.color\x00"))
	lightIntensityLoc = gl.GetUniformLocation(shaderProgram, gl.Str("light.intensity\x00"))

}

func AddModel(model *Model) {
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

	Models = append(Models, model)
}

func Render(camera Camera, deltaTime float64, light *Light) {
	var currentTextureID uint32
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	if Debug {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	} else {
		// Switch back to solid fill mode
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	}

	viewProjection := camera.GetViewProjection()
	gl.UseProgram(shaderProgram)
	gl.UniformMatrix4fv(viewProjLoc, 1, false, &viewProjection[0])

	if light.Mode == "static" && !light.Calculated {
		// We only calculate it once to save performance
		calculateLights(light)
		light.Calculated = true
	} else if !light.Calculated {
		calculateLights(light)
	}

	// Get the uniform location of the texture sampler in your shader program
	textureUniform := gl.GetUniformLocation(shaderProgram, gl.Str("uTexture\x00"))
	gl.Enable(gl.DEPTH_TEST)
	// Culling : https://learnopengl.com/Advanced-OpenGL/Face-culling
	gl.Enable(gl.CULL_FACE)
	// IF FACES OF THE MODEL ARE RENDERED IN THE WRONG ORDER, TRY SWITCHING THE FOLLOWING LINE TO gl.CCW or we need to make sure the winding of each model is consistent
	// CCW = Counter ClockWise
	gl.CullFace(gl.FRONT)
	gl.FrontFace(gl.CW)

	// Calculate frustum
	frustum := camera.CalculateFrustum()
	for _, model := range Models {
		// Skip rendering if the model is outside the frustum
		if !frustum.IntersectsSphere(model.BoundingSphereCenter, model.BoundingSphereRadius) {
			continue
		}

		if !model.IsBatched {
			if model.IsDirty {
				// Recalculate the model matrix only if necessary
				model.ModelMatrix = CalculateModelMatrix(*model)
				model.IsDirty = false
			}
			// Upload the model matrix to the GPU
			gl.UniformMatrix4fv(modelLoc, 1, false, &model.ModelMatrix[0])
		} else {
			// For batched models, you might use an identity matrix or skip setting the model matrix altogether.
			// This depends on whether you pre-transform your vertices or not.
			var identityMatrix = mgl32.Ident4()
			gl.UniformMatrix4fv(modelLoc, 1, false, &identityMatrix[0])
		}

		if model.TextureID != currentTextureID {
			gl.BindTexture(gl.TEXTURE_2D, model.TextureID)
			currentTextureID = model.TextureID
		}

		// Set the sampler to the first texture unit
		gl.Uniform1i(textureUniform, 0)

		gl.DrawElements(gl.TRIANGLES, int32(len(model.Faces)), gl.UNSIGNED_INT, nil)
	}
	// Disable culling after rendering
	gl.Disable(gl.CULL_FACE)
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

func Cleanup() {
	for _, model := range Models {
		gl.DeleteVertexArrays(1, &model.VAO)
		gl.DeleteBuffers(1, &model.VBO)
		gl.DeleteBuffers(1, &model.EBO)
	}
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
func CreateLight() *Light {
	return &Light{
		Position:  mgl32.Vec3{0.0, 300.0, 0.0}, // Example position
		Color:     mgl32.Vec3{1.0, 1.0, 1.0},   // White light
		Intensity: 1.0,                         // Full intensity
	}
}

func (m *Model) RotateModel(angleX, angleY float32, angleZ float32) {
	// Create quaternions for each axis
	rotationX := mgl32.QuatRotate(mgl32.DegToRad(angleX), mgl32.Vec3{1, 0, 0})
	rotationY := mgl32.QuatRotate(mgl32.DegToRad(angleY), mgl32.Vec3{0, 1, 0})
	rotationZ := mgl32.QuatRotate(mgl32.DegToRad(angleZ), mgl32.Vec3{0, 0, 1})

	// Combine new rotation with existing rotation
	m.Rotation = m.Rotation.Mul(rotationX).Mul(rotationY).Mul(rotationZ)
	m.IsDirty = true
}

// SetPosition sets the position of the model
func (m *Model) SetPosition(x, y, z float32) {
	m.ModelMatrix = mgl32.Translate3D(x, y, z)
	m.Position = mgl32.Vec3{x, y, z}
	m.CalculateBoundingSphere()
	m.IsDirty = true
}

func (m *Model) CalculateBoundingSphere() {
	var center mgl32.Vec3
	var maxDistanceSq float32

	numVertices := len(m.Vertices) / 3 // Assuming 3 float32s per vertex
	for i := 0; i < numVertices; i++ {
		// Extracting vertex from the flat array
		vertex := mgl32.Vec3{m.Vertices[i*3], m.Vertices[i*3+1], m.Vertices[i*3+2]}
		transformedVertex := ApplyModelTransformation(vertex, m.Position, m.Scale, m.Rotation)
		center = center.Add(transformedVertex)
	}
	center = center.Mul(1.0 / float32(numVertices))

	for i := 0; i < numVertices; i++ {
		vertex := mgl32.Vec3{m.Vertices[i*3], m.Vertices[i*3+1], m.Vertices[i*3+2]}
		transformedVertex := ApplyModelTransformation(vertex, m.Position, m.Scale, m.Rotation)
		distanceSq := transformedVertex.Sub(center).LenSqr()
		if distanceSq > maxDistanceSq {
			maxDistanceSq = distanceSq
		}
	}

	m.BoundingSphereCenter = center
	m.BoundingSphereRadius = float32(math.Sqrt(float64(maxDistanceSq)))
}

func ApplyModelTransformation(vertex, position, scale mgl32.Vec3, rotation mgl32.Quat) mgl32.Vec3 {
	// Apply scaling
	scaledVertex := mgl32.Vec3{vertex[0] * scale[0], vertex[1] * scale[1], vertex[2] * scale[2]}

	// Apply rotation
	// Note: mgl32.Quat doesn't directly multiply with Vec3, so we convert it to a Mat4 first
	rotatedVertex := rotation.Mat4().Mul4x1(scaledVertex.Vec4(1)).Vec3()

	// Apply translation
	transformedVertex := rotatedVertex.Add(position)

	return transformedVertex
}

func SetTexture(texturePath string, model *Model) {
	textureID, _ := loadTexture(texturePath)
	model.TextureID = textureID // Store the texture ID in the Model struct
}

func loadTexture(filePath string) (uint32, error) { // Consider specifying image format or handling different formats properly

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

func calculateLights(light *Light) {
	gl.Uniform3f(lightPosLoc, light.Position[0], light.Position[1], light.Position[2])
	gl.Uniform3f(lightColorLoc, light.Color[0], light.Color[1], light.Color[2])
	gl.Uniform1f(lightIntensityLoc, light.Intensity)
}
