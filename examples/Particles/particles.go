package main

// This is a basic example of how to setup the engine and behaviour packages
import (
	behaviour "Gopher3D/internal/Behaviour"
	loader "Gopher3D/internal/Loader"
	"Gopher3D/internal/engine"
	"Gopher3D/internal/renderer"

	mgl "github.com/go-gl/mathgl/mgl32"
)

type ParticleBehaviour struct {
	engine *engine.Gopher
	name   string
	model  *renderer.Model
}

func NewParticleBehaviour(engine *engine.Gopher) {
	mb := &ParticleBehaviour{engine: engine, name: "Particle Behaviour"}
	behaviour.GlobalBehaviourManager.Add(mb)
}
func main() {
	engine := engine.NewGopher(engine.OPENGL) // or engine.VULKAN
	NewParticleBehaviour(engine)
	engine.Width = 720
	engine.Height = 480

	// WINDOW POS IN X,Y AND MODEL
	engine.Render(500, 500)
}
func (pb *ParticleBehaviour) Start() {
	pb.engine.Light = renderer.CreateLight()
	//Static light for some extra FPS
	pb.engine.Light.Type = renderer.STATIC_LIGHT
	pb.engine.SetFrustumCulling(false)
	pb.engine.SetFaceCulling(false)
	// Invert mouse for OpenGL
	pb.engine.Camera.InvertMouse = false
	m, err := loader.LoadObjectWithPath("../resources/obj/Sphere.obj", true)
	if err != nil {
		panic(err)
	}
	// I want to increase the scale of the model
	m.Scale = mgl.Vec3{0.1, 0.1, 0.1}
	pb.model = m
	pb.model.SetSpecularColor(100.0, 100.0, 100.0)
	pb.model.SetDiffuseColor(100.0, 100.0, 100.0)
	pb.engine.AddModel(m)
}

func (pb *ParticleBehaviour) Update() {

}
