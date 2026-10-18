//go:build graphics_probe

// Run on a machine with EGL/GLES SDK and a display: go run -tags egl ./scripts/egl-probe/main.go
package main

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/Uzer007-admin/gl/capabilities"
	capnative "github.com/Uzer007-admin/gl/capabilities/native"
	"github.com/Uzer007-admin/gl/egl1.5/egl"
	gles "github.com/Uzer007-admin/gl/v2.0/gles2"
)

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := egl.Init(); err != nil {
		panic(err)
	}
	display := egl.GetDisplay(egl.DEFAULT_DISPLAY)
	var major, minor int32
	if egl.Initialize(display, &major, &minor) != egl.TRUE {
		panic(fmt.Sprintf("eglInitialize: %#x", egl.GetError()))
	}
	defer egl.Terminate(display)
	if egl.BindAPI(egl.OPENGL_ES_API) != egl.TRUE {
		panic("eglBindAPI")
	}
	attributes := []int32{egl.SURFACE_TYPE, egl.PBUFFER_BIT, egl.RENDERABLE_TYPE, egl.OPENGL_ES2_BIT, egl.RED_SIZE, 8, egl.GREEN_SIZE, 8, egl.BLUE_SIZE, 8, egl.ALPHA_SIZE, 8, egl.NONE}
	var config uintptr
	var count int32
	if egl.ChooseConfig(display, &attributes[0], &config, 1, &count) != egl.TRUE || count != 1 {
		panic("eglChooseConfig")
	}
	size := []int32{egl.WIDTH, 8, egl.HEIGHT, 8, egl.NONE}
	surface := egl.CreatePbufferSurface(display, config, &size[0])
	if surface == 0 {
		panic("eglCreatePbufferSurface")
	}
	defer egl.DestroySurface(display, surface)
	version := []int32{egl.CONTEXT_CLIENT_VERSION, 2, egl.NONE}
	context := egl.CreateContext(display, config, 0, &version[0])
	if context == 0 {
		panic("eglCreateContext")
	}
	defer egl.DestroyContext(display, context)
	if egl.MakeCurrent(display, surface, surface, context) != egl.TRUE {
		panic("eglMakeCurrent")
	}
	defer egl.MakeCurrent(display, 0, 0, 0)
	lookups := 0
	if err := gles.InitWithProcAddrFunc(func(name string) unsafe.Pointer { lookups++; return egl.GetProcAddress(gles.Str(name + "\x00")) }); err != nil {
		panic(err)
	}
	gles.ClearColor(0, 1, 0, 1)
	gles.Clear(gles.COLOR_BUFFER_BIT)
	var pixel [4]byte
	gles.ReadPixels(0, 0, 1, 1, gles.RGBA, gles.UNSIGNED_BYTE, unsafe.Pointer(&pixel[0]))
	if pixel != [4]byte{0, 255, 0, 255} || gles.GetError() != gles.NO_ERROR {
		panic(fmt.Sprintf("framebuffer %v", pixel))
	}
	fmt.Printf("EGL %d.%d, %s, GLES2 core lookups=%d, pixel=%v\n", major, minor, gles.GoStr(gles.GetString(gles.VERSION)), lookups, pixel)
	info, err := capnative.Query(func(name string) unsafe.Pointer { return egl.GetProcAddress(gles.Str(name + "\x00")) })
	if err != nil || info.Framebuffers(gles.IsFunctionAvailable) != capabilities.CoreFramebuffer || !info.SeparateBlend(gles.IsFunctionAvailable) {
		panic(fmt.Sprintf("GLES capabilities: %+v; %v", info, err))
	}
	var texture, framebuffer uint32
	gles.GenTextures(1, &texture)
	defer gles.DeleteTextures(1, &texture)
	gles.BindTexture(gles.TEXTURE_2D, texture)
	gles.TexParameteri(gles.TEXTURE_2D, gles.TEXTURE_MIN_FILTER, gles.NEAREST)
	gles.TexParameteri(gles.TEXTURE_2D, gles.TEXTURE_MAG_FILTER, gles.NEAREST)
	gles.TexImage2D(gles.TEXTURE_2D, 0, gles.RGBA, 8, 8, 0, gles.RGBA, gles.UNSIGNED_BYTE, nil)
	gles.GenFramebuffers(1, &framebuffer)
	defer gles.DeleteFramebuffers(1, &framebuffer)
	gles.BindFramebuffer(gles.FRAMEBUFFER, framebuffer)
	defer gles.BindFramebuffer(gles.FRAMEBUFFER, 0)
	gles.FramebufferTexture2D(gles.FRAMEBUFFER, gles.COLOR_ATTACHMENT0, gles.TEXTURE_2D, texture, 0)
	if gles.CheckFramebufferStatus(gles.FRAMEBUFFER) != gles.FRAMEBUFFER_COMPLETE {
		panic("incomplete GLES FBO")
	}
	compile := func(kind uint32, source string) uint32 {
		id := gles.CreateShader(kind)
		strings, free := gles.Strs(source)
		defer free()
		gles.ShaderSource(id, 1, strings, nil)
		gles.CompileShader(id)
		var ok int32
		gles.GetShaderiv(id, gles.COMPILE_STATUS, &ok)
		if ok != gles.TRUE {
			panic("GLES alpha shader failed")
		}
		return id
	}
	vertex := compile(gles.VERTEX_SHADER, "attribute vec2 position;void main(){gl_Position=vec4(position,0.0,1.0);}")
	fragment := compile(gles.FRAGMENT_SHADER, "precision mediump float;void main(){gl_FragColor=vec4(1.0,0.0,0.0,0.5);}")
	defer gles.DeleteShader(vertex)
	defer gles.DeleteShader(fragment)
	program := gles.CreateProgram()
	defer gles.DeleteProgram(program)
	gles.AttachShader(program, vertex)
	gles.AttachShader(program, fragment)
	gles.LinkProgram(program)
	gles.UseProgram(program)
	var linked int32
	gles.GetProgramiv(program, gles.LINK_STATUS, &linked)
	if linked != gles.TRUE {
		panic("GLES alpha program failed")
	}
	var buffer uint32
	gles.GenBuffers(1, &buffer)
	defer gles.DeleteBuffers(1, &buffer)
	points := []float32{-1, -1, 1, -1, -1, 1, 1, 1}
	gles.BindBuffer(gles.ARRAY_BUFFER, buffer)
	gles.BufferData(gles.ARRAY_BUFFER, len(points)*4, gles.SlicePtr(points), gles.STATIC_DRAW)
	position := uint32(gles.GetAttribLocation(program, gles.Str("position\x00")))
	gles.EnableVertexAttribArray(position)
	gles.VertexAttribPointerWithOffset(position, 2, gles.FLOAT, false, 0, 0)
	gles.Viewport(0, 0, 8, 8)
	gles.ClearColor(0, 0.25, 0, 0.25)
	gles.Clear(gles.COLOR_BUFFER_BIT)
	gles.Enable(gles.BLEND)
	gles.BlendFuncSeparate(gles.SRC_ALPHA, gles.ONE_MINUS_SRC_ALPHA, gles.ONE, gles.ONE_MINUS_SRC_ALPHA)
	gles.DrawArrays(gles.TRIANGLE_STRIP, 0, 4)
	gles.ReadPixels(0, 0, 1, 1, gles.RGBA, gles.UNSIGNED_BYTE, unsafe.Pointer(&pixel[0]))
	if pixel[0] < 126 || pixel[0] > 130 || pixel[1] < 30 || pixel[1] > 34 || pixel[2] != 0 || pixel[3] < 157 || pixel[3] > 161 || gles.GetError() != 0 {
		panic(fmt.Sprintf("GLES alpha blend %v", pixel))
	}
	fmt.Printf("GLES %d.%d capabilities and texture-FBO separate alpha blend: %v\n", info.Major, info.Minor, pixel)
}
