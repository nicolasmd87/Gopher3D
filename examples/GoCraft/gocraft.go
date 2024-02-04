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

type GoCraftBehaviour struct {
	engine *engine.Gopher
	name   string
}

func NewGocraftBehaviour(engine *engine.Gopher) {
	gocraftBehaviour := &GoCraftBehaviour{engine: engine, name: "GoCraft"}
	behaviour.GlobalBehaviourManager.Add(gocraftBehaviour)
}
func main() {
	engine := engine.NewGopher()

	NewGocraftBehaviour(engine)
	engine.Light = renderer.CreateLight()
	//Static light for some extra FPS
	engine.Light.Type = renderer.STATIC_LIGHT
	// FULLSCREEN
	engine.Width = 1980
	engine.Height = 1080
	// WINDOW POS IN X,Y AND MODEL
	engine.Render(0, 0, nil)
}
func (mb *GoCraftBehaviour) Start() {
	createWorld()
}

func (mb *GoCraftBehaviour) Update() {

}

// May take a while to load, this is until we fix perfomance issues, this is a good benchmark in the meantime
func createWorld() {
	model, _ := loader.LoadObjectWithPath("../../tmp/examples/GoCraft/Cube.obj", true)
	renderer.SetTexture("../../tmp/textures/Blatt.png", model)

	for x := 0; x < 5000; x++ {
		for z := 0; z < 5000; z++ {
			spawnBlock(*model, x, z)
		}
	}
}

func spawnBlock(model renderer.Model, x, z int) {

	renderer.AddModel(&model)

	y := p.Noise2D(float64(x)*0.1, float64(z)*0.1) // Adjust the multiplier for resolution
	// Get Perlin noise value
	y = scaleNoise(y) // Scale the noise value to your game's scale
	model.SetPosition(float32(x), float32(y), float32(z))
}

func scaleNoise(noiseVal float64) float64 {
	// Scale and adjust the noise value to suit the height range of your terrain
	// Example: scale between 0 and 10
	return (noiseVal + 1) / 2 * 10
}
