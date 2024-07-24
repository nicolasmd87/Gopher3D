package main

// This is a basic example of how to setup the engine and behaviour packages
import (
	behaviour "Gopher3D/internal/Behaviour"
	loader "Gopher3D/internal/Loader"
	"Gopher3D/internal/engine"
	"Gopher3D/internal/renderer"
	"math/rand"
	"time"

	perlin "github.com/aquilax/go-perlin" // Example Perlin noise library
)

var p = perlin.NewPerlin(2, 2, 3, rand.New(rand.NewSource(time.Now().UnixNano())).Int63())
var modelBatch []*renderer.Model

type GoCraftBehaviour struct {
	engine          *engine.Gopher
	name            string
	worldHeight     int
	worldWidth      int
	noiseDistortion float64
	batchModels     bool
}

func NewGocraftBehaviour(engine *engine.Gopher) {
	gocraftBehaviour := &GoCraftBehaviour{engine: engine, name: "GoCraft"}
	behaviour.GlobalBehaviourManager.Add(gocraftBehaviour)
}
func main() {
	engine := engine.NewGopher(engine.OPENGL) // or engine.VULKAN

	NewGocraftBehaviour(engine)

	// FULLSCREEN
	engine.Width = 1080
	engine.Height = 720

	// WINDOW POS IN X,Y AND MODEL
	engine.Render(0, 0)
}
func (mb *GoCraftBehaviour) Start() {
	mb.engine.Light = renderer.CreateLight()
	//Static light for some extra FPS
	mb.engine.Light.Type = renderer.STATIC_LIGHT
	//mb.engine.SetDebugMode(true)
	createWorld(mb)
}

func (mb *GoCraftBehaviour) Update() {

}

// May take a while to load, this is until we fix perfomance issues, this is a good benchmark in the meantime
func createWorld(mb *GoCraftBehaviour) {
	modelBatch = make([]*renderer.Model, mb.engine.Height*mb.engine.Width)
	model, _ := loader.LoadObjectWithPath("../resources/obj/Cube.obj", true)
	//model.SetTexture("../resources/textures/Grass.png")
	// Tweak this params for fun
	// Warning: When batching is on we can spawn the scene before hand
	// If the height and width are too big, it will take a while to load
	mb.worldHeight = 500
	mb.worldWidth = 500
	mb.noiseDistortion = 10
	mb.batchModels = true
	// Camera frustum culling and face culling for some extra FPS
	mb.engine.SetFrustumCulling(true)
	mb.engine.SetFaceCulling(true)

	// OpenGL and Vulkan use different coordinate systems
	mb.engine.Camera.InvertMouse = false

	InitScene(mb, model)
}

func InitScene(mb *GoCraftBehaviour, model *renderer.Model) {
	if mb.batchModels {
		go func() {
			var index int
			for x := 0; x < mb.worldHeight; x++ {
				for z := 0; z < mb.worldWidth; z++ {
					spawnBlock(mb, *model, x, z, index)
					index++
				}
			}
			mb.engine.ModelBatchChan <- modelBatch
		}()
		return
	}
	for x := 0; x < mb.worldHeight; x++ {
		for z := 0; z < mb.worldWidth; z++ {
			spawnBlock(mb, *model, x, z, 0)
		}
	}

}

func spawnBlock(mb *GoCraftBehaviour, model renderer.Model, x, z, index int) {
	y := p.Noise2D(float64(x)*0.1, float64(z)*0.1)
	y = scaleNoise(mb, y)
	model.SetPosition(float32(x), float32(y), float32(z))
	if mb.batchModels {
		modelBatch[index] = &model
		return
	}
	mb.engine.AddModel(&model)
}

func scaleNoise(mb *GoCraftBehaviour, noiseVal float64) float64 {
	// Scale and adjust the noise value to suit the height range of your terrain
	return (noiseVal / 2) * mb.noiseDistortion
}
