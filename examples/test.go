package main

import (
	behaviour "Gopher3D/internal/Behaviour"
	"Gopher3D/internal/engine"
	"fmt"
)

type MyPlayerBehaviour struct {
	name string
}

func NewMyPlayerBehaviour() *MyPlayerBehaviour {
	mb := &MyPlayerBehaviour{}
	behaviour.GlobalBehaviourManager.Add(mb)
	return mb
}
func main() {
	NewMyPlayerBehaviour()
	gopher := engine.NewGopher()
	gopher.Render(768, 50, nil)
}
func (mb *MyPlayerBehaviour) Start() {
	fmt.Println("MyPlayerBehaviour started", mb.name)
}

func (mb *MyPlayerBehaviour) Update() {
	fmt.Println("MyPlayerBehaviour updated", mb.name)
}
