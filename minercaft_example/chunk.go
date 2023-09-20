package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/ojrac/opensimplex-go"
)

const (
	ChunkSizeX = 16
	ChunkSizeY = 64
	ChunkSizeZ = 16
)

type Chunk struct {
	VAO      uint32
	VBO      uint32
	Vertices []float32
}

func NewChunk(vertices []float32) *Chunk {
	chunk := &Chunk{Vertices: vertices}
	chunk.setupRender()
	return chunk
}

func (chunk *Chunk) setupRender() {
	gl.GenVertexArrays(1, &chunk.VAO)
	gl.GenBuffers(1, &chunk.VBO)

	gl.BindVertexArray(chunk.VAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, chunk.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(chunk.Vertices)*4, gl.Ptr(chunk.Vertices), gl.STATIC_DRAW)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func (chunk *Chunk) Render() {
	gl.BindVertexArray(chunk.VAO)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(chunk.Vertices)/3))
	gl.BindVertexArray(0)
}

// CreateChunkVertices generates the vertices for a chunk.
func CreateChunkVertices() []float32 {
	vertices := make([]float32, 0)

	// Generate vertices for each block in the chunk
	for x := 0; x < ChunkSizeX; x++ {
		for y := 0; y < ChunkSizeY; y++ {
			for z := 0; z < ChunkSizeZ; z++ {
				// Add vertices for each face of the block
				blockVertices := generateBlockVertices(float32(x), float32(y), float32(z))
				vertices = append(vertices, blockVertices...)
			}
		}
	}

	return vertices
}

// Generate vertices for a single block.
func generateBlockVertices(x, y, z float32) []float32 {
	// Calculate block position
	blockPos := mgl32.Vec3{x, y, z}

	// Define vertices for a cube
	cubeVertices := []float32{
		// Front face
		-0.5, -0.5, 0.5,
		0.5, -0.5, 0.5,
		0.5, 0.5, 0.5,
		0.5, 0.5, 0.5,
		-0.5, 0.5, 0.5,
		-0.5, -0.5, 0.5,
		// ... Other faces
	}

	// Translate cube vertices based on block position
	for i := 0; i < len(cubeVertices); i += 3 {
		cubeVertices[i] += blockPos[0]
		cubeVertices[i+1] += blockPos[1]
		cubeVertices[i+2] += blockPos[2]
	}

	return cubeVertices
}

func generateTerrainChunk(chunkX, chunkZ int) *Chunk {
	vertices := make([]float32, 0)

	for x := 0; x < ChunkSizeX; x++ {
		for z := 0; z < ChunkSizeZ; z++ {
			// Calculate terrain height using a simple Perlin noise function
			terrainHeight := perlinNoise(float64(chunkX*ChunkSizeX+x)/20.0, float64(chunkZ*ChunkSizeZ+z)/20.0) * 20.0

			for y := 0; y < ChunkSizeY; y++ {
				if float64(y) < terrainHeight {
					blockVertices := generateBlockVertices(float32(x), float32(y), float32(z))
					vertices = append(vertices, blockVertices...)
				}
			}
		}
	}

	return NewChunk(vertices)
}

func perlinNoise(x, y float64) float64 {
	noiseGenerator := opensimplex.NewNormalized(0)
	return noiseGenerator.Eval2(x, y)
}

func main() {
	// Initialize OpenGL and window here
	// ...

	// Create terrain chunks and store them in a 2D array
	if err := glfw.Init(); err != nil {
		panic("failed to initialize GLFW")
	}
	defer glfw.Terminate()

	// Set GLFW window hints
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Create a GLFW window
	window, _ := glfw.CreateWindow(800, 600, "My Window", nil, nil)
	var chunks [][]*Chunk
	numChunksX := 8
	numChunksZ := 8

	for x := 0; x < numChunksX; x++ {
		chunks = append(chunks, []*Chunk{})
		for z := 0; z < numChunksZ; z++ {
			chunk := generateTerrainChunk(x, z)
			chunks[x] = append(chunks[x], chunk)
		}
	}

	// Main loop
	for !window.ShouldClose() {
		// Clear the buffer and render the chunks
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		for x := 0; x < numChunksX; x++ {
			for z := 0; z < numChunksZ; z++ {
				chunks[x][z].Render()
			}
		}

		// Poll events and swap buffers
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
