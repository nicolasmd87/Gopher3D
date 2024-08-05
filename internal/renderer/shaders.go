package renderer

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// =============================================================
//
//	Shaders
//
// =============================================================
type Shader struct {
	vertexSource   string
	fragmentSource string
	program        uint32
}

func (shader *Shader) Use() {
	gl.UseProgram(shader.program)
}

func (shader *Shader) SetVec3(name string, value mgl32.Vec3) {
	location := gl.GetUniformLocation(shader.program, gl.Str(name+"\x00"))
	gl.Uniform3f(location, value.X(), value.Y(), value.Z())
}

func (shader *Shader) SetFloat(name string, value float32) {
	location := gl.GetUniformLocation(shader.program, gl.Str(name+"\x00"))
	gl.Uniform1f(location, value)
}

func (shader *Shader) SetInt(name string, value int32) {
	location := gl.GetUniformLocation(shader.program, gl.Str(name+"\x00"))
	gl.Uniform1i(location, value)
}

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

var fragmentShaderSource = `// Fragment Shader
#version 330 core
in vec2 fragTexCoord;
in vec3 Normal;
in vec3 FragPos;

uniform sampler2D textureSampler;
uniform struct Light {
    vec3 position;
    vec3 color;
    float intensity;
} light;
uniform vec3 viewPos;
uniform vec3 diffuseColor;
uniform vec3 specularColor;
uniform float shininess;

out vec4 FragColor;

void main() {
    vec4 texColor = texture(textureSampler, fragTexCoord);

    float ambientStrength = 0.1;
    vec3 ambient = ambientStrength * light.color * diffuseColor;

    vec3 norm = normalize(Normal);
    vec3 lightDir = normalize(light.position - FragPos);
    float diff = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = diff * light.color * diffuseColor;

    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 reflectDir = reflect(-lightDir, norm);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), shininess);
    vec3 specular = spec * light.color * specularColor;

    vec3 result = (ambient + diffuse + specular) * light.intensity;
    FragColor = vec4(result, 1.0) * texColor;
}
` + "\x00"

func InitShader() Shader {
	return Shader{
		vertexSource:   vertexShaderSource,
		fragmentSource: fragmentShaderSource,
	}
}
