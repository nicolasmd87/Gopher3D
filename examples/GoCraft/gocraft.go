package main

// This is a basic example of how to setup the engine and behaviour packages
import (
	behaviour "Gopher3D/internal/Behaviour"
	loader "Gopher3D/internal/Loader"
	"Gopher3D/internal/engine"
	"Gopher3D/internal/renderer"
	"fmt"
)

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
	model, err := loader.LoadObjectWithPath("../../tmp/examples/GoCraft/Cube.obj")
	if err != nil {
		fmt.Println("Error loading model")
	}
	renderer.AddModel(model)
	renderer.SetTexture("../../tmp/textures/2k_mars.jpg", model)
	fmt.Println("Behaviour started:", mb.name)
}

func (mb *GoCraftBehaviour) Update() {

}
