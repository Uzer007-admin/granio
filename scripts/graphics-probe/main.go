//go:build graphics_probe

// Run explicitly: go run ./scripts/graphics-probe/main.go
package main

import (
	"fmt"
	"runtime"
	"testing"
	"time"
	"unsafe"

	"github.com/Uzer007-admin/gl/v2.1/gl"
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
)

func sample(name string, operation func()) {
	const iterations = 200000
	allocations := testing.AllocsPerRun(1000, operation)
	for repeat := 0; repeat < 3; repeat++ {
		start := time.Now()
		for i := 0; i < iterations; i++ {
			operation()
		}
		fmt.Printf("sample %s: %.1f ns/op, %.3f allocs/op\n", name, float64(time.Since(start).Nanoseconds())/iterations, allocations)
	}
}

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()
	glfw.WindowHint(glfw.Visible, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	w, err := glfw.CreateWindow(8, 8, "graphics stack probe", nil, nil)
	if err != nil {
		panic(err)
	}
	defer w.Destroy()
	w.MakeContextCurrent()
	start := time.Now()
	if err := gl.Init(); err != nil {
		panic(err)
	}
	initDuration := time.Since(start)
	fmt.Println("Go:", runtime.Version(), "GLFW:", glfw.GetVersionString())
	fmt.Println("OpenGL:", gl.GoStr(gl.GetString(gl.VERSION)), "renderer:", gl.GoStr(gl.GetString(gl.RENDERER)))
	fmt.Println("First GL Init:", initDuration)
	lookups := 0
	start = time.Now()
	if err := gl.InitWithProcAddrFunc(func(name string) unsafe.Pointer {
		lookups++
		return glfw.GetProcAddress(name)
	}); err != nil {
		panic(err)
	}
	fmt.Println("Init via GLFW:", time.Since(start), "lookups:", lookups)
	allocations := testing.AllocsPerRun(20, func() {
		if err := gl.Init(); err != nil {
			panic(err)
		}
	})
	start = time.Now()
	for i := 0; i < 50; i++ {
		if err := gl.Init(); err != nil {
			panic(err)
		}
	}
	fmt.Printf("Repeated GL Init: %.1f us/op, %.1f Go allocs/op\n", float64(time.Since(start).Nanoseconds())/50000, allocations)
	gl.ClearColor(0, 1, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	var pixel [4]byte
	gl.ReadPixels(0, 0, 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixel[0]))
	if pixel != [4]byte{0, 255, 0, 255} || gl.GetError() != gl.NO_ERROR {
		panic(fmt.Sprintf("Unexpected framebuffer: %v", pixel))
	}
	fmt.Println("Framebuffer:", pixel)
	strings, free := gl.Strs("A", "B")
	fmt.Printf("Strs without terminators: %q %q\n", gl.GoStr(*strings), gl.GoStr(unsafe.Slice(strings, 2)[1]))
	free()
	shader := gl.CreateShader(gl.VERTEX_SHADER)
	source, release := gl.Strs("#version 120\n", "void main(){gl_Position=gl_Vertex;}")
	gl.ShaderSource(shader, 2, source, nil)
	release()
	gl.CompileShader(shader)
	var compiled int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &compiled)
	gl.DeleteShader(shader)
	if compiled != gl.TRUE {
		panic("NUL-terminated shader parts failed to compile")
	}
	var pbo uint32
	gl.GenBuffers(1, &pbo)
	gl.BindBuffer(gl.PIXEL_PACK_BUFFER, pbo)
	gl.BufferData(gl.PIXEL_PACK_BUFFER, 4, nil, gl.STREAM_READ)
	gl.ReadPixelsWithOffset(0, 0, 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, 0)
	gl.GetBufferSubData(gl.PIXEL_PACK_BUFFER, 0, 4, unsafe.Pointer(&pixel[0]))
	gl.BindBuffer(gl.PIXEL_PACK_BUFFER, 0)
	gl.DeleteBuffers(1, &pbo)
	if pixel != [4]byte{0, 255, 0, 255} || gl.GetError() != gl.NO_ERROR {
		panic("PBO offset read failed")
	}
	sample("GetError", func() { gl.GetError() })
	sample("ClearColor", func() { gl.ClearColor(0, 1, 0, 1) })
	data := make([]float32, 32)
	sample("SlicePtr(slice)", func() {
		if gl.SlicePtr(data) == nil {
			panic("nil pointer")
		}
	})
	sample("Ptr(slice)", func() {
		if gl.Ptr(data) == nil {
			panic("nil pointer")
		}
	})
	var messages int
	gl.Enable(gl.DEBUG_OUTPUT)
	gl.Enable(gl.DEBUG_OUTPUT_SYNCHRONOUS)
	gl.DebugMessageCallback(func(_, _, _, _ uint32, _ int32, _ string, _ unsafe.Pointer) { messages++ }, nil)
	message := "graphics callback probe"
	gl.DebugMessageInsert(gl.DEBUG_SOURCE_APPLICATION, gl.DEBUG_TYPE_OTHER, 42, gl.DEBUG_SEVERITY_NOTIFICATION, int32(len(message)), gl.Str(message+"\x00"))
	gl.DebugMessageCallback(nil, nil)
	fmt.Println("Synchronous debug callback messages:", messages)
	if messages == 0 {
		panic("Debug callback was not invoked")
	}
}
