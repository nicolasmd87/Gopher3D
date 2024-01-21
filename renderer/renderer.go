package renderer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Model struct {
	Id            int
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
	models            []*Model
	shaderProgram     uint32
	modelLoc          int32
	viewProjLoc       int32
	lightPosLoc       int32
	lightColorLoc     int32
	lightIntensityLoc int32
	Debug             bool = false
)

type Light struct {
	Position  mgl32.Vec3
	Color     mgl32.Vec3
	Intensity float32
}

func Init(width, height int32) {
	if err := gl.Init(); err != nil {
		fmt.Println("OpenGL initialization failed:", err)
		return
	}

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)
	gl.Viewport(0, 0, width, height)
	initOpenGL()
}

func initOpenGL() {
	vertexShader := genShader(vertexShaderSource, gl.VERTEX_SHADER)
	fragmentShader := genShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	shaderProgram = genShaderProgram(vertexShader, fragmentShader)

	gl.UseProgram(shaderProgram)

	modelLoc = gl.GetUniformLocation(shaderProgram, gl.Str("model\x00"))
	viewProjLoc = gl.GetUniformLocation(shaderProgram, gl.Str("viewProjection\x00"))

	// Set light properties
	lightPosLoc = gl.GetUniformLocation(shaderProgram, gl.Str("light.position\x00"))
	lightColorLoc = gl.GetUniformLocation(shaderProgram, gl.Str("light.color\x00"))
	lightIntensityLoc = gl.GetUniformLocation(shaderProgram, gl.Str("light.intensity\x00"))

}

func AddModel(model *Model) {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(model.Vertices)*4, gl.Ptr(model.Vertices), gl.STATIC_DRAW)

	stride := int32((3 + 2 + 3) * 4)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, stride, gl.PtrOffset(1*4))
	gl.EnableVertexAttribArray(1)

	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, stride, gl.PtrOffset(5*4))
	gl.EnableVertexAttribArray(2)

	var ebo uint32
	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(model.Faces)*4, gl.Ptr(model.Faces), gl.STATIC_DRAW)

	if Debug {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	}

	model.VAO = vao
	model.VBO = vbo
	model.EBO = ebo
	model.ModelMatrix = mgl32.Ident4()

	models = append(models, model)
}

func Render(camera Camera, deltaTime float64, light Light) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	viewProjection := camera.GetViewProjection()
	gl.UseProgram(shaderProgram)
	gl.UniformMatrix4fv(viewProjLoc, 1, false, &viewProjection[0])
	for _, model := range models {
		gl.BindVertexArray(model.VAO)
		gl.UniformMatrix4fv(modelLoc, 1, false, &model.ModelMatrix[0])

		gl.Uniform3f(lightPosLoc, light.Position[0], light.Position[1], light.Position[2])
		gl.Uniform3f(lightColorLoc, light.Color[0], light.Color[1], light.Color[2])
		gl.Uniform1f(lightIntensityLoc, light.Intensity)

		// Activate the first texture unit and bind your texture
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, model.TextureID)
		// Get the uniform location of the texture sampler in your shader program
		textureUniform := gl.GetUniformLocation(shaderProgram, gl.Str("uTexture\x00"))
		// Set the sampler to the first texture unit
		gl.Uniform1i(textureUniform, 0)

		//RotateModel(model, 1, 1, 0)
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
var vertexShaderSource = `#version 330 core
layout(location = 0) in vec3 inPosition; // Vertex position
layout(location = 1) in vec3 inNormal;   // Vertex normal
layout(location = 2) in vec2 inTexCoord; // Texture Coordinate


uniform mat4 model;
uniform mat4 viewProjection;
out vec2 fragTexCoord;   // Pass to fragment shader
out vec3 Normal;         // Pass normal to fragment shader
out vec3 FragPos;        // Pass position to fragment shader

void main() {
    FragPos = vec3(model * vec4(inPosition, 1.0));
    Normal = mat3(transpose(inverse(model))) * inNormal; // Transforming the normal
    fragTexCoord = inTexCoord;
    gl_Position = viewProjection * model * vec4(inPosition, 1.0);
}

` + "\x00"

// Need to normalize the texture coordinates
var fragmentShaderSource = `#version 330 core
in vec2 fragTexCoord; // Received from vertex shader
in vec3 Normal;       // Received from vertex shader
in vec3 FragPos;      // Received from vertex shader

uniform sampler2D textureSampler;

struct Light {
    vec3 position;
    vec3 color;
    float intensity;
};

uniform Light light;  // Light source uniform
uniform vec3 viewPos; // Camera position (for future use)

out vec4 FragColor;

void main() {
    // Texture color
    vec4 texColor = texture(textureSampler, fragTexCoord);

    // Ambient lighting
    float ambientStrength = 0.1;
    vec3 ambient = ambientStrength * light.color;

    // Diffuse lighting
    vec3 norm = normalize(Normal);
    vec3 lightDir = normalize(light.position - FragPos);
    float diff = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = diff * light.color;

    // Combining the lighting components
    vec3 result = (ambient + diffuse) * light.intensity;
    FragColor = vec4(result, 1.0) * texColor; // Modulate with texture color
} 
` + "\x00"

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
	gl.DetachShader(program, vertexShader)
	gl.DeleteShader(vertexShader)
	gl.DetachShader(program, fragmentShader)
	gl.DeleteShader(fragmentShader)
	return program
}

func parseVertex(parts []string) ([]float32, error) {
	var vertex []float32
	for _, part := range parts {
		val, err := strconv.ParseFloat(part, 32)
		if err != nil {
			return nil, fmt.Errorf("Invalid vertex value %v: %v", part, err)
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
			return nil, fmt.Errorf("Invalid face index %v: %v", vals[0], err)
		}
		face = append(face, int32(idx-1)) // .obj indices start at 1, not 0
	}

	// Convert quads to triangles
	if len(face) == 4 {
		return []int32{face[0], face[1], face[2], face[0], face[2], face[3]}, nil
	} else {
		return face, nil
	}
}

// for 2D textures
func parseTextureCoordinate(parts []string) ([]float32, error) {
	var texCoord []float32
	for _, part := range parts {
		val, err := strconv.ParseFloat(part, 32)
		if err != nil {
			return nil, fmt.Errorf("Invalid texture coordinate value %v: %v", part, err)
		}
		texCoord = append(texCoord, float32(val))
	}
	return texCoord, nil
}

func RotateModel(model *Model, angleX, angleY float32, angleZ float32) {
	// Create rotation matrices for X and Y axes
	rotationX := mgl32.HomogRotate3DX(mgl32.DegToRad(angleX))
	rotationY := mgl32.HomogRotate3DY(mgl32.DegToRad(angleY))
	rotationZ := mgl32.HomogRotate3DY(mgl32.DegToRad(angleZ))

	// Apply the rotations to the model's ModelMatrix
	model.ModelMatrix = model.ModelMatrix.Mul4(rotationX).Mul4(rotationY).Mul4(rotationZ)
}

func CreateLight() Light {
	return Light{
		Position:  mgl32.Vec3{1.0, 100.0, 1.0}, // Example position
		Color:     mgl32.Vec3{1.0, 1.0, 1.0},   // White light
		Intensity: 3.0,                         // Full intensity
	}
}
