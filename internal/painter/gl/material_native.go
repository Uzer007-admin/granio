//go:build !android && !ios && !mobile && !wasm && !test_web_driver

package gl

import (
	"math"
	"strings"
	"time"

	"github.com/Uzer007-admin/gl/capabilities"
	capnative "github.com/Uzer007-admin/gl/capabilities/native"
	"github.com/Uzer007-admin/gl/material"
	es "github.com/Uzer007-admin/gl/v2.0/gles2"
	core "github.com/Uzer007-admin/gl/v2.1/material"
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/internal"
	paint "github.com/Uzer007-admin/granio/v3/internal/painter"
	"github.com/Uzer007-admin/granio/v3/theme"
)

type nativeMaterialPipeline struct {
	ctx           context
	es            bool
	software      bool
	mask          programState
	device        *material.GLDevice
	renderer      *material.Renderer
	source        material.Target
	slot          int
	keys          [material.MaxLayers]material.Key
	owners        [material.MaxLayers]*canvas.Material
	valid         [material.MaxLayers]bool
	quality       [material.MaxLayers]uint8
	serial        uint64
	ownerSerial   [material.MaxLayers]uint64
	frame         uint64
	captureLimit  uint64
	captures      uint64
	snapshot      materialSnapshotKey
	regions       [material.MaxLayers][4]int
	budget        paint.SoftwareBlurBudget
	priorityDrawn bool
}

type materialSnapshotKey struct {
	owner            *canvas.Material
	region           [4]int
	surface, content uint64
	pxScale          float32
}

func newMaterialPipeline(p *painter) materialPipeline {
	context, err := capnative.Query(glfw.GetProcAddress)
	if err != nil {
		return nil
	}
	var api material.DrawProcedures
	if context.API == capabilities.OpenGLES {
		api, err = es.LoadWithProcAddrFunc(glfw.GetProcAddress)
	} else {
		api, err = core.LoadWithProcAddrFunc(glfw.GetProcAddress)
	}
	if err != nil {
		return nil
	}
	device, err := material.NewGLDevice(context, api)
	if err != nil {
		return nil
	}
	// 16 MiB is reserved for the GPU-only lower-layer capture, including its
	// transactional resize. Pool texels and capture together never exceed 64 MiB.
	software := softwareGLRenderer(context.Renderer)
	poolLimit, captureLimit := uint64(48<<20), uint64(16<<20)
	if software {
		captureLimit = 1600000
		poolLimit = (2 << 20) - captureLimit
	}
	renderer, err := material.NewRenderer(context, api, poolLimit)
	if err != nil {
		return nil
	}
	return &nativeMaterialPipeline{
		ctx: p.ctx, es: context.API == capabilities.OpenGLES, software: software,
		device: device, renderer: renderer, captureLimit: captureLimit,
	}
}

func softwareGLRenderer(renderer string) bool {
	renderer = strings.ToLower(renderer)
	return strings.Contains(renderer, "llvmpipe") || strings.Contains(renderer, "softpipe") ||
		strings.Contains(renderer, "swiftshader") || strings.Contains(renderer, "gdi generic") ||
		strings.Contains(renderer, "software rasterizer")
}

