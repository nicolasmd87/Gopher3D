package main

import (
	behaviour "Gopher3D/internal/Behaviour"
	loader "Gopher3D/internal/Loader"
	"Gopher3D/internal/engine"
	"Gopher3D/internal/renderer"
	"math"
	"math/rand"
	"time"

	mgl "github.com/go-gl/mathgl/mgl32"
)

type Particle struct {
	position mgl.Vec3
	velocity mgl.Vec3
	color    string
	model    *renderer.Model
}

type ParticleBehaviour struct {
	engine    *engine.Gopher
	particles []*Particle
}

func NewParticleBehaviour(engine *engine.Gopher) {
	mb := &ParticleBehaviour{engine: engine}
	behaviour.GlobalBehaviourManager.Add(mb)
}

func main() {
	engine := engine.NewGopher(engine.OPENGL) // or engine.VULKAN
	NewParticleBehaviour(engine)
	engine.Width = 1024
	engine.Height = 768

	// WINDOW POS IN X,Y AND MODEL
	engine.Render(600, 200)
}

func (pb *ParticleBehaviour) Start() {
	pb.engine.Camera.InvertMouse = false
	pb.engine.Camera.Position = mgl.Vec3{0, 50, 500}
	pb.engine.Light = renderer.CreateLight()
	pb.engine.SetFrustumCulling(false)
	pb.engine.SetFaceCulling(true)

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Create a bunch of particles
	numParticles := 1250
	for i := 0; i < numParticles; i++ {
		position := mgl.Vec3{
			rand.Float32()*100 - 10, // X: random between -10 and 10
			rand.Float32()*100 - 10, // Y: random between -10 and 10
			rand.Float32()*100 - 10, // Z: random between -10 and 10
		}
		color := randomColor()
		pb.createParticle(position, color)
	}
	//pb.engine.SetDebugMode(true)
}

func randomColor() string {
	colors := []string{"red", "yellow", "green", "blue"}
	return colors[rand.Intn(len(colors))]
}

func (pb *ParticleBehaviour) createParticle(position mgl.Vec3, color string) {
	m, err := loader.LoadObjectWithPath("../resources/obj/Sphere_Low.obj", true)
	if err != nil {
		panic(err)
	}

	m.Id = rand.Int()
	m.Scale = mgl.Vec3{5, 5, 5}

	switch color {
	case "red":
		m.SetDiffuseColor(255.0, 0.0, 0.0)
	case "yellow":
		m.SetDiffuseColor(255.0, 255.0, 0.0)
	case "green":
		m.SetDiffuseColor(0.0, 255.0, 0.0)
	case "blue":
		m.SetDiffuseColor(0.0, 0.0, 255.0)
	}

	m.Material.Name = "Particle_" + color

	pb.engine.AddModel(m)

	particle := &Particle{
		position: position,
		velocity: mgl.Vec3{0.0, 0.0, 0.0},
		color:    color,
		model:    m,
	}
	pb.particles = append(pb.particles, particle)
}

func (pb *ParticleBehaviour) Update() {
	UpdateParticles(pb)
}

func (pb *ParticleBehaviour) UpdateFixed() {
	//UpdateParticles(pb)
}

func UpdateParticles(pb *ParticleBehaviour) {
	for _, p := range pb.particles {
		pb.applyForces(p)
		p.position = p.position.Add(p.velocity)
		p.model.SetPosition(p.position.X(), p.position.Y(), p.position.Z())
	}
}

// We need a physics engine in Gopher3D to handle this kind of stuff
func (pb *ParticleBehaviour) applyForces(p *Particle) {
	for _, other := range pb.particles {
		if p == other {
			continue
		}

		force := pb.calculateForce(p, other)
		p.velocity = p.velocity.Add(force)
	}
}

// Fast inverse square root function (https://en.wikipedia.org/wiki/Fast_inverse_square_root)
func fastInverseSqrt(x float32) float32 {
	i := math.Float32bits(x)
	i = 0x5f3759df - (i >> 1)
	y := math.Float32frombits(i)
	y = y * (1.5 - (0.5 * x * y * y)) // Perform one iteration of Newton's method
	return y
}

func (pb *ParticleBehaviour) calculateForce(p1, p2 *Particle) mgl.Vec3 {
	direction := p2.position.Sub(p1.position)
	distance := direction.Len()

	// Avoid division by zero
	if distance == 0 {
		return mgl.Vec3{0, 0, 0}
	}

	// Normalize the direction vector
	direction = direction.Normalize()

	// Apply the attraction or repulsion based on color
	var magnitude float32

	switch p1.color {
	case "red":
		if p2.color == "yellow" {
			magnitude = 0.01
		} else if p2.color == "green" {
			magnitude = -0.01
		} else if p2.color == "red" {
			magnitude = 0.01
		}
	case "yellow":
		if p2.color == "red" || p2.color == "green" {
			magnitude = 0.01
		} else if p2.color == "yellow" {
			magnitude = -0.01
		} else if p2.color == "green" {
			magnitude = 0.01
		}
	case "green":
		if p2.color == "yellow" {
			magnitude = 0.01
		} else if p2.color == "red" {
			magnitude = -0.01
		} else if p2.color == "green" {
			magnitude = 0.01
		} else if p2.color == "blue" {
			magnitude = -0.01
		}
	case "blue":
		if p2.color == "red" || p2.color == "green" {
			magnitude = 0.01
		} else if p2.color == "yellow" {
			magnitude = -0.01
		} else if p2.color == "blue" {
			magnitude = 0.01
		} else if p2.color == "green" {
			magnitude = 0.01
		}
	}

	// Calculate the force using the fast inverse square root method
	invSqrtDistance := fastInverseSqrt(distance)
	force := direction.Mul(magnitude * invSqrtDistance * invSqrtDistance)

	return force
}
