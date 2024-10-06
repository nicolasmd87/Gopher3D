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
	position    mgl.Vec3
	previousPos mgl.Vec3
	color       string
	model       *renderer.Model
	active      bool // To check if the particle is still active
}

type BlackHole struct {
	position mgl.Vec3
	mass     float32
	radius   float32 // Radius of the black hole's event horizon
}

type BlackHoleBehaviour struct {
	blackHoles []*BlackHole
	particles  []*Particle
	engine     *engine.Gopher
}

func NewBlackHoleBehaviour(engine *engine.Gopher) {
	bhb := &BlackHoleBehaviour{engine: engine}
	behaviour.GlobalBehaviourManager.Add(bhb)
}

func main() {
	engine := engine.NewGopher(engine.OPENGL) // or engine.VULKAN
	NewBlackHoleBehaviour(engine)

	engine.Width = 1204
	engine.Height = 768

	engine.Render(200, 100)
}

func (bhb *BlackHoleBehaviour) Start() {
	startCPUProfile()
	defer stopCPUProfile()
	bhb.engine.Camera.InvertMouse = false
	bhb.engine.Camera.Position = mgl.Vec3{0, 50, 1000}
	bhb.engine.Camera.Speed = 900
	bhb.engine.Light = renderer.CreateLight()
	bhb.engine.Light.Type = renderer.STATIC_LIGHT

	// Create and add a black hole to the scene
	bhPosition := mgl.Vec3{0, 0, 0} // Position of the black hole at the origin
	bhMass := float32(10000)        // Mass of the black hole
	bhRadius := float32(50)         // Radius of the black hole's event horizon
	blackHole := &BlackHole{position: bhPosition, mass: bhMass, radius: bhRadius}
	bhb.blackHoles = append(bhb.blackHoles, blackHole)

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Create a large number of particles with reduced initial velocities
	//	numParticles := 9000
	//	for i := 0; i < numParticles; i++ {
	position := mgl.Vec3{
		rand.Float32()*200 - 100, // X: random between -100 and 100
		rand.Float32()*200 - 100, // Y: random between -100 and 100
		rand.Float32()*200 - 100, // Z: random between -100 and 100
	}

	// Calculate a tangential velocity based on the position to induce orbital motion
	velocity := bhb.calculateTangentialVelocity(position, blackHole)

	color := randomColor()
	particle := bhb.createParticle(position, velocity, color)
	bhb.particles = append(bhb.particles, particle)
	//}
}

func (bhb *BlackHoleBehaviour) calculateTangentialVelocity(position mgl.Vec3, blackHole *BlackHole) mgl.Vec3 {
	// Calculate the direction vector from the black hole to the particle
	direction := position.Sub(blackHole.position).Normalize()

	// Calculate a perpendicular vector to the direction vector for tangential velocity
	tangential := mgl.Vec3{-direction.Y(), direction.X(), 0}.Normalize()

	// Set the magnitude of the tangential velocity based on the distance to the black hole
	distance := position.Sub(blackHole.position).Len()
	speed := float32(math.Sqrt(float64(blackHole.mass) / float64(distance)))

	// Significantly reduce the speed to prevent particles from escaping
	return tangential.Mul(speed * 0.01)
}

func (bhb *BlackHoleBehaviour) createParticle(position, velocity mgl.Vec3, color string) *Particle {
	// Load the model for the particle with instancing enabled
	m, err := loader.LoadObjectInstance("../resources/obj/Sphere_Low.obj", true, 9000)
	if err != nil {
		panic(err)
	}

	m.Id = rand.Int()
	m.Scale = mgl.Vec3{5, 5, 5} // Set the scale for the particle

	// Set the color of the particle
	switch color {
	case "red":
		m.SetDiffuseColor(255.0, 0.0, 0.0)
	case "blue":
		m.SetDiffuseColor(0.0, 0.0, 255.0)
	}

	m.Material.Name = "Particle_" + color

	// Initialize the particle's position in the renderer
	m.SetPosition(position.X(), position.Y(), position.Z())

	// Add the model to the engine
	bhb.engine.AddModel(m)

	return &Particle{
		position:    position,
		previousPos: position.Sub(velocity), // Initialize previous position for Verlet integration
		color:       color,
		model:       m,
		active:      true, // Mark the particle as active
	}
}

func randomColor() string {
	colors := []string{"red", "blue"}
	return colors[rand.Intn(len(colors))]
}

func (bhb *BlackHoleBehaviour) Update() {
	for _, p := range bhb.particles {
		if !p.active {
			continue // Skip inactive particles
		}
		for _, bh := range bhb.blackHoles {
			if bh.isWithinEventHorizon(p) {
				// Remove the particle model from the engine to simulate disappearance
				bhb.engine.RemoveModel(p.model)
				p.active = false // Deactivate the particle
				continue
			}
			bh.ApplyGravity(p)
		}
		// Update the particle's position using Verlet integration
		newPosition := p.position.Mul(2).Sub(p.previousPos)
		p.previousPos = p.position
		p.position = newPosition

		p.model.SetPosition(p.position.X(), p.position.Y(), p.position.Z())
	}
}

func (bh *BlackHole) isWithinEventHorizon(p *Particle) bool {
	// Check if the particle is within the black hole's event horizon
	distance := p.position.Sub(bh.position).Len()
	return distance < bh.radius
}

func (bhb *BlackHoleBehaviour) UpdateFixed() {
	// Not used in this example
}

func (bh *BlackHole) ApplyGravity(p *Particle) {
	direction := bh.position.Sub(p.position)
	distance := direction.Len()

	if distance == 0 {
		return
	}

	direction = direction.Normalize()

	// Calculate the gravitational force based on the black hole's mass
	gravity := (bh.mass * 0.0005) / (distance * distance) // Reduce the gravity to prevent excessive forces

	force := direction.Mul(gravity)

	// Further reduce the force to prevent particles from spiraling outwards too quickly
	maxForce := float32(2.0)
	if force.Len() > maxForce {
		force = force.Normalize().Mul(maxForce)
	}

	// Update the particle's position using Verlet integration
	p.position = p.position.Add(force)
}
