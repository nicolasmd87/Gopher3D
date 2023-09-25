package renderer

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func LoadObj(filename string) (*Model, error) {
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
			// vertex
			vertex, err := parseVertex(parts[1:])
			if err != nil {
				return nil, err
			}
			model.Vertices = append(model.Vertices, vertex...)
		case "vn":
			// vertex normal
			normal, err := parseVertex(parts[1:])
			if err != nil {
				return nil, err
			}
			model.Normals = append(model.Normals, normal...)
		case "f":
			// face
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

	fmt.Print("Model loaded...", model)
	return model, nil
}
