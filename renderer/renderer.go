package renderer

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Model struct {
	Vertices []float32
	Normals  []float32
	Faces    []int32
}

type Camera struct {
	position, front, up mgl32.Vec3
	fov                 float32
	projection          mgl32.Mat4
	view                mgl32.Mat4
}

var (
	models        []*Model
	shaderProgram uint32
	uniformLoc    int32
	camera        Camera
)

// Init initializes OpenGL and sets up the camera.
func Init() {
	if err := gl.Init(); err != nil {
		fmt.Println("OpenGL initialization failed:", err)
		return
	}
	initOpenGL()
	initCamera()
}

func initCamera() {
	camera = Camera{
		position: mgl32.Vec3{1, 0, 100}, // Adjust the camera position as needed
		front:    mgl32.Vec3{0, 0, -10},
		up:       mgl32.Vec3{0, -1, 0},
		fov:      45.0, // Adjust the field of view as needed
	}

	projection := mgl32.Perspective(mgl32.DegToRad(camera.fov), float32(800)/float32(600), 0.1, 100.0)
	camera.projection = projection
}

func initOpenGL() {
	vertexShader := genShader(vertexShaderSource, gl.VERTEX_SHADER)
	fragmentShader := genShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	shaderProgram = genShaderProgram(vertexShader, fragmentShader)

	gl.UseProgram(shaderProgram)
	uniformLoc = gl.GetUniformLocation(shaderProgram, gl.Str("model\x00"))
}

func AddModel(model *Model) {
	models = append(models, model)
}

func Render(deltaTime float64) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// Update camera view matrix (e.g., based on user input or animation)
	view := mgl32.LookAtV(camera.position, camera.position.Add(camera.front), camera.up)
	camera.view = view

	// Combine view and projection matrices
	viewProjection := camera.projection.Mul4(camera.view)

	gl.UseProgram(shaderProgram)
	gl.UniformMatrix4fv(uniformLoc, 1, false, &viewProjection[0])
	DrawModels()
}

func DrawModels() {
	for _, model := range models {
		// Create a Vertex Array Object (VAO)
		var vao uint32
		gl.GenVertexArrays(1, &vao)
		gl.BindVertexArray(vao)

		// Create a Vertex Buffer Object (VBO) and copy the vertex data to it
		var vbo uint32
		gl.GenBuffers(1, &vbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferData(gl.ARRAY_BUFFER, len(model.Vertices)*4, gl.Ptr(model.Vertices), gl.STATIC_DRAW)

		// Create an Element Buffer Object (EBO) and copy the index data to it
		var ebo uint32
		gl.GenBuffers(1, &ebo)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(model.Faces)*4, gl.Ptr(model.Faces), gl.STATIC_DRAW)

		// Set the vertex attributes pointers
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil) // 3*4 because each vertex consists of 3 floats, each 4 bytes
		gl.EnableVertexAttribArray(0)

		// Draw the triangles
		gl.BindVertexArray(vao)
		gl.DrawElements(gl.TRIANGLES, int32(len(model.Faces)), gl.UNSIGNED_INT, nil)

		// Delete the VAO, VBO, and EBO
		gl.DeleteVertexArrays(1, &vao)
		gl.DeleteBuffers(1, &vbo)
		gl.DeleteBuffers(1, &ebo)
	}
}

func LoadObject(filename string) (*Model, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	model := &Model{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "v":
			vertex, err := parseVertex(parts[1:])
			if err != nil {
				return nil, err
			}
			model.Vertices = append(model.Vertices, vertex...)
		case "vn":
			normal, err := parseVertex(parts[1:])
			if err != nil {
				return nil, err
			}
			model.Normals = append(model.Normals, normal...)
		case "f":
			face, err := parseFace(parts[1:])
			if err != nil {
				return nil, err
			}
			model.Faces = append(model.Faces, face...)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return model, nil
}

func parseVertex(parts []string) ([]float32, error) {
	var vertex []float32
	for _, part := range parts {
		val, err := strconv.ParseFloat(part, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid vertex value %v: %v", part, err)
		}
		vertex = append(vertex, float32(val))
	}
	return vertex, nil
}

func parseFace(parts []string) ([]int32, error) {
	var face []int32
	for _, part := range parts {
		vals := strings.Split(part, "/")
		idx, err := strconv.ParseInt(vals[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid face index %v: %v", vals[0], err)
		}
		face = append(face, int32(idx-1)) // .obj indices start at 1, not 0
	}
	return face, nil
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

		fmt.Printf("Failed to compile %v shader: %v\n", shaderType, log)
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

/*
var vertexShaderSource = `
#version 410 core
layout(location = 0) in vec3 inPosition;
uniform mat4 model;
void main() {
    gl_Position = model * vec4(inPosition, 1.0);
}` + "\x00"

var fragmentShaderSource = `
#version 410 core
out vec4 fragColor;
void main() {
    fragColor = vec4(1.0, 0.0, 0.0, 1.0); // Example color: Red
}` + "\x00"
*/
