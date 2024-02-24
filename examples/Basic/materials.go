package main

// This is a basic example of how to setup the engine and behaviour packages
import (
	behaviour "Gopher3D/internal/Behaviour"
	loader "Gopher3D/internal/Loader"
	"Gopher3D/internal/engine"
	"Gopher3D/internal/renderer"

	mgl "github.com/go-gl/mathgl/mgl32"
)

type GoCraftBehaviour struct {
	engine     *engine.Gopher
	name       string
	SceneModel *renderer.Model
}

func NewMaterialExampleBehaviour(engine *engine.Gopher) {
	gocraftBehaviour := &GoCraftBehaviour{engine: engine, name: "GoCraft"}
	behaviour.GlobalBehaviourManager.Add(gocraftBehaviour)
}
func main() {
	engine := engine.NewGopher()
	engine.Light = renderer.CreateLight()
	NewMaterialExampleBehaviour(engine)
	engine.Render(768, 50)
}

func (mb *GoCraftBehaviour) Start() {
	createWorld(mb)
}

func (mb *GoCraftBehaviour) Update() {

}

func createWorld(mb *GoCraftBehaviour) {
	model, _ := loader.LoadObjectWithPath("../resources/obj/f104starfighter.obj", true)
	mb.SceneModel = model
	spawnBlock(mb.engine, mb.SceneModel, 0, 0)
}

func spawnBlock(engine *engine.Gopher, model *renderer.Model, x, z int) {
	model.SetPosition(0, 0, 0)
	model.Scale = mgl.Vec3{20.0, 20.0, 20.0}
	engine.AddModel(model)
}
