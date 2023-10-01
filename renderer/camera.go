// camera.go
package renderer

import (
	"fmt"
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Camera struct {
	position    mgl32.Vec3
	front       mgl32.Vec3
	up          mgl32.Vec3
	right       mgl32.Vec3
	worldUp     mgl32.Vec3
	pitch       float32
	projection  mgl32.Mat4
	yaw         float32
	speed       float32
	sensitivity float32
	fov         float32
}

func NewCamera() Camera {
	camera := Camera{
		position:    mgl32.Vec3{1, 0, 100},
		front:       mgl32.Vec3{0, 0, -10},
		up:          mgl32.Vec3{0, -1, 0},
		pitch:       0.0,   // Looking straight ahead
		yaw:         -90.0, // Initial yaw, facing negative Z
		speed:       2.5,   // Example speed value, adjust as necessary
		sensitivity: 0.1,   // Example sensitivity, adjust as necessary
		fov:         45.0,  // Example field of view, in degrees
	}
	//camera.updateCameraVectors()
	projection := mgl32.Perspective(mgl32.DegToRad(camera.fov), float32(800)/float32(600), 0.1, 100.0)
	camera.projection = projection
	return camera
}

func (c *Camera) UpdateProjection(fov, aspectRatio, near, far float32) {
	c.projection = mgl32.Perspective(mgl32.DegToRad(fov), aspectRatio, near, far)
}

func (c *Camera) GetViewProjection() mgl32.Mat4 {
	view := mgl32.LookAtV(c.position, c.position.Add(c.front), c.up)
	return c.projection.Mul4(view)
}

func (c *Camera) ProcessKeyboard(window *glfw.Window, deltaTime float32) {
	velocity := c.speed * deltaTime
	if window.GetKey(glfw.KeyW) == glfw.Press {
		fmt.Println("W pressed")
		c.position = c.position.Add(c.front.Mul(velocity))
	}
	if window.GetKey(glfw.KeyS) == glfw.Press {
		fmt.Println("S pressed")
		c.position = c.position.Sub(c.front.Mul(velocity))
	}
	if window.GetKey(glfw.KeyA) == glfw.Press {
		fmt.Println("A pressed")
		c.position = c.position.Sub(c.right.Mul(velocity))
	}
	if window.GetKey(glfw.KeyD) == glfw.Press {
		fmt.Println("D pressed")
		c.position = c.position.Add(c.right.Mul(velocity))
	}
}

func (c *Camera) ProcessMouseMovement(xoffset, yoffset float32, constrainPitch bool) {
	xoffset *= c.sensitivity
	yoffset *= c.sensitivity
	c.yaw += xoffset
	c.pitch += yoffset
	if constrainPitch {
		if c.pitch > 89.0 {
			c.pitch = 89.0
		}
		if c.pitch < -89.0 {
			c.pitch = -89.0
		}
	}
	c.updateCameraVectors()
}

func (c *Camera) updateCameraVectors() {
	yawRad := mgl32.DegToRad(c.yaw)
	pitchRad := mgl32.DegToRad(c.pitch)

	front := mgl32.Vec3{
		float32(math.Cos(float64(yawRad)) * math.Cos(float64(pitchRad))),
		float32(math.Sin(float64(pitchRad))),
		float32(math.Sin(float64(yawRad)) * math.Cos(float64(pitchRad))),
	}
	c.front = front.Normalize()
	c.right = c.front.Cross(c.worldUp).Normalize()
	c.up = c.right.Cross(c.front).Normalize()
}
