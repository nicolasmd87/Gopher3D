package loader

import "Gopher3D/internal/renderer"

func CombineStaticModels(models []*renderer.Model) (*renderer.Model, error) {
	var combinedVertices []float32
	var combinedTexCoords []float32
	var combinedNormals []float32
	var combinedFaces []int32
	var combinedInterleavedData []float32
	var currentIndexOffset uint32 = 0

	for _, model := range models {
		// Directly append all vertices, texture coords, and normals
		combinedVertices = append(combinedVertices, model.Vertices...)
		combinedTexCoords = append(combinedTexCoords, model.TextureCoords...)
		combinedNormals = append(combinedNormals, model.Normals...)

		// Properly adjust and append each index in the model's faces
		for _, index := range model.Faces {
			// Adjust index by the current vertex offset
			adjustedIndex := index + int32(currentIndexOffset)
			combinedFaces = append(combinedFaces, adjustedIndex)
		}

		// Update currentIndexOffset for the next model based on the number of vertices added
		currentIndexOffset += uint32(len(model.Vertices) / 3)

		// Assuming interleaved data format is [Vertex, TexCoord, Normal], repeat for all vertices
		for i := 0; i < len(model.Vertices)/3; i++ {
			combinedInterleavedData = append(combinedInterleavedData, model.Vertices[i*3:i*3+3]...) // Vertex
			if len(model.TextureCoords) > i*2+1 {
				combinedInterleavedData = append(combinedInterleavedData, model.TextureCoords[i*2:i*2+2]...) // TexCoord
			} else {
				combinedInterleavedData = append(combinedInterleavedData, 0, 0) // Default TexCoord if missing
			}
			if len(model.Normals) > i*3+2 {
				combinedInterleavedData = append(combinedInterleavedData, model.Normals[i*3:i*3+3]...) // Normal
			} else {
				combinedInterleavedData = append(combinedInterleavedData, 0, 0, 0) // Default Normal if missing
			}
		}
	}

	// Create a new model that represents the combined static batch
	combinedModel := &renderer.Model{
		InterleavedData: combinedInterleavedData,
		Vertices:        combinedVertices,
		TextureCoords:   combinedTexCoords,
		Normals:         combinedNormals,
		Faces:           combinedFaces,
	}

	return combinedModel, nil
}
