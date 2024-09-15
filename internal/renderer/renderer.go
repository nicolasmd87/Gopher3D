package renderer

import (
	"image"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type LightType int

var FrustumCullingEnabled bool = false
var FaceCullingEnabled bool = false
var Debug bool = false

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
	Init(width, height int32, window *glfw.Window)
	Render(camera Camera, light *Light)
	AddModel(model *Model)
	RemoveModel(model *Model)
	LoadTexture(path string) (uint32, error)
	CreateTextureFromImage(img image.Image) (uint32, error)
	Cleanup()
}
