// camera.go
package renderer

import (
	"math"

	"github.com/xlab/linmath"

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

type Plane struct {
	Normal   mgl32.Vec3
	Distance float32
}

type Frustum struct {
	Planes [6]Plane
}

func NewCamera(height int32, width int32) *Camera {
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
	// TODO: Ideally, the aspect ratio should be calculated dynamically based on the window dimensions
	projection := mgl32.Perspective(mgl32.DegToRad(camera.fov), float32(height)/float32(width), 0.1, 1000.0)
	camera.projection = projection
	return &camera
}

func (c *Camera) UpdateProjection(fov, aspectRatio, near, far float32) {
	c.projection = mgl32.Perspective(mgl32.DegToRad(fov), aspectRatio, near, far)
}

func (c *Camera) GetViewProjection() mgl32.Mat4 {
	view := mgl32.LookAtV(c.position, c.position.Add(c.front), c.up)
	return c.projection.Mul4(view)
}

// TODO: THIS IS JUST WHILE I TEST VULKAN INTEGRATION, WE SHOULD NOT BE CONVERTING ANYTHING
//
//	WE NEED TO WORK ON A LINMATH CAMERA IMPLEMENTATION FOR VULKAN
//
// Assume linmath.Mat4x4 is a struct with a flat array of 16 float32s
func convertMGL32Mat4ToLinMathMat4x4(m mgl32.Mat4) linmath.Mat4x4 {
	var l linmath.Mat4x4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			l[i][j] = m[i*4+j]
		}
	}
	return l
}

func (c *Camera) GetViewProjectionVulkan() linmath.Mat4x4 {
	return convertMGL32Mat4ToLinMathMat4x4(c.GetViewProjection())
}

func (c *Camera) GetViewMatrixVulkan() linmath.Mat4x4 {
	view := mgl32.LookAtV(c.position, c.position.Add(c.front), c.up)
	return convertMGL32Mat4ToLinMathMat4x4(view)
}

func (c *Camera) ProcessKeyboard(window *glfw.Window, deltaTime float32) {
	// Compute the right vector
	c.right = c.front.Cross(c.worldUp).Normalize()
	baseVelocity := c.speed * deltaTime

	// If Shift is pressed, multiply the velocity by a factor (e.g., 2.5)
	if window.GetKey(glfw.KeyLeftShift) == glfw.Press || window.GetKey(glfw.KeyRightShift) == glfw.Press {
		baseVelocity *= 2.5
	}

	if window.GetKey(glfw.KeyW) == glfw.Press {
		c.position = c.position.Add(c.front.Mul(baseVelocity))
	}
	if window.GetKey(glfw.KeyS) == glfw.Press {
		c.position = c.position.Sub(c.front.Mul(baseVelocity))
	}
	if window.GetKey(glfw.KeyA) == glfw.Press {
		c.position = c.position.Sub(c.right.Mul(baseVelocity))
	}
	if window.GetKey(glfw.KeyD) == glfw.Press {
		c.position = c.position.Add(c.right.Mul(baseVelocity))
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

func (c *Camera) CalculateFrustum() Frustum {
	var frustum Frustum
	vp := c.GetViewProjection()

	// Left Plane
	frustum.Planes[0] = Plane{
		Normal:   mgl32.Vec3{vp[3] + vp[0], vp[7] + vp[4], vp[11] + vp[8]},
		Distance: vp[15] + vp[12],
	}

	// Right Plane
	frustum.Planes[1] = Plane{
		Normal:   mgl32.Vec3{vp[3] - vp[0], vp[7] - vp[4], vp[11] - vp[8]},
		Distance: vp[15] - vp[12],
	}

	// Bottom Plane
	frustum.Planes[2] = Plane{
		Normal:   mgl32.Vec3{vp[3] + vp[1], vp[7] + vp[5], vp[11] + vp[9]},
		Distance: vp[15] + vp[13],
	}

	// Top Plane
	frustum.Planes[3] = Plane{
		Normal:   mgl32.Vec3{vp[3] - vp[1], vp[7] - vp[5], vp[11] - vp[9]},
		Distance: vp[15] - vp[13],
	}

	// Near Plane
	frustum.Planes[4] = Plane{
		Normal:   mgl32.Vec3{vp[3] + vp[2], vp[7] + vp[6], vp[11] + vp[10]},
		Distance: vp[15] + vp[14],
	}

	// Far Plane
	frustum.Planes[5] = Plane{
		Normal:   mgl32.Vec3{vp[3] - vp[2], vp[7] - vp[6], vp[11] - vp[10]},
		Distance: vp[15] - vp[14],
	}

	// Normalize the planes
	for i := 0; i < 6; i++ {
		length := frustum.Planes[i].Normal.Len()
		frustum.Planes[i].Normal = frustum.Planes[i].Normal.Mul(1.0 / length)
		frustum.Planes[i].Distance /= length
	}

	return frustum
}

func (p *Plane) DistanceToPoint(point mgl32.Vec3) float32 {
	return p.Normal.Dot(point) + p.Distance
}

func (f *Frustum) IntersectsSphere(center mgl32.Vec3, radius float32) bool {
	// Check if the sphere intersects with the frustum
	for _, plane := range f.Planes {
		if plane.DistanceToPoint(center) < -radius {
			return false // Sphere is outside the frustum
		}
	}
	return true
}
