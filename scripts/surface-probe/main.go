//go:build graphics_probe

// Surface probe exercises two native windows and alpha readback with an
// explicit backend. It is not a desktop-compositor backdrop acceptance test.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"
	"unsafe"

	gl "github.com/Uzer007-admin/gl/v2.1/fyne"
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
)

func main() {
	backend := flag.String("backend", "auto", "auto, x11 or wayland")
	flag.Parse()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	switch *backend {
	case "x11":
		glfw.InitHint(glfw.PlatformHint, int(glfw.PlatformX11))
	case "wayland":
		glfw.InitHint(glfw.PlatformHint, int(glfw.PlatformWayland))
	case "auto":
	default:
		panic("unknown backend")
	}
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	if glfw.GetPlatform() == glfw.PlatformWayland {
		glfw.WindowHint(glfw.ContextCreationAPI, glfw.EGLContextAPI)
	}
	glfw.WindowHint(glfw.Visible, glfw.False)
	glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)
	alpha, err := glfw.CreateWindow(192, 128, "alpha surface", nil, nil)
	if err != nil {
		panic(err)
	}
	defer alpha.Destroy()
	glfw.WindowHint(glfw.TransparentFramebuffer, glfw.False)
	opaque, err := glfw.CreateWindow(224, 160, "opaque surface", nil, nil)
	if err != nil {
		panic(err)
	}
	defer opaque.Destroy()
	alpha.MakeContextCurrent()
	if err := gl.InitWithProcAddrFunc(glfw.GetProcAddress); err != nil {
		panic(err)
	}
	if !gl.IsFunctionAvailable("glViewport") || gl.IsFunctionAvailable("glImaginaryProcedure") {
		panic("invalid procedure availability")
	}
	alpha.Show()
	opaque.Show()
	deadline := time.Now().Add(250 * time.Millisecond)
	for {
		for _, window := range []*glfw.Window{alpha, opaque} {
			window.MakeContextCurrent()
			gl.ClearColor(0.25, 0, 0, 0.5)
			gl.Clear(gl.COLOR_BUFFER_BIT)
			window.SwapBuffers()
		}
		glfw.PollEvents()
		if time.Now().After(deadline) {
			break
		}
		glfw.WaitEventsTimeout(0.01)
	}
	for _, window := range []*glfw.Window{alpha, opaque} {
		state := window.GetSurfaceMetrics()
		if _, _, valid := state.PixelScale(); !valid {
			panic(fmt.Sprintf("invalid native geometry: %+v", state))
		}
	}
	alpha.MakeContextCurrent()
	gl.ClearColor(0.25, 0, 0, 0.5)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	var pixel [4]byte
	gl.ReadPixels(0, 0, 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixel[0]))
	if pixel[0] < 63 || pixel[0] > 65 || pixel[1] != 0 || pixel[2] != 0 || pixel[3] < 127 || pixel[3] > 129 {
		panic(fmt.Sprintf("unexpected premultiplied alpha pixel: %v", pixel))
	}
	result := struct {
		Go, OS, Backend, Renderer string
		Platform                  glfw.Platform
		Alpha, Opaque             glfw.SurfaceMetrics
		AlphaPixel                [4]byte
	}{runtime.Version(), runtime.GOOS, *backend, gl.GoStr(gl.GetString(gl.RENDERER)), glfw.GetPlatform(),
		alpha.GetSurfaceMetrics(), opaque.GetSurfaceMetrics(), pixel}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		panic(err)
	}
}
