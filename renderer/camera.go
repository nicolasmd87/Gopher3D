// camera.go
package renderer

import (
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Camera struct {
	position     mgl32.Vec3
	front        mgl32.Vec3
	up           mgl32.Vec3
	right        mgl32.Vec3
	worldUp      mgl32.Vec3
	pitch        float32
	projection   mgl32.Mat4
	yaw          float32
	speed        float32
	sensitivity  float32
	fov          float32
	lastX, lastY float32
	firstMouse   bool
}

func NewCamera(height int32, width int32) Camera {
	camera := Camera{
		position:    mgl32.Vec3{1, 0, 200},
		front:       mgl32.Vec3{0, 0, -1},
		up:          mgl32.Vec3{0, 1, 0}, // Changed to the conventional up vector
		worldUp:     mgl32.Vec3{0, 1, 0},
		pitch:       0.0,
		yaw:         -90.0,
		speed:       70,
		sensitivity: 0.1,
		fov:         45.0,
		lastX:       float32(width) / 2,
		lastY:       float32(height) / 2,
		firstMouse:  true,
	}
	camera.updateCameraVectors()
	// Ideally, the aspect ratio should be calculated dynamically based on the window dimensions
	projection := mgl32.Perspective(mgl32.DegToRad(camera.fov), float32(height)/float32(width), 0.1, 1000.0)
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
	// Compute the right vector
	c.right = c.front.Cross(c.worldUp).Normalize()
<<<<<<< HEAD

	baseVelocity := c.speed * deltaTime

	// If Shift is pressed, multiply the velocity by a factor (e.g., 2.5)
	if window.GetKey(glfw.KeyLeftShift) == glfw.Press || window.GetKey(glfw.KeyRightShift) == glfw.Press {
		baseVelocity *= 2.5
	}

=======

	velocity := c.speed * deltaTime

	// If Shift is pressed, multiply the velocity by a factor (e.g., 2.5)
	if window.GetKey(glfw.KeyLeftShift) == glfw.Press || window.GetKey(glfw.KeyRightShift) == glfw.Press {
		velocity *= 2.5
	}

>>>>>>> main
	if window.GetKey(glfw.KeyW) == glfw.Press {
		c.position = c.position.Add(c.front.Mul(baseVelocity))
	}
	if window.GetKey(glfw.KeyS) == glfw.Press {
		c.position = c.position.Sub(c.front.Mul(baseVelocity))
	}
	if window.GetKey(glfw.KeyA) == glfw.Press {
<<<<<<< HEAD
		c.position = c.position.Sub(c.right.Mul(baseVelocity))
	}
	if window.GetKey(glfw.KeyD) == glfw.Press {
		c.position = c.position.Add(c.right.Mul(baseVelocity))
=======
		c.position = c.position.Sub(c.right.Mul(velocity))
	}
	if window.GetKey(glfw.KeyD) == glfw.Press {
		c.position = c.position.Add(c.right.Mul(velocity))
>>>>>>> main
	}
}

func (c *Camera) ProcessMouseMovement(xoffset, yoffset float32, constrainPitch bool) {
	xoffset *= c.sensitivity
	yoffset *= c.sensitivity

	c.yaw += xoffset
	c.pitch += yoffset // subtract yoffset to invert vertical mouse movement

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

func (c *Camera) LookAt(target mgl32.Vec3) {
	direction := c.position.Sub(target).Normalize()
	c.yaw = float32(math.Atan2(float64(direction.Z()), float64(direction.X())))
	c.pitch = float32(math.Atan2(float64(direction.Y()), math.Sqrt(float64(direction.X()*direction.X()+direction.Z()*direction.Z()))))
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
