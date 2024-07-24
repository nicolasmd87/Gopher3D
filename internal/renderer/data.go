package renderer

import (
	"Gopher3D/internal/logger"
	"unsafe"

	lin "github.com/xlab/linmath"
	"go.uber.org/zap"
)

type vkTexCubeUniform struct {
	mvp      lin.Mat4x4
	position [][]float32
	attr     [][]float32
}

const vkTexCubeUniformSize = int(unsafe.Sizeof(vkTexCubeUniform{}))

func (u *vkTexCubeUniform) Data() []byte {
	const m = 0x7fffffff
	return (*[m]byte)(unsafe.Pointer(u))[:vkTexCubeUniformSize]
}

// This was a placeholder for the vertex data of a cube
// The cube was used to test the texture mapping
// I need to use the loader package to load the vertex data
// from an obj file
// mode.Vertices is actually a slice of float32

func (u *vkTexCubeUniform) gVertexBufferData(model *Model) {
	logger.Log.Info("Loading vertex data")
	logger.Log.Info("Model vertices: ", zap.Any("Vertices: ", model.Vertices))
	for i := 0; i < len(model.Vertices); i++ {
		if i+1 > len(model.Vertices) {
			break
		}
		u.position = append(u.position, []float32{model.Vertices[i], model.Vertices[i+1], model.Vertices[i+2]})
		i += 2
	}
	// I Need to load uv or texture coordinate data now
	for i := 0; i < len(model.TextureCoords); i++ {
		if i+1 > len(model.TextureCoords) {
			break
		}
		u.attr = append(u.attr, []float32{model.TextureCoords[i], model.TextureCoords[i+1]})
		i++
	}
}

/*
	var gVertexBufferData = []float32{
		-1.0, -1.0, -1.0, // -X side
		-1.0, -1.0, 1.0,
		-1.0, 1.0, 1.0,
		-1.0, 1.0, 1.0,
		-1.0, 1.0, -1.0,
		-1.0, -1.0, -1.0,

		-1.0, -1.0, -1.0, // -Z side
		1.0, 1.0, -1.0,
		1.0, -1.0, -1.0,
		-1.0, -1.0, -1.0,
		-1.0, 1.0, -1.0,
		1.0, 1.0, -1.0,

		-1.0, -1.0, -1.0, // -Y side
		1.0, -1.0, -1.0,
		1.0, -1.0, 1.0,
		-1.0, -1.0, -1.0,
		1.0, -1.0, 1.0,
		-1.0, -1.0, 1.0,

		-1.0, 1.0, -1.0, // +Y side
		-1.0, 1.0, 1.0,
		1.0, 1.0, 1.0,
		-1.0, 1.0, -1.0,
		1.0, 1.0, 1.0,
		1.0, 1.0, -1.0,

		1.0, 1.0, -1.0, // +X side
		1.0, 1.0, 1.0,
		1.0, -1.0, 1.0,
		1.0, -1.0, 1.0,
		1.0, -1.0, -1.0,
		1.0, 1.0, -1.0,

		-1.0, 1.0, 1.0, // +Z side
		-1.0, -1.0, 1.0,
		1.0, 1.0, 1.0,
		-1.0, -1.0, 1.0,
		1.0, -1.0, 1.0,
		1.0, 1.0, 1.0,
	}

var gUVBufferData = []float32{
	0.0, 1.0, // -X side
	1.0, 1.0,
	1.0, 0.0,
	1.0, 0.0,
	0.0, 0.0,
	0.0, 1.0,

	1.0, 1.0, // -Z side
	0.0, 0.0,
	0.0, 1.0,
	1.0, 1.0,
	1.0, 0.0,
	0.0, 0.0,

	1.0, 0.0, // -Y side
	1.0, 1.0,
	0.0, 1.0,
	1.0, 0.0,
	0.0, 1.0,
	0.0, 0.0,

	1.0, 0.0, // +Y side
	0.0, 0.0,
	0.0, 1.0,
	1.0, 0.0,
	0.0, 1.0,
	1.0, 1.0,

	1.0, 0.0, // +X side
	0.0, 0.0,
	0.0, 1.0,
	0.0, 1.0,
	1.0, 1.0,
	1.0, 0.0,

	0.0, 0.0, // +Z side
	0.0, 1.0,
	1.0, 0.0,
	0.0, 1.0,
	1.0, 1.0,
	1.0, 0.0,
}
*/
