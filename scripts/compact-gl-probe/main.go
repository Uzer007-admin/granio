//go:build graphics_probe

package main

import (
	"fmt"
	"runtime"
	"testing"
	"time"
	"unsafe"

	gl "github.com/Uzer007-admin/gl/v2.1/fyne"
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
)

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
	window, err := glfw.CreateWindow(16, 16, "compact GL", nil, nil)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	window.MakeContextCurrent()
	lookups := 0
	if err := gl.InitWithProcAddrFunc(func(name string) unsafe.Pointer {
		lookups++
		return glfw.GetProcAddress(name)
	}); err != nil {
		panic(err)
	}
	fmt.Println("Go:", runtime.Version(), "lookups:", lookups, "renderer:", gl.GoStr(gl.GetString(gl.RENDERER)))
	allocations := testing.AllocsPerRun(20, func() {
		if err := gl.Init(); err != nil {
			panic(err)
		}
	})
	start := time.Now()
	for i := 0; i < 1000; i++ {
		if err := gl.Init(); err != nil {
			panic(err)
		}
	}
	fmt.Printf("Compact GL Init: %.1f us/op, %.1f Go allocs/op\n", float64(time.Since(start).Nanoseconds())/1000000, allocations)
	gl.ClearColor(0, 1, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	var pixel [4]byte
	gl.ReadPixels(0, 0, 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixel[0]))
	if pixel != [4]byte{0, 255, 0, 255} {
		panic(fmt.Sprint(pixel))
	}
	fmt.Println("Framebuffer:", pixel)
	shader := gl.CreateShader(gl.VERTEX_SHADER)
	source, free := gl.Strs("#version 120\n", "void main(){gl_Position=gl_Vertex;}")
	gl.ShaderSource(shader, 2, source, nil)
	free()
	gl.CompileShader(shader)
	var compiled int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &compiled)
	gl.DeleteShader(shader)
	if compiled != gl.TRUE {
		panic("shader failed")
	}
	fmt.Println("Shader compilation: passed")
}
