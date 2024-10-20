// sand.go
package main

import (
	behaviour "Gopher3D/internal/Behaviour"
	loader "Gopher3D/internal/Loader"
	"Gopher3D/internal/engine"
	"Gopher3D/internal/renderer"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/g3n/engine/experimental/physics"
	"github.com/g3n/engine/math32"
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
	active   bool
	grabbed  bool // To determine if the particle is currently grabbed
}

type SandSimulation struct {
	sandParticles []*Particle
	sandModel     *renderer.Model
	engine        *engine.Gopher
	forceField    *physics.AttractorForceField // Force field for gravity-like effects
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
	ss.engine.Camera.Speed = 200 // Set camera speed to 200
	ss.engine.Light = renderer.CreateLight()
	ss.engine.Light.Type = renderer.STATIC_LIGHT
	ss.engine.Light.Intensity = 0.05                    // Tone down the light intensity even more
	ss.engine.Light.Position = mgl.Vec3{0, 1500, -1000} // Move the light further away

	// Load the sand particle model with instancing enabled
	instances := 100000 // Increased number of particles for denser effect
	sandModel, err := loader.LoadObjectInstance("../resources/obj/Sphere_Low.obj", true, instances)
	if err != nil {
		panic(err)
	}
	sandModel.Scale = mgl.Vec3{0.2, 0.2, 0.2} // Smaller particles
	sandModel.SetDiffuseColor(139, 69, 19)    // Dark brown (RGB: 139, 69, 19)
	ss.sandModel = sandModel
	ss.engine.AddModel(sandModel)

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Initialize sand particles
	for i := 0; i < instances; i++ {
		position := mgl.Vec3{
			rand.Float32()*200 - 100, // Larger spread for more particles
			300 + rand.Float32()*50,
			rand.Float32()*200 - 100,
		}
		velocity := mgl.Vec3{0, 0, 0}

		particle := &Particle{
			position: position,
			velocity: velocity,
			active:   true,
		}

		ss.sandParticles = append(ss.sandParticles, particle)
		ss.sandModel.SetInstancePosition(i, position)
	}

	// Set realistic gravitational force (scaled down for simulation)
	gravity := &math32.Vector3{0, -9.81, 0}
	ss.forceField = physics.NewAttractorForceField(gravity, 1)
}

func (ss *SandSimulation) Update() {
	dt := float32(0.016) // Time step for updating position

	// Radius and force applied to particles
	radiusOfInfluence := float32(600.0) // Allow interaction from far away
	friction := float32(0.98)
	attractionForce := float32(10.0) // Force to smoothly attract particles toward the mouse

	// Get mouse input for interaction
	mousePos := ss.engine.GetMousePosition()
	mousePressed := ss.engine.IsMouseButtonPressed(glfw.MouseButtonLeft)

	// Convert the mouse screen position to world coordinates
	mouseWorldPos := ss.engine.Camera.ScreenToWorld(mousePos, int(ss.engine.Width), int(ss.engine.Height))

	numWorkers := 16 // Increased number of workers to distribute load more efficiently
	batchSize := len(ss.sandParticles) / numWorkers

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			end := start + batchSize
			if end > len(ss.sandParticles) {
				end = len(ss.sandParticles)
			}
			for i := start; i < end; i++ {
				p := ss.sandParticles[i]
				if !p.active {
					continue
				}

				// Apply gravity to each particle
				gravity := mgl.Vec3{0, -9.81, 0}
				p.velocity = p.velocity.Add(gravity.Mul(dt))

				// Apply attraction force to particles within the radius
				if mousePressed && p.position.Sub(mouseWorldPos).Len() < radiusOfInfluence {
					p.grabbed = true
					// Attract the particle toward the mouse position
					direction := mouseWorldPos.Sub(p.position).Normalize()
					p.velocity = p.velocity.Add(direction.Mul(attractionForce * dt))
				} else if p.grabbed && !mousePressed {
					// Release the particle when the mouse is unclicked
					p.grabbed = false
				}

				// Update position based on velocity
				if !p.grabbed {
					p.position = p.position.Add(p.velocity.Mul(dt))
				}

				// Check for collisions with the floor at y = 0
				if p.position.Y() <= 0 {
					p.position[1] = 0
					p.velocity = p.velocity.Mul(friction)
					p.velocity[1] = 0 // Prevent passing through the floor
				}

				// Update the instance position in the renderer
				ss.sandModel.SetInstancePosition(i, p.position)
			}
		}(w * batchSize)
	}
	wg.Wait()

	// No more grabbing once the mouse is released
	if !mousePressed {
		for i := range ss.sandParticles {
			ss.sandParticles[i].grabbed = false
		}
	}
}

func (ss *SandSimulation) UpdateFixed() {
	// No fixed update required for this example
}
