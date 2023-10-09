package renderer

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

func LoadObject() *Model {
	files, err := os.ReadDir("../obj")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".obj") {
			model, err := LoadModel("../obj/"+file.Name(), "../obj/DirtMetal.jpg")
			if err != nil {
				log.Fatalf("Could not load the obj file %s: %v", file.Name(), err)
			}
			return model
		}
	}
	return nil
}

func LoadModel(filename string, texturePath string) (*Model, error) {
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
				fmt.Println("Error parsing vertex: ", err)
				return nil, err
			}
			model.Vertices = append(model.Vertices, vertex...)
		case "vn":
			normal, err := parseVertex(parts[1:])
			if err != nil {
				fmt.Println("Error parsing normal: ", err)
				return nil, err
			}
			model.Normals = append(model.Normals, normal...)
		case "vt":
			texCoord, err := parseTextureCoordinate(parts[1:])
			if err != nil {
				fmt.Println("Error parsing texture coordinate: ", err)
				return nil, err
			}
			model.TextureCoords = append(model.TextureCoords, texCoord...)
		case "f":
			face, err := parseFace(parts[1:])
			if err != nil {
				fmt.Println("Error parsing face: ", err)
				return nil, err
			}
			model.Faces = append(model.Faces, face...)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Compute the centroid
	var sumX, sumY, sumZ float32
	vertexCount := len(model.Vertices) / 3

	for i := 0; i < vertexCount; i++ {
		sumX += model.Vertices[i*3]
		sumY += model.Vertices[i*3+1]
		sumZ += model.Vertices[i*3+2]
	}

	centroid := mgl32.Vec3{sumX / float32(vertexCount), sumY / float32(vertexCount), sumZ / float32(vertexCount)}

	// Shift all vertex positions
	for i := 0; i < vertexCount; i++ {
		model.Vertices[i*3] -= centroid.X()
		model.Vertices[i*3+1] -= centroid.Y()
		model.Vertices[i*3+2] -= centroid.Z()
	}

	// Load and bind the texture
	textureID, err := loadTexture(texturePath)
	if err != nil {
		return nil, err
	}

	model.TextureID = textureID // Store the texture ID in the Model struct

	return model, nil
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
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.REPEAT)
	// GL_NEAREST results in blocked patterns where we can clearly see the pixels that form the texture while GL_LINEAR produces a smoother pattern where the individual pixels are less visible.
	// GL_LINEAR produces a more realistic output, but some developers prefer a more 8-bit look and as a result pick the GL_NEAREST option
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	return textureID, nil
}
