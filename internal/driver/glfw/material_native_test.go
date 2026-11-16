//go:build !no_glfw && !wasm && !test_web_driver && !mobile

package glfw

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/container"
	"github.com/Uzer007-admin/granio/v3/widget"
	"github.com/stretchr/testify/require"
)

func TestNativeCanvasAndPopupMaterials(t *testing.T) {
	if os.Getenv("FYNE_TEST_MATERIAL") != "1" {
		t.Skip("native material FBO acceptance; unsupported devices use opaque fallback")
	}
	w := createWindowWithResizeCallback("Canvas material integration")
	defer w.Close()
	base := canvas.NewRasterWithPixels(func(x, y, width, height int) color.Color {
		return color.NRGBA{R: uint8(x * 251 / max(1, width)), G: uint8(y * 239 / max(1, height)), B: 31, A: 255}
	})
	m := canvas.NewMaterial(color.NRGBA{R: 200, G: 10, B: 20, A: 255})
	m.BlurRadius = 0
	m.CornerRadius = 0
	m.Tint = color.NRGBA{}
	m.Downsample = 1
	m.SoftwareBlur = true
	var reference, filtered image.Image
	var stats map[string]uint64
	runOnMain(func() { w.window.SetContent(base); w.window.Resize(granio.NewSize(360, 240)); w.window.Show() })
	require.Eventually(t, func() bool {
		ready := false
		runOnMain(func() { ready = w.canvas.Size().Width > 300 && w.canvas.Size().Height > 200 })
		return ready
	}, time.Second, 10*time.Millisecond)
	runOnMain(func() {
		reference = w.canvas.Capture()
		m.Resize(granio.NewSize(160, 100))
		m.Move(granio.NewPos(40, 50))
		overlay := container.NewWithoutLayout(m)
		w.canvas.Overlays().Add(overlay)
		filtered = w.canvas.Capture()
		// A zero radius is a tint-only operation. Nonzero radius must run
		// the filter once and reuse the result on the next frame.
		m.BlurRadius = 9
		m.Refresh()
		w.canvas.Capture()
		w.canvas.Capture()
		stats = w.canvas.Painter().(interface{ MaterialStatistics() map[string]uint64 }).MaterialStatistics()
	})
	require.NotEmpty(t, stats, "native GPU pipeline must run, not silently fall back")
	require.Equal(t, uint64(1), stats["captures"])
	require.Greater(t, stats["cache_hits"], uint64(0))
	require.LessOrEqual(t, stats["pool_bytes"]+stats["capture_bytes"], stats["budget_bytes"])
	for _, pos := range []image.Point{{60, 70}, {160, 120}, {100, 95}} {
		require.Equal(t, reference.At(pos.X, pos.Y), filtered.At(pos.X, pos.Y), "GPU snapshot orientation and lower-layer composition at %v", pos)
	}
	var fallback image.Image
	runOnMain(func() { m.Enabled = false; m.Refresh(); fallback = w.canvas.Capture() })
	r, g, b, a := fallback.At(100, 95).RGBA()
	require.Equal(t, []uint32{200 * 257, 10 * 257, 20 * 257, 65535}, []uint32{r, g, b, a})
	var popup *widget.PopUp
	runOnMain(func() {
		for _, overlay := range w.canvas.Overlays().List() {
			w.canvas.Overlays().Remove(overlay)
		}
		popup = widget.NewPopUp(widget.NewLabel("Material popup"), w.canvas)
		popup.Material = canvas.NewMaterial(color.NRGBA{A: 255})
		popup.ShowAtPosition(granio.NewPos(30, 20))
		popup.Resize(granio.NewSize(180, 80))
		w.canvas.Capture()
		popup.Hide()
		w.canvas.Capture()
	})
}

