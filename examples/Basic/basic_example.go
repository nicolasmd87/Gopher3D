package main

// This is a basic example of how to setup the engine and behaviour packages
import (
	behaviour "Gopher3D/internal/Behaviour"
	"Gopher3D/internal/engine"
)

type TestBehaviour struct {
	engine *engine.Gopher
	name   string
}

func NewTestBehaviour(engine *engine.Gopher) {
	mb := &TestBehaviour{engine: engine, name: "Test"}
	behaviour.GlobalBehaviourManager.Add(mb)
}
func main() {
	engine := engine.NewGopher(engine.VULKAN)
	NewTestBehaviour(engine)
	engine.Render(768, 50)
}
func (mb *TestBehaviour) Start() {
}

func (mb *TestBehaviour) Update() {
}
