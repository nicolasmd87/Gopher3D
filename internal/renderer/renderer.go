package renderer

import (
	"fmt"
	"image"
	"image/draw"
	"os"
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
	Models            []*Model
	shaderProgram     uint32
	modelLoc          int32
	viewProjLoc       int32
	lightPosLoc       int32
	lightColorLoc     int32
	lightIntensityLoc int32
	Debug             bool = false
)

// =============================================================
//
//	Shaders
//
// =============================================================
var vertexShaderSource = `#version 330 core
layout(location = 0) in vec3 inPosition; // Vertex position
layout(location = 1) in vec2 inTexCoord; // Texture Coordinate
layout(location = 2) in vec3 inNormal;   // Vertex normal



uniform mat4 model;
uniform mat4 viewProjection;
out vec2 fragTexCoord;   // Pass to fragment shader
out vec3 Normal;         // Pass normal to fragment shader
out vec3 FragPos;        // Pass position to fragment shader

void main() {
    FragPos = vec3(model * vec4(inPosition, 1.0));
	// Vertex Shader
	Normal = mat3(model) * inNormal; // Use this if the model matrix has no non-uniform scaling
    fragTexCoord = inTexCoord;
    gl_Position = viewProjection * model * vec4(inPosition, 1.0);
}
` + "\x00"

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
	// Culling : https://learnopengl.com/Advanced-OpenGL/Face-culling
	gl.Enable(gl.CULL_FACE)
	// IF FACES OF THE MODEL ARE RENDERED IN THE WRONG ORDER, TRY SWITCHING THE FOLLOWING LINE TO gl.CCW or we need to make sure the winding of each model is consistent
	// CCW = Counter ClockWise
	gl.CullFace(gl.FRONT)
	gl.FrontFace(gl.CW)
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

	var ebo uint32
	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(model.Faces)*4, gl.Ptr(model.Faces), gl.STATIC_DRAW)

	stride := int32((3 + 2 + 3) * 4)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, stride, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, stride, gl.PtrOffset(5*4))
	gl.EnableVertexAttribArray(2)

	model.VAO = vao
	model.VBO = vbo
	model.EBO = ebo
	model.ModelMatrix = mgl32.Ident4()

	Models = append(Models, model)
}

func Render(camera Camera, deltaTime float64, light Light) {
	if Debug {
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	} else {
		// Switch back to solid fill mode
		gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
	}
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	viewProjection := camera.GetViewProjection()
	gl.UseProgram(shaderProgram)
	gl.UniformMatrix4fv(viewProjLoc, 1, false, &viewProjection[0])
	for _, model := range Models {
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

		//RotateModel(model, 0, 1, 0)
		gl.DrawElements(gl.TRIANGLES, int32(len(model.Faces)), gl.UNSIGNED_INT, nil)

	}
}

func Cleanup() {
	for _, model := range Models {
		gl.DeleteVertexArrays(1, &model.VAO)
		gl.DeleteBuffers(1, &model.VBO)
		gl.DeleteBuffers(1, &model.EBO)
	}
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
	gl.DetachShader(program, vertexShader)
	gl.DeleteShader(vertexShader)
	gl.DetachShader(program, fragmentShader)
	gl.DeleteShader(fragmentShader)
	return program
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
		Position:  mgl32.Vec3{0.0, 300.0, 0.0}, // Example position
		Color:     mgl32.Vec3{1.0, 1.0, 1.0},   // White light
		Intensity: 1.0,                         // Full intensity
	}
}

func SetTexture(texturePath string, model *Model) {
	textureID, _ := loadTexture(texturePath)
	model.TextureID = textureID // Store the texture ID in the Model struct
}

// SetPosition sets the position of the model
func (m *Model) SetPosition(x, y, z float32) {
	m.ModelMatrix = mgl32.Translate3D(x, y, z)
}

func loadTexture(filePath string) (uint32, error) { // Consider specifying image format or handling different formats properly

	imgFile, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0, err
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var textureID uint32
	// Like any of the previous objects in OpenGL, textures are referenced with an ID; let's create one:
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D, textureID)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(rgba.Rect.Size().X), int32(rgba.Rect.Size().Y), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))

	// Set texture parameters (optional)
	//gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.REPEAT)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	// GL_NEAREST results in blocked patterns where we can clearly see the pixels that form the texture while GL_LINEAR produces a smoother pattern where the individual pixels are less visible.
	// GL_LINEAR produces a more realistic output, but some developers prefer a more 8-bit look and as a result pick the GL_NEAREST option
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	return textureID, nil
}
