//go:build graphics_probe

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"testing"
	"unsafe"

	"github.com/Uzer007-admin/gl/capabilities"
	capnative "github.com/Uzer007-admin/gl/capabilities/native"
	"github.com/Uzer007-admin/gl/material"
	es "github.com/Uzer007-admin/gl/v2.0/gles2"
	gl "github.com/Uzer007-admin/gl/v2.1/material"
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
)

type procedures interface {
	material.DrawProcedures
	ClearColor(float32, float32, float32, float32)
	Clear(uint32)
	Enable(uint32)
	Scissor(int32, int32, int32, int32)
	ReadPixels(int32, int32, int32, int32, uint32, uint32, unsafe.Pointer)
	GetFramebufferViewport(bool) (int32, int32, [4]int32)
	GetTargetBindings(bool, bool) (int32, int32, int32, int32)
}

type report struct {
	GPUSnapshot                                                    bool
	NoiseStable                                                    bool
	ShaderSourceAllocs                                             float64
	Version, Renderer                                              string
	Lookups                                                        int
	FramebufferAPI                                                 capabilities.FramebufferAPI
	MaximumPixelError                                              float64
	CacheHits, TargetAllocations, UsedBytes, PeakBytes             uint64
	RenderTargetAllocs                                             float64
	LinearAndEncoded, StateRestored, AlphaContract, ResizeFallback bool
	GPUDownsample, SurfaceTracked                                  bool
	CaptureAllocs, BatchedQueryAllocs                              float64
	BatchedQueryTransitions                                        int64
	CaptureMaximumPixelError                                       float64
	AdaptiveRenderer                                               bool
	AdaptiveCacheAllocs                                            float64
}

