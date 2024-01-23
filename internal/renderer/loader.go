package renderer

import (
	"bufio"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strings"
)

func LoadObjectWithPath(Path string) (*Model, error) {
	fmt.Println("Loading object from path: " + Path)
	model, err := LoadModel(Path)
	return model, err
}

func LoadObject() *Model {
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

func LoadModel(filename string) (*Model, error) {
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

	model := &Model{
		Vertices: interleavedData,
		Faces:    faces,
	}

	return model, nil
}
