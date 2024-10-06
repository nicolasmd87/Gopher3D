package renderer

import (
	"Gopher3D/internal/logger"
	"bytes"
	"embed"
	"image"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	vk "github.com/vulkan-go/vulkan"
	"go.uber.org/zap"
)

// DefaultMaterial provides a basic material to fall back on
var DefaultMaterial = &Material{
	Name:          "default",
	DiffuseColor:  [3]float32{1.0, 1.0, 1.0}, // White color
	SpecularColor: [3]float32{1.0, 1.0, 1.0},
	Shininess:     32.0,
	TextureID:     0,
}

//go:embed resources/default.png
var defaultTextureFS embed.FS

type Model struct {
	Id                    int
	Name                  string
	Position              mgl32.Vec3
	Scale                 mgl32.Vec3
	Rotation              mgl32.Quat
	Vertices              []float32
	Indices               []uint32
	vertexBuffer          vk.Buffer
	vertexMemory          vk.DeviceMemory
	indexBuffer           vk.Buffer
	indexMemory           vk.DeviceMemory
	Normals               []float32
	Faces                 []int32
	TextureCoords         []float32
	InterleavedData       []float32
	Material              *Material
	VAO                   uint32 // Vertex Array Object
	VBO                   uint32 // Vertex Buffer Object
	EBO                   uint32 // Element Buffer Object
	ModelMatrix           mgl32.Mat4
	BoundingSphereCenter  mgl32.Vec3
	BoundingSphereRadius  float32
	IsDirty               bool
	IsBatched             bool
	IsInstanced           bool
	InstanceCount         int
	InstanceModelMatrices []mgl32.Mat4 // Instance model matrices
}

type Material struct {
	Name          string
	DiffuseColor  [3]float32
	SpecularColor [3]float32
	Shininess     float32
	TextureID     uint32 // OpenGL texture ID
}

func (m *Model) X() float32 {
	return m.Position[0]
}

func (m *Model) Y() float32 {
	return m.Position[1]
}

func (m *Model) Z() float32 {
	return m.Position[2]
}

func (m *Model) Rotate(angleX, angleY, angleZ float32) {
	if m.Rotation == (mgl32.Quat{}) {
		m.Rotation = mgl32.QuatIdent()
	}
	rotationX := mgl32.QuatRotate(mgl32.DegToRad(angleX), mgl32.Vec3{1, 0, 0})
	rotationY := mgl32.QuatRotate(mgl32.DegToRad(angleY), mgl32.Vec3{0, 1, 0})
	rotationZ := mgl32.QuatRotate(mgl32.DegToRad(angleZ), mgl32.Vec3{0, 0, 1})
	m.Rotation = m.Rotation.Mul(rotationX).Mul(rotationY).Mul(rotationZ)
	m.updateModelMatrix()
	m.IsDirty = true
}

// SetPosition sets the position of the model
func (m *Model) SetPosition(x, y, z float32) {
	m.Position = mgl32.Vec3{x, y, z}
	m.CalculateBoundingSphere()
	m.updateModelMatrix()
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

func (m *Model) updateModelMatrix() {
	// Matrix multiplication order: scale -> rotate -> translate
	scaleMatrix := mgl32.Scale3D(m.Scale[0], m.Scale[1], m.Scale[2])
	rotationMatrix := m.Rotation.Mat4()
	translationMatrix := mgl32.Translate3D(m.Position[0], m.Position[1], m.Position[2])
	// Combine the transformations: ModelMatrix = translation * rotation * scale
	m.ModelMatrix = translationMatrix.Mul4(rotationMatrix).Mul4(scaleMatrix)
	m.CalculateBoundingSphere()
}

// Aux functions, maybe I need to move them to another package
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

func (m *Model) SetDiffuseColor(r, g, b float32) {
	if m.Material == nil {
		logger.Log.Info("Setting default material")
		m.Material = DefaultMaterial
	}
	m.Material.DiffuseColor = [3]float32{r, g, b}
	m.IsDirty = true // Mark the model as dirty for re-rendering
}

func (m *Model) SetSpecularColor(r, g, b float32) {
	if m.Material == nil {
		logger.Log.Info("Setting default material")
		m.Material = DefaultMaterial
	}
	m.Material.SpecularColor = [3]float32{r, g, b}
	m.IsDirty = true // Mark the model as dirty for re-rendering
}

func (m *Model) SetTexture(texturePath string) {
	// TODO: Use THE CONFIG to know which renderer to use
	textureID, err := (&OpenGLRenderer{}).LoadTexture(texturePath)
	if err != nil {
		logger.Log.Error("Failed to load texture", zap.String("path", texturePath), zap.Error(err))
		return
	}

	if m.Material == nil {
		logger.Log.Info("Setting default material")
		m.Material = DefaultMaterial

	}
	m.Material.TextureID = textureID
}

func SetDefaultTexture(RendererAPI Render) {
	// Read the embedded texture
	textureBytes, err := defaultTextureFS.ReadFile("resources/default.png")
	if err != nil {
		logger.Log.Error("Failed to read embedded default texture", zap.Error(err))
		return
	}

	// Create an image from the texture bytes
	img, _, err := image.Decode(bytes.NewReader(textureBytes))
	if err != nil {
		logger.Log.Error("Failed to decode embedded default texture", zap.Error(err))
		return
	}

	// Convert the image to a texture and set it as the default texture
	// TODO: It should use the renderer API to create the texture and not an OpenGL-specific function
	textureID, err := RendererAPI.CreateTextureFromImage(img)
	if err != nil {
		logger.Log.Error("Failed to create texture from embedded default image", zap.Error(err))
		return
	}

	DefaultMaterial.TextureID = textureID
}
