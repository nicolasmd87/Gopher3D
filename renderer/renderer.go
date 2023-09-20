package renderer

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	triangleVertices = []float32{
		0.0, 0.5, 0.0,
		-0.5, -0.5, 0.0,
		0.5, -0.5, 0.0,
	}
	vertexArray   uint32
	vertexBuffer  uint32
	shaderProgram uint32
	model         mgl32.Mat4
	uniformLoc    int32
)

func Init() {
	if err := gl.Init(); err != nil {
		fmt.Println("OpenGL initialization failed:", err)
		return
	}
	initOpenGL()
}

func initOpenGL() {
	gl.GenVertexArrays(1, &vertexArray)
	gl.BindVertexArray(vertexArray)

	gl.GenBuffers(1, &vertexBuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, len(triangleVertices)*4, gl.Ptr(triangleVertices), gl.STATIC_DRAW)

	vertexShader := genShader(vertexShaderSource, gl.VERTEX_SHADER)
	fragmentShader := genShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	shaderProgram = genShaderProgram(vertexShader, fragmentShader)

	gl.UseProgram(shaderProgram)

	uniformLoc = gl.GetUniformLocation(shaderProgram, gl.Str("model\x00"))

	setupVertexAttribs()

	model = mgl32.Ident4()
}

func setupVertexAttribs() {
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)
}

func Render(deltaTime float64) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	rotationAngle := float32(deltaTime) * 50.0
	rotation := mgl32.HomogRotate3D(mgl32.DegToRad(rotationAngle), mgl32.Vec3{0, 1, 0})
	model = model.Mul4(rotation)

	gl.UseProgram(shaderProgram)
	gl.UniformMatrix4fv(uniformLoc, 1, false, &model[0])

	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(triangleVertices)/3))
}

func genShader(source string, shaderType uint32) uint32 {
	shader := gl.CreateShader(shaderType)
	cSources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, cSources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		fmt.Printf("Failed to compile %v: %v\n", shaderType, log)
	}
	return shader
}

func genShaderProgram(vertexShader, fragmentShader uint32) uint32 {
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		fmt.Printf("Failed to link program: %v\n", log)
	}

	gl.ValidateProgram(program)
	gl.GetProgramiv(program, gl.VALIDATE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		fmt.Printf("Program validation failed: %v\n", log)
	}
	return program
}

// My OpenGL Shaders seem to be very old, so I'm using an old version
var vertexShaderSource = `
#version 120
attribute vec3 inPosition;
uniform mat4 model;
void main() {
    gl_Position = model * vec4(inPosition, 1.0);
}` + "\x00"

var fragmentShaderSource = `
#version 120
void main() {
    gl_FragColor = vec4(1.0, 0.0, 0.0, 1.0);
}` + "\x00"
