// sand.go
package main

import (
	behaviour "Gopher3D/internal/Behaviour"
	loader "Gopher3D/internal/Loader"
	"Gopher3D/internal/engine"
	"Gopher3D/internal/renderer"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
	mgl "github.com/go-gl/mathgl/mgl32"
)

func startCPUProfile() {
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
}

func stopCPUProfile() {
	pprof.StopCPUProfile()
}

type Particle struct {
	position mgl.Vec3
	velocity mgl.Vec3
	active   bool // To check if the particle is still active
}

type SandSimulation struct {
	sandParticles []*Particle
	sandModel     *renderer.Model
	engine        *engine.Gopher
	grabbed       *Particle // Particle being grabbed with the mouse
}

func NewSandSimulation(engine *engine.Gopher) {
	ss := &SandSimulation{engine: engine}
	behaviour.GlobalBehaviourManager.Add(ss)
}

func main() {
	engine := engine.NewGopher(engine.OPENGL)
	NewSandSimulation(engine)

	engine.Width = 1980
	engine.Height = 1080

	engine.Render(0, 0)
}

func (ss *SandSimulation) Start() {
	// Set up the camera and light
	ss.engine.Camera.InvertMouse = false
	ss.engine.Camera.Position = mgl.Vec3{0, 100, 300}
	ss.engine.Camera.Speed = 20
	ss.engine.Light = renderer.CreateLight()
	ss.engine.Light.Type = renderer.STATIC_LIGHT

	// Load the sand particle model with instancing enabled
	instances := 50000 // Increased number of particles for a denser effect
	sandModel, err := loader.LoadObjectInstance("../resources/obj/Sphere_Low.obj", true, instances)
	if err != nil {
		panic(err)
	}
	sandModel.Scale = mgl.Vec3{0.5, 0.5, 0.5} // Smaller particles
	sandModel.SetDiffuseColor(210, 180, 140)  // Softer, less intense sand color
	ss.sandModel = sandModel
	ss.engine.AddModel(sandModel)

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Initialize sand particles falling from a height
	for i := 0; i < instances; i++ {
		position := mgl.Vec3{
			rand.Float32()*100 - 50, // Random x position within a certain width
			300 + rand.Float32()*50, // Random height between 300 and 350
			rand.Float32()*100 - 50, // Random z position within a certain depth
		}
		velocity := mgl.Vec3{0, 0, 0} // Initial velocity

		particle := &Particle{
			position: position,
			velocity: velocity,
			active:   true,
		}

		ss.sandParticles = append(ss.sandParticles, particle)
		ss.sandModel.SetInstancePosition(i, position)
	}
}

func (ss *SandSimulation) Update() {
	gravity := mgl.Vec3{0, -5.0, 0} // Gravity
	dt := float32(0.016)            // Time step for updating position

	// Radius and force applied to particles
	radiusOfInfluence := float32(50.0)
	mouseForce := float32(100.0)   // Lowered force for smoother drag
	cohesionForce := float32(50.0) // Cohesion to make particles stick together
	friction := float32(0.8)       // Friction to reduce speed on the plane

	// Get mouse input for interaction
	mousePos := ss.engine.GetMousePosition()
	mousePressed := ss.engine.IsMouseButtonPressed(glfw.MouseButtonLeft)

	// Convert the mouse screen position to world coordinates
	mouseWorldPos := ss.engine.Camera.ScreenToWorld(mousePos, int(ss.engine.Width), int(ss.engine.Height))

	for i := range ss.sandParticles {
		p := ss.sandParticles[i]
		if !p.active {
			continue
		}

		// Apply gravity to each particle
		p.velocity = p.velocity.Add(gravity.Mul(dt))

		// Apply force to particles within the radius only if the mouse is pressed
		if mousePressed && p.position.Sub(mouseWorldPos).Len() < radiusOfInfluence {
			// Calculate direction from particle to mouse
			direction := mouseWorldPos.Sub(p.position).Normalize()
			// Apply force towards the mouse position
			p.velocity = p.velocity.Add(direction.Mul(mouseForce * dt))

			// Apply a small cohesion force to neighboring particles
			for j := range ss.sandParticles {
				if i == j || !ss.sandParticles[j].active {
					continue
				}

				neighbor := ss.sandParticles[j]
				if p.position.Sub(neighbor.position).Len() < radiusOfInfluence {
					// Attract particles within the radius to each other
					cohesionDir := neighbor.position.Sub(p.position).Normalize()
					p.velocity = p.velocity.Add(cohesionDir.Mul(cohesionForce * dt))
				}
			}

			// Debug log
			fmt.Printf("Applying force to particle at %v with velocity %v\n", p.position, p.velocity)
		}

		// Update particle position based on velocity
		p.position = p.position.Add(p.velocity.Mul(dt))

		// Check for collisions with the floor at y = 0
		if p.position.Y() <= 0 {
			p.position[1] = 0                     // Set particle on the floor
			p.velocity = p.velocity.Mul(friction) // Apply friction instead of zeroing the velocity
			p.velocity[1] = 0                     // Keep the Y velocity zero to prevent bouncing
		}

		// Update the instance position in the renderer
		ss.sandModel.SetInstancePosition(i, p.position)
	}

	// Allow input to reset on the next frame if the mouse is not pressed
	if !mousePressed {
		ss.grabbed = nil
	}
}

func (ss *SandSimulation) UpdateFixed() {
	// No fixed update required for this example
}
