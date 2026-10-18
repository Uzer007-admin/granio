//go:build graphics_probe

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"unsafe"

	"github.com/Uzer007-admin/gl/capabilities"
	capnative "github.com/Uzer007-admin/gl/capabilities/native"
	gl "github.com/Uzer007-admin/gl/v2.1/material"
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
)

func shader(kind uint32, source string) uint32 {
	id := gl.CreateShader(kind)
	strings, free := gl.Strs(source)
	defer free()
	gl.ShaderSource(id, 1, strings, nil)
	gl.CompileShader(id)
	var compiled int32
	gl.GetShaderiv(id, gl.COMPILE_STATUS, &compiled)
	if compiled != gl.TRUE {
		panic("material shader failed")
	}
	return id
}

func checkTarget(path capabilities.FramebufferAPI, program uint32) [][4]byte {
	var texture, framebuffer uint32
	gl.GenTextures(1, &texture)
	defer gl.DeleteTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, 8, 8, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	if path == capabilities.CoreFramebuffer {
		gl.GenFramebuffers(1, &framebuffer)
		defer gl.DeleteFramebuffers(1, &framebuffer)
		gl.BindFramebuffer(gl.FRAMEBUFFER, framebuffer)
		defer gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, texture, 0)
		if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
			panic("incomplete core FBO")
		}
	} else {
		gl.GenFramebuffersEXT(1, &framebuffer)
		defer gl.DeleteFramebuffersEXT(1, &framebuffer)
		gl.BindFramebufferEXT(gl.FRAMEBUFFER_EXT, framebuffer)
		defer gl.BindFramebufferEXT(gl.FRAMEBUFFER_EXT, 0)
		gl.FramebufferTexture2DEXT(gl.FRAMEBUFFER_EXT, gl.COLOR_ATTACHMENT0_EXT, gl.TEXTURE_2D, texture, 0)
		if gl.CheckFramebufferStatusEXT(gl.FRAMEBUFFER_EXT) != gl.FRAMEBUFFER_COMPLETE_EXT {
			panic("incomplete EXT FBO")
		}
	}
	gl.Viewport(0, 0, 8, 8)
	gl.Enable(gl.BLEND)
	gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
	gl.ColorMask(true, true, true, true)
	location := gl.GetUniformLocation(program, gl.Str("alpha\x00"))
	var pixels [][4]byte
	for _, alpha := range []float32{0, 0.25, 0.5, 0.75, 1} {
		gl.ClearColor(0, 0.25, 0, 0.25)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.Uniform1f(location, alpha)
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
		var pixel [4]byte
		gl.ReadPixels(0, 0, 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixel[0]))
		expected := [4]float64{float64(alpha), 0.25 * (1 - float64(alpha)), 0, float64(alpha) + 0.25*(1-float64(alpha))}
		for channel := range pixel {
			if math.Abs(float64(pixel[channel])-255*expected[channel]) > 2 {
				panic(fmt.Sprintf("incorrect separate alpha blend: path=%d alpha=%v pixel=%v", path, alpha, pixel))
			}
		}
		pixels = append(pixels, pixel)
	}
	if err := gl.GetError(); err != 0 {
		panic(fmt.Sprintf("GL error 0x%x", err))
	}
	return pixels
}

func main() {
	backend := flag.String("backend", "auto", "auto, x11 or wayland")
	flag.Parse()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if *backend == "wayland" {
		glfw.InitHint(glfw.PlatformHint, int(glfw.PlatformWayland))
	} else if *backend == "x11" {
		glfw.InitHint(glfw.PlatformHint, int(glfw.PlatformX11))
	}
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()
	glfw.WindowHint(glfw.Visible, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	if glfw.GetPlatform() == glfw.PlatformWayland {
		glfw.WindowHint(glfw.ContextCreationAPI, glfw.EGLContextAPI)
	}
	window, err := glfw.CreateWindow(64, 64, "material FBO", nil, nil)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	window.MakeContextCurrent()
	lookups := 0
	if err := gl.InitWithProcAddrFunc(func(name string) unsafe.Pointer { lookups++; return glfw.GetProcAddress(name) }); err != nil {
		panic(err)
	}
	context, err := capnative.Query(glfw.GetProcAddress)
	if err != nil {
		panic(err)
	}
	path := context.Framebuffers(gl.IsFunctionAvailable)
	if path == capabilities.NoFramebuffer || !context.SeparateBlend(gl.IsFunctionAvailable) {
		panic("material support not verified")
	}
	vertex := shader(gl.VERTEX_SHADER, "#version 120\nattribute vec2 position; void main(){ gl_Position=vec4(position,0.0,1.0); }")
	fragment := shader(gl.FRAGMENT_SHADER, "#version 120\nuniform float alpha; void main(){ gl_FragColor=vec4(1.0,0.0,0.0,alpha); }")
	defer gl.DeleteShader(vertex)
	defer gl.DeleteShader(fragment)
	program := gl.CreateProgram()
	defer gl.DeleteProgram(program)
	gl.AttachShader(program, vertex)
	gl.AttachShader(program, fragment)
	gl.LinkProgram(program)
	var linked int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &linked)
	if linked != gl.TRUE {
		panic("material program failed")
	}
	gl.UseProgram(program)
	points := []float32{-1, -1, 1, -1, -1, 1, 1, 1}
	var buffer uint32
	gl.GenBuffers(1, &buffer)
	defer gl.DeleteBuffers(1, &buffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer)
	gl.BufferData(gl.ARRAY_BUFFER, len(points)*4, gl.SlicePtr(points), gl.STATIC_DRAW)
	position := uint32(gl.GetAttribLocation(program, gl.Str("position\x00")))
	gl.EnableVertexAttribArray(position)
	gl.VertexAttribPointerWithOffset(position, 2, gl.FLOAT, false, 0, 0)
	paths := map[string][][4]byte{"selected": checkTarget(path, program)}
	if context.HasExtension("GL_EXT_framebuffer_object") && gl.IsFunctionAvailable("glGenFramebuffersEXT") {
		paths["EXT"] = checkTarget(capabilities.EXTFramebuffer, program)
	}
	json.NewEncoder(os.Stdout).Encode(struct {
		Version, Renderer                 string
		API                               capabilities.API
		Major, Minor, Extensions, Lookups int
		Path                              capabilities.FramebufferAPI
		Geometry                          glfw.SurfaceGeometry
		Pixels                            map[string][][4]byte
	}{context.Version, context.Renderer, context.API, context.Major, context.Minor, len(context.Extensions()), lookups, path, window.GetSurfaceGeometry(), paths})
}
