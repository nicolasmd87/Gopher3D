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
)

func LoadObjectWithPath(Path string) (*renderer.Model, error) {
	fmt.Println("Loading object from path: " + Path)
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

	interleavedData := make([]float32, 0, vertexCount*8)
	for i := 0; i < vertexCount; i++ {
		interleavedData = append(interleavedData, vertices[i*3:i*3+3]...)
		interleavedData = append(interleavedData, textureCoords[i*2:i*2+2]...)
		interleavedData = append(interleavedData, normals[i*3:i*3+3]...)
	}

	model := &renderer.Model{
		Vertices: interleavedData,
		Faces:    faces,
	}

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
