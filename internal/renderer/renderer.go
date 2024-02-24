package renderer

import (
	"embed"

	"github.com/go-gl/mathgl/mgl32"
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
	Material             *Material
	VAO                  uint32 // Vertex Array Object
	VBO                  uint32 // Vertex Buffer Object
	EBO                  uint32 // Element Buffer Object
	ModelMatrix          mgl32.Mat4
	BoundingSphereCenter mgl32.Vec3
	BoundingSphereRadius float32
	IsDirty              bool
	IsBatched            bool
}

type Material struct {
	Name          string
	DiffuseColor  [3]float32
	SpecularColor [3]float32
	Shininess     float32
	TextureID     uint32 // OpenGL texture ID
}

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

type LightType int

const (
	STATIC_LIGHT LightType = iota
	DYNAMIC_LIGHT
)

type Light struct {
	Position   mgl32.Vec3
	Color      mgl32.Vec3
	Intensity  float32
	Type       LightType // "static", "dynamic"
	Mode       string    // "directional", "point", "spot"
	Calculated bool
}

type Render interface {
	Init(width, height int32)
	Render(camera Camera, deltaTime float64, light *Light)
	AddModel(model *Model)
	Cleanup()
}
