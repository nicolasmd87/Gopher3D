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
	lastMousePos  mgl.Vec2                     // Store the last mouse position to detect movement
	mousePressed  bool                         // Track mouse press state
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
	ss.engine.Light.Intensity = 0.03                    // Tone down the light intensity even more
	ss.engine.Light.Position = mgl.Vec3{0, 2000, -1200} // Move the light further away

	// Load the sand particle model with instancing enabled
	instances := 200000 // More particles for a denser effect
	sandModel, err := loader.LoadObjectInstance("../resources/obj/Sphere_Low.obj", true, instances)
	if err != nil {
		panic(err)
	}
	sandModel.Scale = mgl.Vec3{0.1, 0.1, 0.1} // Smaller particles
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
	ss.lastMousePos = ss.engine.GetMousePosition() // Initialize mouse position
	ss.mousePressed = false
}

func (ss *SandSimulation) Update() {
	dt := float32(0.016) // Time step for updating position

	// Force applied to particles
	attractionForce := float32(15.0) // Balanced force for smoother movement

	// Get mouse input for interaction
	mousePos := ss.engine.GetMousePosition()
	mousePressed := ss.engine.IsMouseButtonPressed(glfw.MouseButtonLeft)
	ss.mousePressed = mousePressed

	// Convert the mouse screen position to world coordinates
	mouseWorldPos := ss.engine.Camera.ScreenToWorld(mousePos, int(ss.engine.Width), int(ss.engine.Height))

	// Adjust attraction force when mouse is clicked
	if mousePressed {
		ss.ApplyForcesToParticles(mouseWorldPos, attractionForce, dt)
	}

	ss.UpdateParticles(dt, 0.95) // friction set to 0.95
}

func (ss *SandSimulation) ApplyForcesToParticles(mouseWorldPos mgl.Vec3, attractionForce float32, dt float32) {
	radiusOfInfluence := float32(300.0)

	// Split the particle processing across workers
	numWorkers := 16
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

				// Apply force only within radius
				if p.position.Sub(mouseWorldPos).Len() < radiusOfInfluence {
					p.grabbed = true
					// Apply smoother force to move toward the mouse's world position
					direction := mouseWorldPos.Sub(p.position).Normalize()
					forceToMouse := direction.Mul(attractionForce * dt)
					p.velocity = p.velocity.Add(forceToMouse)
				}
			}
		}(w * batchSize)
	}
	wg.Wait()
}

func (ss *SandSimulation) UpdateParticles(dt float32, friction float32) {
	numWorkers := 16
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

				// Apply gravity to each particle to ensure they fall
				gravity := mgl.Vec3{0, -9.81, 0}
				p.velocity = p.velocity.Add(gravity.Mul(dt))

				// Update position based on velocity
				p.position = p.position.Add(p.velocity.Mul(dt))

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
}

func (ss *SandSimulation) UpdateFixed() {
	// No fixed update required for this example
}
