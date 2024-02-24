package renderer

// =============================================================
//
//	Shaders
//
// =============================================================
type Shader struct {
	vertexSource   string
	fragmentSource string
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

var fragmentShaderSource = `
// Fragment Shader
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
out vec4 FragColor;

void main() {
    vec4 texColor = texture(textureSampler, fragTexCoord);
    float ambientStrength = 0.1;
    vec3 ambient = ambientStrength * light.color;
    vec3 norm = normalize(Normal);
    vec3 lightDir = normalize(light.position - FragPos);
    float diff = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = diff * light.color;
    vec3 result = (ambient + diffuse) * light.intensity;
    FragColor = vec4(result, 1.0) * texColor;
}
` + "\x00"

func InitShader() Shader {
	return Shader{
		vertexSource:   vertexShaderSource,
		fragmentSource: fragmentShaderSource,
	}
}
