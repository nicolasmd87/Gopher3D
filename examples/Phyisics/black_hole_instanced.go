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
	model      *renderer.Model
	engine     *engine.Gopher
}

func NewBlackHoleBehaviour(engine *engine.Gopher) {
	bhb := &BlackHoleBehaviour{engine: engine}
	behaviour.GlobalBehaviourManager.Add(bhb)
}

func main() {
	engine := engine.NewGopher(engine.OPENGL)
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
	bhPosition := mgl.Vec3{0, 0, 0}
	bhMass := float32(10000)
	bhRadius := float32(50)
	blackHole := &BlackHole{position: bhPosition, mass: bhMass, radius: bhRadius}
	bhb.blackHoles = append(bhb.blackHoles, blackHole)

	// Load the particle model with instancing enabled
	model, err := loader.LoadObjectInstance("../resources/obj/Sphere_Low.obj", true, 1000000) // 1000 instances
	if err != nil {
		panic(err)
	}
	model.Scale = mgl.Vec3{5, 5, 5}
	bhb.model = model
	bhb.engine.AddModel(model)

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Initialize particles
	for i := 0; i < 1000000; i++ {
		position := mgl.Vec3{
			rand.Float32()*200 - 100,
			rand.Float32()*200 - 100,
			rand.Float32()*200 - 100,
		}

		velocity := bhb.calculateTangentialVelocity(position, blackHole)
		color := randomColor()

		particle := &Particle{
			position:    position,
			previousPos: position.Sub(velocity), // Initialize previous position for Verlet integration
			color:       color,
			active:      true,
		}

		bhb.particles = append(bhb.particles, particle)

		// Update the instance position in the model (no need to worry about instancing)
		bhb.model.SetInstancePosition(i, position)

		// Set the color for the particle
		switch color {
		case "red":
			bhb.model.SetDiffuseColor(255.0, 0.0, 0.0)
		case "blue":
			bhb.model.SetDiffuseColor(0.0, 0.0, 255.0)
		}
	}
}

func (bhb *BlackHoleBehaviour) calculateTangentialVelocity(position mgl.Vec3, blackHole *BlackHole) mgl.Vec3 {
	// Calculate the direction vector from the black hole to the particle
	direction := position.Sub(blackHole.position).Normalize()

	// Calculate a perpendicular vector for tangential velocity
	tangential := mgl.Vec3{-direction.Y(), direction.X(), 0}.Normalize()

	// Set the magnitude of the tangential velocity based on the distance
	distance := position.Sub(blackHole.position).Len()
	speed := float32(math.Sqrt(float64(blackHole.mass) / float64(distance)))

	// Reduce the speed to prevent particles from escaping
	return tangential.Mul(speed * 0.01)
}

func randomColor() string {
	colors := []string{"red", "blue"}
	return colors[rand.Intn(len(colors))]
}

func (bhb *BlackHoleBehaviour) Update() {
	for i, p := range bhb.particles {
		if !p.active {
			continue
		}

		for _, bh := range bhb.blackHoles {
			if bh.isWithinEventHorizon(p) {
				p.active = false // Deactivate particle
				continue
			}

			bh.ApplyGravity(p)
		}

		// Update the particle position using Verlet integration
		newPosition := p.position.Mul(2).Sub(p.previousPos)
		p.previousPos = p.position
		p.position = newPosition

		// Update the instance position in the renderer
		bhb.model.SetInstancePosition(i, p.position)
	}
}

func (bhb *BlackHoleBehaviour) UpdateFixed() {
}

func (bh *BlackHole) isWithinEventHorizon(p *Particle) bool {
	distance := p.position.Sub(bh.position).Len()
	return distance < bh.radius
}

func (bh *BlackHole) ApplyGravity(p *Particle) {
	direction := bh.position.Sub(p.position)
	distance := direction.Len()

	if distance == 0 {
		return
	}

	direction = direction.Normalize()

	// Calculate the gravitational force based on the black hole's mass
	gravity := (bh.mass * 0.0005) / (distance * distance)
	force := direction.Mul(gravity)

	// Limit the maximum force to avoid excessive acceleration
	maxForce := float32(2.0)
	if force.Len() > maxForce {
		force = force.Normalize().Mul(maxForce)
	}

	// Update the particle's position
	p.position = p.position.Add(force)
}