func (n *nativeMaterialPipeline) ResetFrame() {
	n.slot = 0
	n.priorityDrawn = false
	n.frame++
	n.budget.ResetFrame()
}
func (n *nativeMaterialPipeline) Close() {
	if n.mask.ref != noProgram {
		n.ctx.DeleteProgram(n.mask.ref)
		n.ctx.DeleteBuffer(n.mask.buff)
	}
	n.renderer.Close()
	if n.source.Texture != 0 {
		n.device.DeleteTarget(n.source.Handle)
	}
	n.source = material.Target{}
}
func (n *nativeMaterialPipeline) Statistics() map[string]uint64 {
	s := n.renderer.Stats()
	source := uint64(n.source.Width) * uint64(n.source.Height) * 4
	return map[string]uint64{"pool_bytes": s.UsedBytes, "capture_bytes": source, "captures": n.captures, "cache_hits": s.CacheHits, "target_allocations": s.TargetAllocations, "budget_bytes": s.BudgetBytes + n.captureLimit}
}
func (n *nativeMaterialPipeline) Draw(p *painter, m *canvas.Material, pos granio.Position, frame granio.Size, clip *internal.ClipItem) bool {
	if n.software && !m.SoftwareBlur {
		return false
	}
	if !n.ensureMask(p) {
		return false
	}
	x, y := int(math.Round(float64(pos.X*p.pixScale))), int(math.Round(float64(pos.Y*p.pixScale)))
	w, h := int(math.Round(float64(m.Size().Width*p.pixScale))), int(math.Round(float64(m.Size().Height*p.pixScale)))
	fw := int(math.Round(float64(frame.Width * p.pixScale)))
	if w <= 0 || h <= 0 || x < 0 || y < 0 || x > fw-w || y > p.fbHeight-h || w > n.device.MaxTextureSize() || h > n.device.MaxTextureSize() {
		return false
	}
	bytes := uint64(w) * uint64(h) * 4
	if bytes > n.captureLimit {
		return false
	}
	quality := m.Downsample
	var started time.Time
	if n.software {
		started = time.Now()
		quality = n.budget.Begin(m, w, h, 8, started)
		if quality == 0 {
			return false
		}
		defer func() { n.budget.End(time.Since(started)) }()
	}
	slot := material.MaxLayers - 1 - n.slot
	if m.Priority {
		if n.priorityDrawn {
			return false
		}
		slot = 0
		n.priorityDrawn = true
	} else {
		if slot < 0 || slot == 0 && n.priorityDrawn {
			return false
		}
		n.slot++
	}
	if n.source.Width < w || n.source.Height < h {
		width, height := max(w, n.source.Width), max(h, n.source.Height)
		bytes = uint64(width) * uint64(height) * 4
		oldBytes := uint64(n.source.Width) * uint64(n.source.Height) * 4
		if oldBytes+bytes > n.captureLimit {
			if !n.software {
				return false
			}
			width, height = w, h
			n.device.DeleteTarget(n.source.Handle)
			n.source = material.Target{}
			n.snapshot = materialSnapshotKey{}
			n.valid = [material.MaxLayers]bool{}
		}
		handle, err := n.device.CreateTarget(width, height)
		if err != nil {
			return false
		}
		if n.source.Texture != 0 {
			n.device.DeleteTarget(n.source.Handle)
		}
		n.source = material.Target{Handle: handle, Width: width, Height: height}
		n.valid = [material.MaxLayers]bool{}
		n.snapshot = materialSnapshotKey{}
	}
	if quality == 0 {
		quality = 2
	}
	bg := theme.Color(theme.ColorNameBackground)
	r, g, b, a := bg.RGBA()
	key := material.Key{
		Surface: uint64(uint32(fw))<<32 | uint64(uint32(p.fbHeight)),
		Content: p.materialGeneration, Style: uint64(uint32(x))<<32 | uint64(uint32(y)),
		ScaleX: p.pixScale, ScaleY: p.pixScale, Radius: m.BlurRadius, Noise: m.Noise,
		Tint:   [4]float32{float32(m.Tint.R) / 255, float32(m.Tint.G) / 255, float32(m.Tint.B) / 255, float32(m.Tint.A) / 255},
		Linear: m.Linear,
	}
	if m.Live {
		key.Content = n.frame
		key.Surface ^= 1 << 63
	} else if m.TrackBackdrop {
		key.Content = m.BackdropRevision()
		key.Surface ^= 1 << 62
	}
	if p.windowSurface == WindowSurfaceBackdrop {
		key.Surface ^= 1 << 61
		t := p.windowBackdropTint
		r, g, b, a = uint32(t.R)*257, uint32(t.G)*257, uint32(t.B)*257, uint32(t.A)*257
	}
	key.Content ^= uint64(r)<<48 | uint64(g)<<32 | uint64(b)<<16 | uint64(a)
	if n.owners[slot] != m {
		n.serial++
		n.owners[slot] = m
		n.ownerSerial[slot] = n.serial
		n.valid[slot] = false
	}
	key.Surface ^= n.ownerSerial[slot]
	p.FlushGlyphs()
	region := [4]int{x, y, w, h}
	if !n.valid[slot] || n.keys[slot] != key || n.quality[slot] != quality || n.regions[slot] != region {
		snapshot := materialSnapshotKey{m, region, key.Surface, key.Content, p.pixScale}
		if n.snapshot != snapshot {
			capture := n.source
			capture.Width, capture.Height = w, h
			n.snapshot = materialSnapshotKey{}
			if err := n.device.CopySnapshot(capture, x, p.fbHeight-y-h); err != nil {
				return false
			}
			n.snapshot = snapshot
			n.captures++
		}
	}
	result, err := n.renderer.Render(slot, n.source, material.Region{Width: w, Height: h}, quality, key)
	if err != nil || result.Rendered {
		// A GPU pass owns the program, vertex input and blend/scissor state.
		p.bindings.reset()
		p.scissorEnabled, p.scissorKnown = false, false
		p.ctx.Enable(0x0BE2)
		if clip != nil {
			p.StartClipping(clip.Rect())
		} else {
			p.ctx.Disable(scissorTest)
		}
	}
	if err != nil {
		n.valid[slot] = false
		return false
	}
	n.keys[slot], n.valid[slot], n.quality[slot], n.regions[slot] = key, true, quality, region
	points, _, _ := p.rectCoords(m.Size(), pos, frame, canvas.ImageFillStretch, 1, 0)
	for i := 3; i < len(points); i += 4 {
		points[i] = 1 - points[i]
	} // framebuffer textures have a bottom-left origin
	mask := n.mask
	corners := m.EffectiveCornerRadii()
	p.useProgram(mask.ref)
	p.updateBuffer(mask.buff, points[:])
	p.UpdateVertexArray(mask, attrVertex, 2, 4, 0)
	p.UpdateVertexArray(mask, attrVertexTextureCoordinates, 2, 4, 2)
	p.SetUniform2f(mask, "extent", float32(w), float32(h))
	p.SetUniform4f(mask, "radii", corners.TopLeft*p.pixScale, corners.TopRight*p.pixScale,
		corners.BottomRight*p.pixScale, corners.BottomLeft*p.pixScale)
	if p.windowSurface == WindowSurfaceBackdrop {
		p.SetUniform1f(mask, "pass", 0)
		p.blendFunc(zero, oneMinusSrcAlpha)
		p.ctx.DrawArrays(triangleStrip, 0, 4)
		p.blendFunc(one, one)
	} else {
		p.blendFunc(one, oneMinusSrcAlpha)
	}
	p.SetUniform1f(mask, "pass", 1)
	p.ctx.ActiveTexture(texture0)
	p.ctx.BindTexture(texture2D, Texture(result.Target.Texture))
	p.ctx.DrawArrays(triangleStrip, 0, 4)
	if n.software {
		m.SetEffect(canvas.MaterialSoftwareBlur)
	} else {
		m.SetEffect(canvas.MaterialHardwareBlur)
	}
	return true
}
