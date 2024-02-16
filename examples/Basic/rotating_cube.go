package main

// This is a basic example of how to setup the engine and behaviour packages
import (
	behaviour "Gopher3D/internal/Behaviour"
	loader "Gopher3D/internal/Loader"
	"Gopher3D/internal/engine"
	"Gopher3D/internal/renderer"
	"math/rand"
	"time"

	mgl "github.com/go-gl/mathgl/mgl32"

	perlin "github.com/aquilax/go-perlin" // Example Perlin noise library
)

var p = perlin.NewPerlin(2, 2, 3, rand.New(rand.NewSource(time.Now().UnixNano())).Int63())

type GoCraftBehaviour struct {
	engine     *engine.Gopher
	name       string
	SceneModel *renderer.Model
}

func NewGocraftBehaviour(engine *engine.Gopher) {
	gocraftBehaviour := &GoCraftBehaviour{engine: engine, name: "GoCraft"}
	behaviour.GlobalBehaviourManager.Add(gocraftBehaviour)
}
func main() {
	engine := engine.NewGopher()
	engine.Light = renderer.CreateLight()
	NewGocraftBehaviour(engine)
	engine.Render(768, 50)
}
func (mb *GoCraftBehaviour) Start() {
	createWorld(mb)
}

func (mb *GoCraftBehaviour) Update() {
	if mb.SceneModel != nil {
		// TODO: NEED TO FIX THIS, it doesn't rotate
		mb.SceneModel.RotateModel(1.0, 1.0, 0.0)
	}
}

// May take a while to load, this is until we fix perfomance issues, this is a good benchmark in the meantime
func createWorld(mb *GoCraftBehaviour) {
	model, _ := loader.LoadObjectWithPath("../tmp/examples/GoCraft/Cube.obj", true)
	model.SetTexture("../tmp/textures/Blatt.png")
	mb.SceneModel = model
	spawnBlock(mb.SceneModel, 0, 0)
}

func spawnBlock(model *renderer.Model, x, z int) {
	model.SetPosition(0, 0, 0)
	model.Scale = mgl.Vec3{20.0, 20.0, 20.0}
	renderer.AddModel(model)
}
