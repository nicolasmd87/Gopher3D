package renderer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Model struct {
	Vertices      []float32
	Normals       []float32
	Faces         []int32
	TextureCoords []float32
	TextureID     uint32
	VAO           uint32 // Vertex Array Object
	VBO           uint32 // Vertex Buffer Object
	EBO           uint32 // Element Buffer Object
	ModelMatrix   mgl32.Mat4
}

var (
	models        []*Model
	shaderProgram uint32
	modelLoc      int32
	viewProjLoc   int32
)

func Init() {
	if err := gl.Init(); err != nil {
		fmt.Println("OpenGL initialization failed:", err)
		return
	}
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Viewport(0, 0, 800, 600)
	initOpenGL()
}

func initOpenGL() {
	vertexShader := genShader(vertexShaderSource, gl.VERTEX_SHADER)
	fragmentShader := genShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	shaderProgram = genShaderProgram(vertexShader, fragmentShader)

	gl.UseProgram(shaderProgram)
	modelLoc = gl.GetUniformLocation(shaderProgram, gl.Str("model\x00"))
	viewProjLoc = gl.GetUniformLocation(shaderProgram, gl.Str("viewProjection\x00"))
}

func AddModel(model *Model) {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(model.Vertices)*4+len(model.TextureCoords)*4, nil, gl.STATIC_DRAW)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(model.Vertices)*4, gl.Ptr(model.Vertices))
	gl.BufferSubData(gl.ARRAY_BUFFER, len(model.Vertices)*4, len(model.TextureCoords)*4, gl.Ptr(model.TextureCoords))

	var ebo uint32
	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(model.Faces)*4, gl.Ptr(model.Faces), gl.STATIC_DRAW)

	stride := int32(5 * 4)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, nil)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, stride, gl.PtrOffset(3*4))

	model.VAO = vao
	model.VBO = vbo
	model.EBO = ebo
	model.ModelMatrix = mgl32.Ident4() // Adjust with actual model matrix

	models = append(models, model)
}

func Render(camera Camera, deltaTime float64) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	view := mgl32.LookAtV(camera.position, camera.position.Add(camera.front), camera.up)
	viewProjection := camera.projection.Mul4(view)
	gl.UseProgram(shaderProgram)
	gl.UniformMatrix4fv(viewProjLoc, 1, false, &viewProjection[0])

	for _, model := range models {
		gl.BindVertexArray(model.VAO)
		gl.UniformMatrix4fv(modelLoc, 1, false, &model.ModelMatrix[0])
		gl.DrawElements(gl.TRIANGLES, int32(len(model.Faces)), gl.UNSIGNED_INT, nil)
	}
}

func Cleanup() {
	for _, model := range models {
		gl.DeleteVertexArrays(1, &model.VAO)
		gl.DeleteBuffers(1, &model.VBO)
		gl.DeleteBuffers(1, &model.EBO)
	}
}

// Shader modifications
var vertexShaderSource = `
#version 330 core
layout(location = 0) in vec3 inPosition;
layout(location = 1) in vec2 inTexCoord;
uniform mat4 model;
uniform mat4 viewProjection;
out vec2 fragTexCoord;

void main() {
	gl_Position = viewProjection * model * vec4(inPosition, 1.0);
	fragTexCoord = inTexCoord;
}` + "\x00"

// Need to normalize the texture coordinates
var fragmentShaderSource = `
#version 330 core
in vec2 fragTexCoord;
uniform sampler2D textureSampler;
out vec4 FragColor;

void main() {
	FragColor = texture(textureSampler, fragTexCoord*0.01);
}` + "\x00"

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

// for 2D textures
func parseTextureCoordinate(parts []string) ([]float32, error) {
	var texCoord []float32
	for _, part := range parts {
		val, err := strconv.ParseFloat(part, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid texture coordinate value %v: %v", part, err)
		}
		texCoord = append(texCoord, float32(val))
	}
	return texCoord, nil
}
