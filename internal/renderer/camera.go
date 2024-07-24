// camera.go
package renderer

import (
	"math"

	"github.com/xlab/linmath"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Camera struct {
	Position     mgl32.Vec3
	Front        mgl32.Vec3
	Up           mgl32.Vec3
	Right        mgl32.Vec3
	WorldUp      mgl32.Vec3
	Pitch        float32
	Projection   mgl32.Mat4
	Yaw          float32
	Speed        float32
	Sensitivity  float32
	Fov          float32
	Near         float32
	Far          float32
	AspectRatio  float32
	LastX, LastY float32
	InvertMouse  bool
	firstMouse   bool
}

type Plane struct {
	Normal   mgl32.Vec3
	Distance float32
}

type Frustum struct {
	Planes [6]Plane
}

func NewDefaultCamera(height int32, width int32) *Camera {
	camera := Camera{
		Position:    mgl32.Vec3{1, 0, 100},
		Front:       mgl32.Vec3{0, 0, -1},
		Up:          mgl32.Vec3{0, 1, 0},
		WorldUp:     mgl32.Vec3{0, 1, 0},
		Pitch:       0.0,
		Yaw:         -90.0,
		Speed:       70,
		Sensitivity: 0.1,
		Fov:         45.0,
		Near:        0.1,
		Far:         1000.0,
		LastX:       float32(width) / 2,
		LastY:       float32(height) / 2,
		AspectRatio: float32(height) / float32(width),
		firstMouse:  true,
		InvertMouse: true,
	}
	camera.updateCameraVectors()
	camera.UpdateProjection()
	return &camera
}

func (c *Camera) UpdateProjection() {
	c.Projection = mgl32.Perspective(mgl32.DegToRad(c.Fov), c.AspectRatio, c.Near, c.Far)
}

func (c *Camera) GetViewProjection() mgl32.Mat4 {
	view := mgl32.LookAtV(c.Position, c.Position.Add(c.Front), c.Up)
	return c.Projection.Mul4(view)
}

// TODO: THIS IS JUST WHILE I TEST VULKAN INTEGRATION, WE SHOULD NOT BE CONVERTING ANYTHING
//
//	WE NEED TO WORK ON A LINMATH CAMERA IMPLEMENTATION FOR VULKAN
//
// Assume linmath.Mat4x4 is a struct with a flat array of 16 float32s
func convertMGL32Mat4ToLinMathMat4x4(m mgl32.Mat4) linmath.Mat4x4 {
	var viewProjection linmath.Mat4x4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			viewProjection[i][j] = m[i*4+j]
		}
	}
	return viewProjection
}

func (c *Camera) GetViewProjectionVulkan() linmath.Mat4x4 {
	return convertMGL32Mat4ToLinMathMat4x4(c.GetViewProjection())
}

func (c *Camera) GetViewMatrixVulkan() linmath.Mat4x4 {
	view := mgl32.LookAtV(c.Position, c.Position.Add(c.Front), c.Up)
	return convertMGL32Mat4ToLinMathMat4x4(view)
}

func (c *Camera) ProcessKeyboard(window *glfw.Window, deltaTime float32) {
	// Compute the right vector
	c.Right = c.Front.Cross(c.WorldUp).Normalize()
	baseVelocity := c.Speed * deltaTime

	// If Shift is pressed, multiply the velocity by a factor (e.g., 2.5)
	if window.GetKey(glfw.KeyLeftShift) == glfw.Press || window.GetKey(glfw.KeyRightShift) == glfw.Press {
		baseVelocity *= 2.5
	}

	if window.GetKey(glfw.KeyW) == glfw.Press {
		c.Position = c.Position.Add(c.Front.Mul(baseVelocity))
	}
	if window.GetKey(glfw.KeyS) == glfw.Press {
		c.Position = c.Position.Sub(c.Front.Mul(baseVelocity))
	}
	if window.GetKey(glfw.KeyA) == glfw.Press {
		c.Position = c.Position.Sub(c.Right.Mul(baseVelocity))
	}
	if window.GetKey(glfw.KeyD) == glfw.Press {
		c.Position = c.Position.Add(c.Right.Mul(baseVelocity))
	}
}

func (c *Camera) ProcessMouseMovement(xoffset, yoffset float32, constrainPitch bool) {
	xoffset *= c.Sensitivity
	yoffset *= c.Sensitivity

	c.Yaw += xoffset

	if c.InvertMouse {
		c.Pitch -= yoffset
	} else {
		c.Pitch += yoffset
	}
	if constrainPitch {
		c.Pitch = mgl32.Clamp(c.Pitch, -89.0, 89.0) // Prevent extreme pitch values
	}
	c.updateCameraVectors()
}

func (c *Camera) LookAt(target mgl32.Vec3) {
	direction := c.Position.Sub(target).Normalize()
	c.Yaw = float32(math.Atan2(float64(direction.Z()), float64(direction.X())))
	c.Pitch = float32(math.Atan2(float64(direction.Y()), math.Sqrt(float64(direction.X()*direction.X()+direction.Z()*direction.Z()))))
	c.updateCameraVectors()
}

func (c *Camera) updateCameraVectors() {
	yawRad := mgl32.DegToRad(c.Yaw)
	pitchRad := mgl32.DegToRad(c.Pitch)

	front := mgl32.Vec3{
		float32(math.Cos(float64(yawRad)) * math.Cos(float64(pitchRad))),
		float32(math.Sin(float64(pitchRad))),
		float32(math.Sin(float64(yawRad)) * math.Cos(float64(pitchRad))),
	}

	c.Front = front.Normalize()
	c.Right = c.WorldUp.Cross(c.Front).Normalize()
	c.Up = c.Front.Cross(c.Right).Normalize()
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
