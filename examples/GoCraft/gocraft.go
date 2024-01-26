package main

// This is a basic example of how to setup the engine and behaviour packages
import (
	behaviour "Gopher3D/internal/Behaviour"
	loader "Gopher3D/internal/Loader"
	"Gopher3D/internal/engine"
	"Gopher3D/internal/renderer"
	"fmt"
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
	engine.Render(768, 50, nil)
}
func (mb *GoCraftBehaviour) Start() {
	//fmt.Println("Perlin:", p)
	createWorld()
	fmt.Println("Behaviour started:", mb.name)
}

func (mb *GoCraftBehaviour) Update() {

}

// May take a while to load, this is until we fix perfomance issues, this is a good benchmark in the meantime
func createWorld() {
	model, _ := loader.LoadObjectWithPath("../../tmp/examples/GoCraft/Blatt.obj")
	renderer.SetTexture("../../tmp/textures/Blatt.png", model)

	for x := 0; x < 500; x++ {
		for z := 0; z < 500; z++ {
			spawnBlock(*model, x, z)
		}
	}
}

func spawnBlock(model renderer.Model, x, z int) {

	renderer.AddModel(&model)

	y := p.Noise2D(float64(x)*0.1, float64(z)*0.1) // Adjust the multiplier for resolution
	// Get Perlin noise value
	y = scaleNoise(y) // Scale the noise value to your game's scale
	model.SetPosition(float32(x)+1.0, float32(y), float32(z))
}

func scaleNoise(noiseVal float64) float64 {
	// Scale and adjust the noise value to suit the height range of your terrain
	// Example: scale between 0 and 10
	return (noiseVal + 1) / 2 * 10
}