type scene struct {
	window  *glfw.Window
	api     procedures
	context capabilities.Context
	device  *material.GLDevice
	pool    *material.Pool
	blur    *material.Blur
	lookups int
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func integer(api procedures, enum uint32) int32 {
	var value int32
	api.GetIntegerv(enum, &value)
	return value
}

func makeScene(gles bool, index int) *scene {
	glfw.DefaultWindowHints()
	glfw.WindowHint(glfw.Visible, glfw.False)
	glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	minor := 1
	if gles {
		minor = 0
		glfw.WindowHint(glfw.ClientAPI, glfw.OpenGLESAPI)
	}
	glfw.WindowHint(glfw.ContextVersionMinor, minor)
	if gles || glfw.GetPlatform() == glfw.PlatformWayland {
		glfw.WindowHint(glfw.ContextCreationAPI, glfw.EGLContextAPI)
	}
	w, err := glfw.CreateWindow(96+index*16, 64, "material runtime", nil, nil)
	must(err)
	w.MakeContextCurrent()
	s := &scene{window: w}
	resolve := func(name string) unsafe.Pointer { s.lookups++; return glfw.GetProcAddress(name) }
	if gles {
		s.api, err = es.LoadWithProcAddrFunc(resolve)
	} else {
		s.api, err = gl.LoadWithProcAddrFunc(resolve)
	}
	must(err)
	s.context, err = capnative.Query(glfw.GetProcAddress)
	must(err)
	s.device, err = material.NewGLDevice(s.context, s.api)
	must(err)
	s.pool, err = material.NewPool(s.device, 32768)
	must(err)
	s.blur, err = material.NewBlur(s.context, s.device, s.api)
	must(err)
	return s
}

func (s *scene) close() {
	s.window.MakeContextCurrent()
	s.blur.Close()
	s.pool.Close()
	s.window.Destroy()
}

func seed(s *scene, target material.Target) error {
	return s.device.RenderTarget(target, func() error {
		s.api.Disable(0x0BE2)
		s.api.Enable(0x0C11)
		colors := [4][4]float32{{0.1, 0.2, 0.3, 0.5}, {0, 0, 0, 0}, {0.8, 0, 0, 1}, {0, 0.6, 0.3, 0.75}}
		for i, color := range colors {
			x, y := i%2*(target.Width/2), i/2*(target.Height/2)
			s.api.Scissor(int32(x), int32(y), int32(target.Width/2), int32(target.Height/2))
			s.api.ClearColor(color[0], color[1], color[2], color[3])
			s.api.Clear(0x4000)
		}
		s.api.Disable(0x0C11)
		return nil
	})
}

func source(width, height int) [][4]float64 {
	colors := [4][4]float64{{0.1, 0.2, 0.3, 0.5}, {0, 0, 0, 0}, {0.8, 0, 0, 1}, {0, 0.6, 0.3, 0.75}}
	data := make([][4]float64, width*height)
	for y := range height {
		for x := range width {
			i := 0
			if x >= width/2 {
				i++
			}
			if y >= height/2 {
				i += 2
			}
			data[y*width+x] = quantize(colors[i])
		}
	}
	return data
}
func quantize(c [4]float64) [4]float64 {
	for i := range c {
		c[i] = math.Round(min(1, max(0, c[i]))*255) / 255
	}
	return c
}
func decode(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}
func encode(c float64) float64 {
	if c <= 0.0031308 {
		return c * 12.92
	}
	return 1.055*math.Pow(c, 1/2.4) - 0.055
}

func sample(data [][4]float64, width, height int, x, y float64, linear bool) [4]float64 {
	x0, y0 := int(math.Floor(x)), int(math.Floor(y))
	fx, fy := x-float64(x0), y-float64(y0)
	var c [4]float64
	for row := range 2 {
		for col := range 2 {
			p := data[min(height-1, max(0, y0+row))*width+min(width-1, max(0, x0+col))]
			if linear {
				for ch := range 3 {
					p[ch] = decode(min(1, max(0, p[ch]/max(p[3], 0.00001)))) * p[3]
				}
			}
			weight := []float64{1 - fx, fx}[col] * []float64{1 - fy, fy}[row]
			for ch := range c {
				c[ch] += weight * p[ch]
			}
		}
	}
	return c
}

func reference(size material.Size, key material.Key) [][4]float64 {
	w, h := size.Width, size.Height
	src := source(w, h)
	back := make([][4]float64, w*h)
	front := make([][4]float64, w*h)
	weights := [5]float64{1, 4, 6, 4, 1}
	for y := range h {
		for x := range w {
			var c [4]float64
			for tap := range 5 {
				p := sample(src, w, h, float64(x)+float64(tap-2)*float64(key.Radius)*float64(key.ScaleX)/2, float64(y), key.Linear)
				for ch := range c {
					c[ch] += p[ch] * weights[tap] / 16
				}
			}
			back[y*w+x] = quantize(c)
		}
	}
	for y := range h {
		for x := range w {
			var c [4]float64
			for tap := range 5 {
				p := sample(back, w, h, float64(x), float64(y)+float64(tap-2)*float64(key.Radius)*float64(key.ScaleY)/2, false)
				for ch := range c {
					c[ch] += p[ch] * weights[tap] / 16
				}
			}
			for ch := range 3 {
				tint := float64(key.Tint[ch])
				if key.Linear {
					tint = decode(tint)
				}
				c[ch] = tint*float64(key.Tint[3]) + c[ch]*(1-float64(key.Tint[3]))
			}
			c[3] = float64(key.Tint[3]) + c[3]*(1-float64(key.Tint[3]))
			if key.Linear {
				for ch := range 3 {
					c[ch] = encode(min(1, max(0, c[ch]/max(c[3], 0.00001)))) * c[3]
				}
			}
			ew, eh := float64(w)/float64(key.ScaleX), float64(h)/float64(key.ScaleY)
			r := min(float64(key.Corners), min(ew, eh)/2)
			qx, qy := math.Abs((float64(x)+0.5)/float64(w)*ew-ew/2)-ew/2+r, math.Abs((float64(y)+0.5)/float64(h)*eh-eh/2)-eh/2+r
			distance := math.Hypot(max(0, qx), max(0, qy)) + min(max(qx, qy), 0) - r
			mask := min(1, max(0, 0.5-distance*float64(min(key.ScaleX, key.ScaleY))))
			for ch := range 3 {
				c[ch] = min(c[3], max(0, c[ch]))
			}
			for ch := range c {
				c[ch] *= mask
			}
			front[y*w+x] = quantize(c)
		}
	}
	return front
}

func validate(s *scene) report {
	s.window.MakeContextCurrent()
	api := s.api
	external, err := s.device.CreateTarget(4, 4)
	must(err)
	defer s.device.DeleteTarget(external)
	copyHandle, err := s.device.CreateTarget(4, 4)
	must(err)
	defer s.device.DeleteTarget(copyHandle)
	from, to := material.Target{Handle: external, Width: 4, Height: 4}, material.Target{Handle: copyHandle, Width: 4, Height: 4}
	must(seed(s, from))
	must(s.device.RenderTarget(from, func() error { return s.device.CopySnapshot(to, 0, 0) }))
	must(s.device.RenderTarget(to, func() error {
		pixels := make([]byte, 64)
		api.ReadPixels(0, 0, 4, 4, 0x1908, 0x1401, unsafe.Pointer(&pixels[0]))
		want := source(4, 4)
		for i, pixel := range want {
			for ch := range pixel {
				if math.Abs(float64(pixels[4*i+ch])-pixel[ch]*255) > 1 {
					panic("GPU snapshot differs from lower layers")
				}
			}
		}
		if err := s.device.CopySnapshot(to, 0, 0); !errors.Is(err, material.ErrInvalid) {
			panic("snapshot feedback allowed")
		}
		return nil
	}))
	// Distinct read/draw bindings and a tiny bound PBO expose incorrect restores
	// and accidental TexImage2D uploads from a pixel-unpack buffer.
	split := s.context.API == capabilities.OpenGL && (s.context.AtLeast(3, 0) || s.context.HasExtension("GL_ARB_framebuffer_object")) || s.context.API == capabilities.OpenGLES && s.context.AtLeast(3, 0)
	unpack := s.context.API == capabilities.OpenGL && s.context.AtLeast(2, 1) || s.context.API == capabilities.OpenGLES && s.context.AtLeast(3, 0)
	if split {
		if f, ok := api.(interface{ BindFramebuffer(uint32, uint32) }); ok {
			f.BindFramebuffer(0x8CA8, external.Framebuffer)
			defer f.BindFramebuffer(0x8CA8, 0)
		}
	}
	var pbo uint32
	if unpack {
		api.GenBuffers(1, &pbo)
		api.BindBuffer(0x88EC, pbo)
		api.BufferData(0x88EC, 1, nil, 0x88E4)
		defer api.DeleteBuffers(1, &pbo)
		defer api.BindBuffer(0x88EC, 0)
	}
	draw, read, texture := integer(api, 0x8CA6), int32(0), integer(api, 0x8069)
	if split {
		read = integer(api, 0x8CAA)
	}
	var viewport [4]int32
	api.GetIntegerv(0x0BA2, &viewport[0])
	size := material.Size{Width: 32, Height: 24, Downsample: 1}
	key := material.Key{Surface: 1, Content: 1, Style: 1, ScaleX: 1.25, ScaleY: 1.5, Radius: 3, Corners: 5, Tint: [4]float32{0.1, 0.2, 0.3, 0.15}}
	renders := 0
	render := func(pair material.Pair) error {
		renders++
		if err := seed(s, pair.Front); err != nil {
			return err
		}
		return s.blur.Apply(pair, size, key)
	}
	r := report{Version: s.context.Version, Renderer: s.context.Renderer, Lookups: s.lookups, FramebufferAPI: s.context.Framebuffers(api.IsFunctionAvailable)}
	r.GPUSnapshot = true
	var tracker glfw.SurfaceTracker
	surface, _, err := s.window.UpdateSurface(&tracker)
	must(err)
	if surface.Drawable || surface.ResourceGeneration != 1 {
		panic("hidden surface was drawable")
	}
	if stable, changes, err := s.window.UpdateSurface(&tracker); err != nil || changes != 0 || stable != surface {
		panic("static surface generation changed")
	}
	if lost := tracker.InvalidateContext(); lost.ResourceGeneration != 2 {
		panic("context invalidation lost")
	}
	r.SurfaceTracked = true
	beforeCgo := runtime.NumCgoCall()
	batchedDraw, batchedRead, batchedViewport := api.GetFramebufferViewport(split)
	r.BatchedQueryTransitions = runtime.NumCgoCall() - beforeCgo
	if batchedDraw != draw || split && batchedRead != read || batchedViewport != viewport || r.BatchedQueryTransitions != 1 {
		panic("batched framebuffer query differs")
	}
	batchedTexture, batchedDraw, batchedRead, batchedUnpack := api.GetTargetBindings(split, unpack)
	if batchedTexture != texture || batchedDraw != draw || split && batchedRead != read || unpack && batchedUnpack != int32(pbo) {
		panic("batched target query differs")
	}
	r.BatchedQueryAllocs = testing.AllocsPerRun(20, func() { api.GetFramebufferViewport(split); api.GetTargetBindings(split, unpack) })
	if r.BatchedQueryAllocs != 0 {
		panic("batched queries allocate")
	}
	r.CaptureMaximumPixelError, r.CaptureAllocs = validateCapture(s)
	r.GPUDownsample = true
	r.AdaptiveCacheAllocs = validateRenderer(s)
	r.AdaptiveRenderer = true
	var cleanPixels []byte
	shader := api.CreateShader(0x8B31)
	defer api.DeleteShader(shader)
	r.ShaderSourceAllocs = testing.AllocsPerRun(20, func() { must(api.ShaderSourceString(shader, "void main(){}")) })
	if r.ShaderSourceAllocs != 0 {
		panic("ShaderSourceString allocates")
	}
	for _, linear := range []bool{false, true} {
		key.Linear = linear
		pair, updated, err := s.pool.Update(0, size, key, render)
		must(err)
		if !updated {
			panic("missing changed pass")
		}
		if integer(api, 0x8CA6) != draw || split && integer(api, 0x8CAA) != read || unpack && integer(api, 0x88EF) != int32(pbo) {
			panic("framebuffer/PBO state changed")
		}
		want := reference(size, key)
		pixels := make([]byte, size.Width*size.Height*4)
		must(s.device.RenderTarget(pair.Front, func() error {
			api.ReadPixels(0, 0, int32(size.Width), int32(size.Height), 0x1908, 0x1401, unsafe.Pointer(&pixels[0]))
			return nil
		}))
		for i, p := range want {
			for ch := range p {
				delta := math.Abs(float64(pixels[4*i+ch]) - p[ch]*255)
				r.MaximumPixelError = max(r.MaximumPixelError, delta)
				if delta > 3 {
					panic(fmt.Sprintf("reference mismatch linear=%v pixel=%d channel=%d got=%d expected=%v", linear, i, ch, pixels[4*i+ch], p[ch]*255))
				}
			}
			for ch := range 3 {
				if pixels[4*i+ch] > pixels[4*i+3] {
					panic("non-premultiplied output")
				}
			}
			if pixels[4*i+3] == 0 && (pixels[4*i] != 0 || pixels[4*i+1] != 0 || pixels[4*i+2] != 0) {
				panic("RGB at zero alpha")
			}
		}
		cleanPixels = pixels
		before := renders
		for range 1000 {
			_, changed, err := s.pool.Update(0, size, key, render)
			must(err)
			if changed {
				panic("static material rendered again")
			}
		}
		if renders != before {
			panic("cache ran shader callback")
		}
		r.RenderTargetAllocs = testing.AllocsPerRun(20, func() { must(s.device.RenderTarget(pair.Front, func() error { return nil })) })
		if r.RenderTargetAllocs != 0 {
			panic("RenderTarget allocates per pass")
		}
		failure := errors.New("draw failed")
		if err := s.device.RenderTarget(pair.Front, func() error { return failure }); err != failure {
			panic("draw error lost")
		}
		func() {
			defer func() {
				if recover() != failure {
					panic("draw panic lost")
				}
			}()
			_ = s.device.RenderTarget(pair.Front, func() error { panic(failure) })
		}()
		if integer(api, 0x8CA6) != draw || split && integer(api, 0x8CAA) != read {
			panic("panic changed framebuffer bindings")
		}
	}
	key.Noise = 0.02
	noisy, changed, err := s.pool.Update(0, size, key, render)
	must(err)
	if !changed {
		panic("noise key was ignored")
	}
	readTarget := func(target material.Target) []byte {
		pixels := make([]byte, target.Width*target.Height*4)
		must(s.device.RenderTarget(target, func() error {
			api.ReadPixels(0, 0, int32(target.Width), int32(target.Height), 0x1908, 0x1401, unsafe.Pointer(&pixels[0]))
			return nil
		}))
		return pixels
	}
	noisePixels := readTarget(noisy.Front)
	if bytes.Equal(noisePixels, cleanPixels) {
		panic("noise changed no pixels")
	}
	for i := 0; i < len(noisePixels); i += 4 {
		for ch := range 3 {
			if noisePixels[i+ch] > noisePixels[i+3] {
				panic("noise broke alpha contract")
			}
		}
	}
	_, changed, err = s.pool.Update(0, size, key, render)
	must(err)
	if changed || !bytes.Equal(noisePixels, readTarget(noisy.Front)) {
		panic("static noise changed on a cached frame")
	}
	r.NoiseStable = true
	var currentViewport [4]int32
	api.GetIntegerv(0x0BA2, &currentViewport[0])
	if currentViewport != viewport {
		panic("viewport not restored")
	}
	// Creation, unlike Blur's owned draw state, preserves texture bindings too.
	api.BindTexture(0x0DE1, uint32(texture))
	_, _, err = s.pool.Update(1, size, key, render)
	must(err)
	if integer(api, 0x8CA6) != draw || split && integer(api, 0x8CAA) != read {
		panic("second layer changed bindings")
	}
	size = material.Size{Width: 72, Height: 40, Downsample: 1}
	key.Content++
	if _, _, err := s.pool.Update(0, size, key, render); !errors.Is(err, material.ErrBudget) {
		panic("resize peak exceeded budget")
	}
	must(s.pool.Release(1))
	_, _, err = s.pool.Update(0, size, key, render)
	must(err)
	stats := s.pool.Stats()
	r.CacheHits, r.TargetAllocations, r.UsedBytes, r.PeakBytes = stats.CacheHits, stats.TargetAllocations, stats.UsedBytes, stats.PeakBytes
	r.LinearAndEncoded, r.StateRestored, r.AlphaContract, r.ResizeFallback = true, true, true, true
	if code := api.GetError(); code != 0 {
		panic(fmt.Sprintf("GL error 0x%x", code))
	}
	return r
}

// Independent CPU reference for the documented four-sample capture filter.
// Odd source region sizes, crop boundaries and factor 1/2/4/8 are exercised.
func validateCapture(s *scene) (float64, float64) {
	handle, err := s.device.CreateTarget(32, 24)
	must(err)
	defer s.device.DeleteTarget(handle)
	input := material.Target{Handle: handle, Width: 32, Height: 24}
	must(seed(s, input))
	data := source(32, 24)
	region := material.Region{X: 3, Y: 2, Width: 25, Height: 19}
	var maximum, allocations float64
	for _, linear := range []bool{false, true} {
		for _, factor := range []int{1, 2, 4, 8} {
			w, h := (region.Width-1)/factor+1, (region.Height-1)/factor+1
			handle, err := s.device.CreateTarget(w, h)
			must(err)
			target := material.Target{Handle: handle, Width: w, Height: h}
			must(s.blur.Capture(target, input, region, linear))
			pixels := make([]byte, w*h*4)
			must(s.device.RenderTarget(target, func() error {
				s.api.ReadPixels(0, 0, int32(w), int32(h), 0x1908, 0x1401, unsafe.Pointer(&pixels[0]))
				return nil
			}))
			stepX, stepY := float64(region.Width)/float64(w), float64(region.Height)/float64(h)
			dx, dy := stepX*.25, stepY*.25
			if stepX == 1 {
				dx = 0
			}
			if stepY == 1 {
				dy = 0
			}
			for y := range h {
				for x := range w {
					centerX, centerY := float64(region.X)+(float64(x)+.5)*stepX-.5, float64(region.Y)+(float64(y)+.5)*stepY-.5
					var expected [4]float64
					for _, offset := range [4][2]float64{{-dx, -dy}, {dx, -dy}, {-dx, dy}, {dx, dy}} {
						p := sample(data, 32, 24, min(float64(region.X+region.Width-1), max(float64(region.X), centerX+offset[0])), min(float64(region.Y+region.Height-1), max(float64(region.Y), centerY+offset[1])), linear)
						for ch := range expected {
							expected[ch] += p[ch] * .25
						}
					}
					if linear {
						for ch := range 3 {
							expected[ch] = encode(min(1, max(0, expected[ch]/max(expected[3], .00001)))) * expected[3]
						}
					}
					expected = quantize(expected)
					for ch := range expected {
						got := pixels[4*(y*w+x)+ch]
						delta := math.Abs(float64(got) - expected[ch]*255)
						maximum = max(maximum, delta)
						if delta > 3 {
							panic(fmt.Sprintf("capture mismatch factor=%d linear=%v x=%d y=%d ch=%d got=%d want=%v", factor, linear, x, y, ch, got, expected[ch]*255))
						}
						if ch < 3 && got > pixels[4*(y*w+x)+3] {
							panic("capture alpha contract")
						}
					}
				}
			}
			allocations = max(allocations, testing.AllocsPerRun(20, func() { must(s.blur.Capture(target, input, region, linear)) }))
			if allocations != 0 {
				panic("capture allocates")
			}
			s.device.DeleteTarget(handle)
		}
	}
	if err := s.blur.Capture(input, input, region, false); !errors.Is(err, material.ErrInvalid) {
		panic("capture feedback allowed")
	}
	if err := s.blur.Capture(input, input, material.Region{X: -1, Y: 0, Width: 1, Height: 1}, false); !errors.Is(err, material.ErrInvalid) {
		panic("invalid capture allowed")
	}
	return maximum, allocations
}

func validateRenderer(s *scene) float64 {
	handle, err := s.device.CreateTarget(32, 24)
	must(err)
	defer s.device.DeleteTarget(handle)
	source := material.Target{Handle: handle, Width: 32, Height: 24}
	must(seed(s, source))
	r, err := material.NewRenderer(s.context, s.api, 1024)
	must(err)
	defer r.Close()
	key := material.Key{Surface: 1, Content: 1, Style: 1, ScaleX: 1.25, ScaleY: 1.5, Radius: 3, Corners: 4, Tint: [4]float32{.1, .2, .3, .1}, Linear: true}
	region := material.Region{X: 3, Y: 2, Width: 25, Height: 19}
	first, err := r.Render(0, source, region, 1, key)
	must(err)
	if !first.Rendered || first.Size.Downsample != 4 {
		panic("adaptive renderer did not reduce quality")
	}
	beforeCgo := runtime.NumCgoCall()
	second, err := r.Render(0, source, region, 1, key)
	must(err)
	if second.Rendered || second.Target != first.Target || runtime.NumCgoCall() != beforeCgo {
		panic("adaptive static cache called GL")
	}
	allocs := testing.AllocsPerRun(20, func() {
		result, err := r.Render(0, source, region, 1, key)
		must(err)
		if result.Rendered {
			panic("adaptive static material rendered")
		}
	})
	if allocs != 0 {
		panic(fmt.Sprintf("adaptive cache allocates: %v", allocs))
	}
	region.X++
	shifted, err := r.Render(0, source, region, 1, key)
	must(err)
	if !shifted.Rendered || shifted.Target != first.Target || r.Stats().TargetAllocations != 2 {
		panic("crop identity not cached correctly")
	}
	_, err = r.Render(1, source, material.Region{Width: 32, Height: 24}, 1, key)
	must(err)
	if r.Stats().PeakBytes > 1024 || r.Stats().UsedBytes != 664 {
		panic("shared renderer texel budget")
	}
	if _, err = r.Render(0, first.Target, material.Region{Width: first.Target.Width, Height: first.Target.Height}, 1, key); !errors.Is(err, material.ErrInvalid) {
		panic("renderer sampled its old output")
	}
	must(r.Release(1))
	must(r.Release(0))
	if r.Stats().UsedBytes != 0 {
		panic("renderer release leaked")
	}
	if _, err = r.Render(0, source, material.Region{Width: 10, Height: 10}, 1, key); err != nil {
		panic(err)
	}
	r.Close()
	r.Close()
	if r.Stats().UsedBytes != 0 {
		panic("renderer close leaked")
	}
	if _, err = r.Render(0, source, region, 1, key); !errors.Is(err, material.ErrClosed) {
		panic("closed renderer accepted render")
	}
	return allocs
}

func main() {
	backend := flag.String("backend", "auto", "auto, x11, wayland")
	gles := flag.Bool("gles2", false, "use strict GLES 2 procedures and GLSL 100")
	flag.Parse()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if *backend == "wayland" {
		glfw.InitHint(glfw.PlatformHint, int(glfw.PlatformWayland))
	} else if *backend == "x11" {
		glfw.InitHint(glfw.PlatformHint, int(glfw.PlatformX11))
	}
	must(glfw.Init())
	defer glfw.Terminate()
	first := makeScene(*gles, 0)
	defer first.close()
	second := makeScene(*gles, 1)
	defer second.close()
	// Replace the legacy global table with failed addresses after both immutable
	// tables were created. Rendering through either table must remain valid.
	if !*gles {
		if err := gl.InitWithProcAddrFunc(func(string) unsafe.Pointer { return nil }); err == nil {
			panic("expected failed global reload")
		}
	} else {
		if err := es.InitWithProcAddrFunc(func(string) unsafe.Pointer { return nil }); err == nil {
			panic("expected failed global reload")
		}
	}
	results := []report{validate(first), validate(second)}
	json.NewEncoder(os.Stdout).Encode(results)
}
