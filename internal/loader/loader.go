package loader

import (
	"Gopher3D/internal/renderer"
	"bufio"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
)

func LoadObjectWithPath(Path string) (*renderer.Model, error) {
	model, err := LoadModel(Path)
	return model, err
}

func LoadObject() *renderer.Model {
	files, err := os.ReadDir("../obj")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".obj") {
			model, err := LoadModel("../obj/" + file.Name())
			if err != nil {
				log.Fatalf("Could not load the obj file %s: %v", file.Name(), err)
			}
			return model
		}
	}
	return nil
}

func LoadModel(filename string) (*renderer.Model, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var vertices []float32
	var textureCoords []float32
	var normals []float32
	var faces []int32

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
			vertices = append(vertices, vertex...)
		case "vn":
			normal, err := parseVertex(parts[1:])
			if err != nil {
				return nil, err
			}
			normals = append(normals, normal...)
		case "vt":
			texCoord, err := parseTextureCoordinate(parts[1:])
			if err != nil {
				return nil, err
			}
			textureCoords = append(textureCoords, texCoord[0], texCoord[1])
		case "f":
			face, err := parseFace(parts[1:])
			if err != nil {
				return nil, err
			}
			faces = append(faces, face...)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	vertexCount := len(vertices) / 3

	for len(textureCoords)/2 < vertexCount {
		textureCoords = append(textureCoords, 0, 0)
	}
	for len(normals)/3 < vertexCount {
		normals = append(normals, 0, 0, 0)

	}

	// Some models have broken normals, so we recalculate them ourselves
	// TODO: Add an option while importing the model
	normals = RecalculateNormals(vertices, faces)

	interleavedData := make([]float32, 0, vertexCount*8)
	for i := 0; i < vertexCount; i++ {
		interleavedData = append(interleavedData, vertices[i*3:i*3+3]...)
		interleavedData = append(interleavedData, textureCoords[i*2:i*2+2]...)
		interleavedData = append(interleavedData, normals[i*3:i*3+3]...)
	}

	model := &renderer.Model{
		InterleavedData: interleavedData,
		Vertices:        vertices,
		Faces:           faces,
	}

	model.Position = [3]float32{0, 0, 0}
	model.Rotation = mgl32.Quat{}
	model.Scale = [3]float32{1, 1, 1}
	model.CalculateBoundingSphere()
	return model, nil
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

func RecalculateNormals(vertices []float32, faces []int32) []float32 {
	var normals = make([]float32, len(vertices))

	// Calculate normals for each face
	for i := 0; i < len(faces); i += 3 {
		idx0 := faces[i] * 3
		idx1 := faces[i+1] * 3
		idx2 := faces[i+2] * 3

		v0 := mgl32.Vec3{vertices[idx0], vertices[idx0+1], vertices[idx0+2]}
		v1 := mgl32.Vec3{vertices[idx1], vertices[idx1+1], vertices[idx1+2]}
		v2 := mgl32.Vec3{vertices[idx2], vertices[idx2+1], vertices[idx2+2]}

		edge1 := v1.Sub(v0)
		edge2 := v2.Sub(v0)
		normal := edge1.Cross(edge2).Normalize()

		// Add this normal to each vertex's normals and average them
		for j := 0; j < 3; j++ {
			normals[idx0+int32(j)] += normal[j]
			normals[idx1+int32(j)] += normal[j]
			normals[idx2+int32(j)] += normal[j]
		}
	}

	// Normalize the normals
	for i := 0; i < len(normals); i += 3 {
		normal := mgl32.Vec3{normals[i], normals[i+1], normals[i+2]}.Normalize()
		normals[i], normals[i+1], normals[i+2] = normal[0], normal[1], normal[2]
	}

	return normals
}