func TestNativeLiveMaterialUpdatesAndKeepsParentClip(t *testing.T) {
	if os.Getenv("FYNE_TEST_MATERIAL") != "1" {
		t.Skip("native material FBO acceptance")
	}
	w := createWindowWithResizeCallback("Live glass and clipping")
	defer w.Close()
	base := canvas.NewRectangle(color.NRGBA{R: 24, G: 24, B: 24, A: 255})
	source := canvas.NewRectangle(color.NRGBA{R: 240, A: 255})
	m := canvas.NewMaterial(color.NRGBA{G: 240, A: 255})
	m.Live, m.Downsample, m.BlurRadius, m.CornerRadius = true, 1, 6, 18
	m.SoftwareBlur = true
	m.Tint = color.NRGBA{A: 84}
	clip := container.NewClip(container.NewWithoutLayout(m))
	var before, coarse, after image.Image
	runOnMain(func() {
		base.Resize(granio.NewSize(250, 180))
		source.Resize(granio.NewSize(30, 24))
		source.Move(granio.NewPos(72, 72))
		m.Resize(granio.NewSize(120, 80))
		clip.Move(granio.NewPos(40, 50))
		clip.Resize(granio.NewSize(90, 64))
		w.window.SetPadded(false)
		w.window.SetContent(container.NewWithoutLayout(base, source, clip))
		w.window.Resize(granio.NewSize(250, 180))
		w.window.Show()
		before = w.canvas.Capture()
		m.Downsample = 4
		m.Refresh()
		coarse = w.canvas.Capture()
		// Move without discarding the texture or changing the material geometry.
		source.Move(granio.NewPos(138, 72))
		after = w.canvas.Capture()
	})
	require.NotNil(t, before)
	require.NotNil(t, coarse)
	require.NotNil(t, after)
	scale := float64(before.Bounds().Dx()) / 250
	at := func(frame image.Image, x, y float64) color.NRGBA {
		return color.NRGBAModel.Convert(frame.At(int(x*scale), int(y*scale))).(color.NRGBA)
	}
	for y := 50; y < 67; y++ {
		for x := 40; x < 57; x++ {
			require.InDelta(t, at(before, float64(x), float64(y)).R, at(coarse, float64(x), float64(y)).R, 2,
				"coverage at %d,%d must stay smooth when lowering blur resolution", x, y)
		}
	}
	require.Greater(t, at(before, 85, 83).R, uint8(110))
	require.Less(t, at(after, 85, 83).R, uint8(55))
	require.Equal(t, color.NRGBA{R: 240, A: 255}, at(after, 139, 83), "pixels outside parent clip must retain the source color")
}

func TestNativeMaterialOptics(t *testing.T) {
	if os.Getenv("FYNE_TEST_MATERIAL") != "1" {
		t.Skip("проверка оптики внутри окна")
	}
	w := createWindowWithResizeCallback("Проверка прозрачности и размытия")
	defer w.Close()
	base := canvas.NewRectangle(color.Black)
	foreground := canvas.NewRectangle(color.White)
	m := canvas.NewMaterial(color.NRGBA{G: 120, A: 255})
	m.Tint, m.Downsample = color.NRGBA{G: 120, A: 255}, 4
	m.TrackBackdrop, m.SoftwareBlur, m.Priority, m.TranslucentFallback = true, true, true, true
	m.SetCornerRadii(&granio.CornerRadii{TopLeft: 24, BottomRight: 20})
	runOnMain(func() {
		base.Resize(granio.NewSize(240, 150))
		foreground.Resize(granio.NewSize(96, 150))
		m.Move(granio.NewPos(16, 15))
		m.Resize(granio.NewSize(198, 110))
		w.window.SetPadded(false)
		w.window.SetContent(container.NewWithoutLayout(base, foreground, m))
		w.window.Resize(granio.NewSize(240, 150))
		w.window.Show()
	})
	ensureCanvasSize(t, w, granio.NewSize(240, 150))
	capture := func(transparency, blur float32) *image.NRGBA {
		var frame image.Image
		var err error
		runOnMain(func() {
			err = m.ApplyOptics(transparency, blur)
			frame = w.canvas.Capture()
		})
		require.NoError(t, err)
		require.NotNil(t, frame)
		out := image.NewNRGBA(frame.Bounds())
		draw.Draw(out, out.Bounds(), frame, frame.Bounds().Min, draw.Src)
		return out
	}
	flat := capture(0, 50)
	scale := float64(flat.Bounds().Dx()) / 240
	at := func(frame *image.NRGBA, x, y int) color.NRGBA {
		return frame.NRGBAAt(int((float64(x)+.1)*scale), int((float64(y)+.1)*scale))
	}
	// Sample past the edge's antialiasing band, within the space a
	// rounded corner would remove.
	require.Equal(t, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, at(flat, 19, 18))
	require.Equal(t, color.NRGBA{G: 120, A: 255}, at(flat, 210, 18))
	sharp := capture(1, 0)
	require.Equal(t, uint8(255), at(sharp, 94, 70).R)
	require.Zero(t, at(sharp, 98, 70).R)
	width := func(frame *image.NRGBA) int {
		n := 0
		for x := 35; x < 160; x++ {
			if r := at(frame, x, 70).R; r > 20 && r < 235 {
				n++
			}
		}
		return n
	}
	small, large := capture(1, 10), capture(1, 42)
	require.Positive(t, width(small))
	require.Greater(t, width(large), width(small))
	require.Contains(t, []canvas.MaterialEffect{canvas.MaterialHardwareBlur, canvas.MaterialSoftwareBlur}, m.Effect())
	half := capture(.5, 42)
	require.InDelta(t, 60, at(half, 180, 70).G, 2)
	t.Logf("blur edge widths: radius 10 = %d dp, radius 42 = %d dp, effect=%v", width(small), width(large), m.Effect())
}

