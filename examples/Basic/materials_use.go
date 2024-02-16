package main

// This is a basic example of how to setup the engine and behaviour packages
import (
	behaviour "Gopher3D/internal/Behaviour"
	"Gopher3D/internal/engine"
	"fmt"
)

type MaterialsTestBehaviour struct {
	engine *engine.Gopher
	name   string
}

func NewMaterialsTesBehaviour(engine *engine.Gopher) {
	mb := &MaterialsTestBehaviour{engine: engine, name: "Materials Test"}
	behaviour.GlobalBehaviourManager.Add(mb)
}
func main() {
	engine := engine.NewGopher()
	NewMaterialsTesBehaviour(engine)
	engine.Render(768, 50)
}
func (mb *MaterialsTestBehaviour) Start() {
	fmt.Println("Behaviour started:", mb.name)
}

func (mb *MaterialsTestBehaviour) Update() {
	fmt.Println("Frame update for:", mb.name)
}
