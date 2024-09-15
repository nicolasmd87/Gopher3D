package main

import (
	behaviour "Gopher3D/internal/Behaviour"
	loader "Gopher3D/internal/Loader"
	"Gopher3D/internal/engine"
	"Gopher3D/internal/renderer"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"

	"github.com/g3n/engine/experimental/physics"
	"github.com/g3n/engine/math32"
	mgl "github.com/go-gl/mathgl/mgl32"
)

type Particle struct {
	position    mgl.Vec3
	previousPos mgl.Vec3
	velocity    mgl.Vec3
	color       string
	model       *renderer.Model
	active      bool // To check if the particle is still active
}

type BlackHole struct {
	position   mgl.Vec3
	mass       float32
	radius     float32 // Radius of the black hole's event horizon
	forceField *physics.AttractorForceField
}

type BlackHoleBehaviour struct {
	blackHoles []*BlackHole
	particles  []*Particle
	engine     *engine.Gopher
	simulation *physics.Simulation
}

func NewBlackHoleBehaviour(engine *engine.Gopher) {
	bhb := &BlackHoleBehaviour{engine: engine}
	behaviour.GlobalBehaviourManager.Add(bhb)
}

func main() {

	// Profiling
	// ========================================
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	// ========================================

	engine := engine.NewGopher(engine.OPENGL) // or engine.VULKAN
	NewBlackHoleBehaviour(engine)

	engine.Width = 1204
	engine.Height = 768

	engine.Render(200, 100)
}

func (bhb *BlackHoleBehaviour) Start() {
	bhb.engine.Camera.InvertMouse = false
	bhb.engine.Camera.Position = mgl.Vec3{0, 50, 1000}
	bhb.engine.Camera.Speed = 900
	bhb.engine.Light = renderer.CreateLight()
	bhb.engine.Light.Type = renderer.STATIC_LIGHT
	bhb.engine.SetFaceCulling(true)

	// Create and add a black hole to the scene
	bhPosition := mgl.Vec3{0, 0, 0} // Position of the black hole at the origin
	bhMass := float32(10000)        // Mass of the black hole
	bhRadius := float32(50)         // Radius of the black hole's event horizon

	// Create the AttractorForceField
	attractor := physics.NewAttractorForceField(&math32.Vector3{bhPosition.X(), bhPosition.Y(), bhPosition.Z()}, bhMass)
	blackHole := &BlackHole{position: bhPosition, mass: bhMass, radius: bhRadius, forceField: attractor}

	bhb.blackHoles = append(bhb.blackHoles, blackHole)

	// Initialize the physics simulation
	bhb.simulation = physics.NewSimulation(nil)
	bhb.simulation.AddForceField(attractor) // Add the force field to the simulation

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Create a large number of particles
	numParticles := 16000
	for i := 0; i < numParticles; i++ {
		position := mgl.Vec3{
			rand.Float32()*200 - 100, // X: random between -100 and 100
			rand.Float32()*200 - 100, // Y: random between -100 and 100
			rand.Float32()*200 - 100, // Z: random between -100 and 100
		}

		velocity := bhb.calculateInitialVelocity(position, blackHole)

		color := randomColor()
		particle := bhb.createParticle(position, velocity, color)
		bhb.particles = append(bhb.particles, particle)
	}
}

func (bhb *BlackHoleBehaviour) calculateInitialVelocity(position mgl.Vec3, blackHole *BlackHole) mgl.Vec3 {
	// Calculate the direction vector from the black hole to the particle
	direction := position.Sub(blackHole.position).Normalize()

	// Calculate a perpendicular vector to the direction vector for tangential velocity
	tangential := mgl.Vec3{-direction.Y(), direction.X(), 0}.Normalize()

	// Set the magnitude of the tangential velocity based on the distance to the black hole
	distance := position.Sub(blackHole.position).Len()

	// Calculate escape velocity to induce orbiting behavior
	escapeVelocity := float32(math.Sqrt(float64(2*blackHole.mass) / float64(distance)))

	// Multiply tangential velocity by the escape velocity factor
	return tangential.Mul(escapeVelocity * 0.75) // Lower than escape velocity to ensure orbit
}

func (bhb *BlackHoleBehaviour) createParticle(position, velocity mgl.Vec3, color string) *Particle {
	m, err := loader.LoadObjectWithPath("../resources/obj/Sphere_Low.obj", true)
	if err != nil {
		panic(err)
	}

	m.Id = rand.Int()
	m.Scale = mgl.Vec3{5, 5, 5}

	switch color {
	case "red":
		m.SetDiffuseColor(255.0, 0.0, 0.0)
	case "blue":
		m.SetDiffuseColor(0.0, 0.0, 255.0)
	}

	m.Material.Name = "Particle_" + color
	bhb.engine.AddModel(m)

	return &Particle{
		position:    position,
		previousPos: position.Sub(velocity),
		velocity:    velocity, // Initialize velocity here
		color:       color,
		model:       m,
		active:      true,
	}
}

func randomColor() string {
	colors := []string{"red", "blue"}
	return colors[rand.Intn(len(colors))]
}

func (bhb *BlackHoleBehaviour) Update() {
	for _, p := range bhb.particles {
		if !p.active {
			continue
		}
		for _, bh := range bhb.blackHoles {
			if bh.isWithinEventHorizon(p) {
				bhb.engine.RemoveModel(p.model)
				p.active = false
				continue
			}
			// Apply force to particle based on force field
			force := bh.forceField.ForceAt(&math32.Vector3{p.position.X(), p.position.Y(), p.position.Z()})
			gravity := mgl.Vec3{force.X, force.Y, force.Z}

			// Adjust the velocity based on gravity but add damping for realism
			p.velocity = p.velocity.Add(gravity.Mul(0.1)) // Dampen the force for realism

			// Update particle position based on velocity
			p.position = p.position.Add(p.velocity.Mul(1.0 / 60.0))
		}
		p.model.SetPosition(p.position.X(), p.position.Y(), p.position.Z())
	}
}

func (bh *BlackHole) isWithinEventHorizon(p *Particle) bool {
	distance := p.position.Sub(bh.position).Len()
	return distance < bh.radius
}

// Not used in this example
func (bhb *BlackHoleBehaviour) UpdateFixed() {}
