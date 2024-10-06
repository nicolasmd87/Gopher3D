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
	blackHoles    []*BlackHole
	redParticles  []*Particle
	blueParticles []*Particle
	redModel      *renderer.Model
	blueModel     *renderer.Model
	engine        *engine.Gopher
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

	// Num of instances for each color
	numRedInstances := 90000
	numBlueInstances := 90000

	// Load the red particle model with instancing enabled
	redModel, err := loader.LoadObjectInstance("../resources/obj/Sphere_Low.obj", true, numRedInstances)
	if err != nil {
		panic(err)
	}
	redModel.Scale = mgl.Vec3{5, 5, 5}
	redModel.SetDiffuseColor(255.0, 0.0, 0.0) // Red
	bhb.redModel = redModel
	bhb.engine.AddModel(redModel)

	// Load the blue particle model with instancing enabled
	blueModel, err := loader.LoadObjectInstance("../resources/obj/Sphere_Low.obj", true, numBlueInstances)
	if err != nil {
		panic(err)
	}
	blueModel.Scale = mgl.Vec3{5, 5, 5}
	blueModel.SetDiffuseColor(0.0, 0.0, 255.0) // Blue
	bhb.blueModel = blueModel
	bhb.engine.AddModel(blueModel)

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Initialize red particles
	for i := 0; i < numRedInstances; i++ {
		position := mgl.Vec3{
			rand.Float32()*500 - 100,
			rand.Float32()*500 - 100,
			rand.Float32()*500 - 100,
		}

		velocity := bhb.calculateTangentialVelocity(position, blackHole)

		particle := &Particle{
			position:    position,
			previousPos: position.Sub(velocity), // Initialize previous position for Verlet integration
			color:       "red",
			active:      true,
		}

		bhb.redParticles = append(bhb.redParticles, particle)
		bhb.redModel.SetInstancePosition(i, position)
	}

	// Initialize blue particles
	for i := 0; i < numBlueInstances; i++ {
		position := mgl.Vec3{
			rand.Float32()*500 - 100,
			rand.Float32()*500 - 100,
			rand.Float32()*500 - 100,
		}

		velocity := bhb.calculateTangentialVelocity(position, blackHole)

		particle := &Particle{
			position:    position,
			previousPos: position.Sub(velocity), // Initialize previous position for Verlet integration
			color:       "blue",
			active:      true,
		}

		bhb.blueParticles = append(bhb.blueParticles, particle)
		bhb.blueModel.SetInstancePosition(i, position)
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

func (bhb *BlackHoleBehaviour) Update() {
	// Update red particles
	for i := len(bhb.redParticles) - 1; i >= 0; i-- { // Iterate backwards for safe removal
		p := bhb.redParticles[i]
		if !p.active {
			continue
		}

		for _, bh := range bhb.blackHoles {
			if bh.isWithinEventHorizon(p) {
				// Deactivate the particle and remove its instance
				p.active = false
				bhb.redModel.RemoveModelInstance(i) // Remove instance from the instanced model

				// Remove particle from the list
				bhb.redParticles = append(bhb.redParticles[:i], bhb.redParticles[i+1:]...)
				continue
			}

			bh.ApplyGravity(p)
		}

		// Update particle position using Verlet integration
		newPosition := p.position.Mul(2).Sub(p.previousPos)
		p.previousPos = p.position
		p.position = newPosition

		// Update the instance position in the renderer
		bhb.redModel.SetInstancePosition(i, p.position)
	}

	// Update blue particles
	for i := len(bhb.blueParticles) - 1; i >= 0; i-- { // Iterate backwards for safe removal
		p := bhb.blueParticles[i]
		if !p.active {
			continue
		}

		for _, bh := range bhb.blackHoles {
			if bh.isWithinEventHorizon(p) {
				// Deactivate the particle and remove its instance
				p.active = false
				bhb.blueModel.RemoveModelInstance(i) // Remove instance from the instanced model

				// Remove particle from the list
				bhb.blueParticles = append(bhb.blueParticles[:i], bhb.blueParticles[i+1:]...)
				continue
			}

			bh.ApplyGravity(p)
		}

		// Update particle position using Verlet integration
		newPosition := p.position.Mul(2).Sub(p.previousPos)
		p.previousPos = p.position
		p.position = newPosition

		// Update the instance position in the renderer
		bhb.blueModel.SetInstancePosition(i, p.position)
	}
}

func (bhb *BlackHoleBehaviour) UpdateFixed() {
	// No fixed update required for this example
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