func TestNativeMaterialShortProfile(t *testing.T) {
	if os.Getenv("FYNE_TEST_MATERIAL_PROFILE") != "1" {
		t.Skip("короткое сравнение кэша материала в живом окне")
	}
	w := createWindowWithResizeCallback("Измерение материала на геометрическом образце")
	defer w.Close()
	base := canvas.NewRectangle(color.NRGBA{R: 24, G: 24, B: 24, A: 255})
	source := canvas.NewRectangle(color.NRGBA{R: 210, G: 145, A: 255})
	m := canvas.NewMaterial(color.NRGBA{A: 255})
	m.Tint, m.BlurRadius, m.Downsample = color.NRGBA{R: 56, G: 56, B: 56, A: 55}, 22, 4
	if os.Getenv("FYNE_TEST_MATERIAL_FULL_RES") == "1" {
		m.Downsample = 1
	}
	m.Priority, m.SoftwareBlur = true, true
	m.SetCornerRadii(&granio.CornerRadii{TopRight: 20, BottomLeft: 22, BottomRight: 10})
	var pipeline interface{ MaterialStatistics() map[string]uint64 }
	runOnMain(func() {
		base.Resize(granio.NewSize(640, 440))
		source.Resize(granio.NewSize(210, 230))
		source.Move(granio.NewPos(40, 40))
		m.Resize(granio.NewSize(380, 100))
		m.Move(granio.NewPos(35, 90))
		w.window.SetPadded(false)
		w.window.SetContent(container.NewWithoutLayout(base, source, m))
		w.window.Resize(granio.NewSize(480, 320))
		w.window.Show()
		w.canvas.Capture()
		pipeline, _ = w.canvas.Painter().(interface{ MaterialStatistics() map[string]uint64 })
	})
	require.Eventually(t, func() bool {
		ready := false
		runOnMain(func() { ready = !w.canvas.Size().IsZero() && w.canvas.Size().Width >= 480 })
		return ready
	}, time.Second, 10*time.Millisecond)
	statistics := func() (stats map[string]uint64) {
		if pipeline != nil {
			runOnMain(func() { stats = pipeline.MaterialStatistics() })
		}
		return
	}
	p95 := func(samples []time.Duration) time.Duration {
		slices.Sort(samples)
		return samples[(len(samples)*95+99)/100-1]
	}
	for run := range 3 {
		var liveCaptures uint64
		for _, tracked := range []bool{false, true} {
			runOnMain(func() {
				m.Live, m.TrackBackdrop = !tracked, tracked
				source.Move(granio.NewPos(40, 40))
				m.Resize(granio.NewSize(380, 100))
				m.InvalidateBackdrop()
				w.RunWithContext(func() { d.repaintWindow(w.window) })
			})
			capturesBefore := statistics()["captures"]
			var before, after runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&before)
			var queued, frame [40]time.Duration
			for i := range queued {
				posted := time.Now()
				runOnMain(func() {
					queued[i] = time.Since(posted)
					if i >= 20 && i < 30 {
						source.Move(granio.NewPos(40+float32(i), 40))
						m.InvalidateBackdrop()
					}
					if i == 30 {
						m.Resize(granio.NewSize(395, 107))
					}
					started := time.Now()
					w.RunWithContext(func() { d.repaintWindow(w.window) })
					frame[i] = time.Since(started)
				})
			}
			runtime.ReadMemStats(&after)
			stats := statistics()
			captures := stats["captures"] - capturesBefore
			t.Logf("run=%d tracked=%v effect=%v queue_p95=%s frame_p95=%s mallocs=%d total_B=%d heap_B=%d material_B=%d captures=%d",
				run+1, tracked, m.Effect(), p95(queued[:]), p95(frame[:]), after.Mallocs-before.Mallocs, after.TotalAlloc-before.TotalAlloc,
				after.HeapAlloc, stats["pool_bytes"]+stats["capture_bytes"], captures)
			if !tracked {
				liveCaptures = captures
			} else if pipeline != nil && liveCaptures >= 40 {
				require.Less(t, captures, liveCaptures)
			}
		}
	}
}
