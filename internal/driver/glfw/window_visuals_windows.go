//go:build !wasm && !test_web_driver

package glfw

import (
	"errors"
	"sync"
	"time"
	"unsafe"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/driver/desktop"
	"github.com/Uzer007-admin/granio/v3/theme"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var setWindowCompositionAttribute = windows.NewLazySystemDLL("user32.dll").NewProc("SetWindowCompositionAttribute")
var getSystemPowerStatus = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetSystemPowerStatus")
var errClearAcrylicAccent = errors.New("failed to remove the Windows Acrylic accent")

const (
	acrylicBackdropWorkaroundBuild = 26300
	acrylicBackdropRefreshRevision = 9974
	acrylicBackdropRefreshInterval = 66 * time.Millisecond
	windowAccentPolicy             = 19
	accentAcrylicBlurBehind        = 4
	accentModernAcrylic            = 1 << 1
)

// The additional DWM invalidation is needed on this exact revision. The
// caller has already checked the base build before reading the registry.
var windowsAcrylicRefreshRequired = sync.OnceValue(func() bool {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()
	revision, _, err := key.GetIntegerValue("UBR")
	return err == nil && revision == acrylicBackdropRefreshRevision
})

// Owned by the UI thread, like the HWND and its presentation resources.
type acrylicBackdropRefresh struct {
	requestID, failedRequestID uint64
	next                       time.Time
	dark                       bool
}

type windowAccent struct {
	state, flags, color, animation uint32
}

type windowCompositionAttribute struct {
	attribute uint32
	value     *windowAccent
	size      uintptr
}

// SYSTEM_POWER_STATUS from winbase.h.
type windowsPowerStatus struct {
	acLine, battery, charge, systemStatus uint8
	batteryLife, batteryFullLife          uint32
}

var windowsAcrylicPowerAllowed = func() bool {
	var status windowsPowerStatus
	ok, _, _ := getSystemPowerStatus.Call(uintptr(unsafe.Pointer(&status)))
	return ok != 0 && status.systemStatus == 0
}

// The accent policy requests the older system Acrylic effect for one HWND.
// It is undocumented and limited to the Windows build where it was introduced
// to work around a stale system backdrop.
func (w *window) setAcrylicAccent(enabled, dark bool) bool {
	if setWindowCompositionAttribute.Find() != nil {
		return false
	}
	var accent windowAccent
	if enabled {
		accent.state = accentAcrylicBlurBehind
		accent.color = 0x6bffffff // Alpha and BGR, 42% white tint.
		if dark {
			accent.color = 0x99202020 // 60% dark tint.
			// The legacy dark recipe can start with a solid tint even when
			// the DirectComposition target already has a full alpha frame.
			// The caller limits AccentPolicy to build 26300 with composition.
			if windowsAcrylicRefreshRequired() {
				accent.flags = accentModernAcrylic
			}
		}
	}
	data := windowCompositionAttribute{windowAccentPolicy, &accent, unsafe.Sizeof(accent)}
	ok, _, _ := setWindowCompositionAttribute.Call(uintptr(unsafe.Pointer(w.viewport.GetWin32Window())), uintptr(unsafe.Pointer(&data)))
	if ok == 0 {
		return false
	}
	w.presentation.acrylicAccent = enabled
	return true
}

func (w *window) clearAcrylicAccent() bool {
	w.presentation.acrylicRefresh.requestID = 0
	return !w.presentation.acrylicAccent || w.setAcrylicAccent(false, false)
}

// The documented DWM material can enforce system policy when the per-window
// Acrylic accent is unavailable or should not be used.
func windowsBackdropMaterial(backdrop desktop.WindowBackdrop, policy windowVisualPolicy, composition bool) (glfw.SystemBackdrop, bool) {
	switch backdrop {
	case desktop.BackdropDesktopAcrylic:
		return glfw.BackdropTransientWindow, composition && policy.build == acrylicBackdropWorkaroundBuild && policy.accentPowerAllowed
	case desktop.BackdropSystemBlur:
		return glfw.BackdropTransientWindow, false
	case desktop.BackdropMicaAlt:
		return glfw.BackdropTabbedWindow, false
	case desktop.BackdropTransparent, desktop.BackdropOpaque:
		return glfw.BackdropNone, false
	default:
		return glfw.BackdropMainWindow, false
	}
}

func windowVisualAlphaFormat() bool { return true }

type windowVisualPolicy struct {
	composition, highContrast, highContrastKnown, remote, transparent, transparencyKnown bool
	build                                                                                uint32
	accentPowerAllowed                                                                   bool
}

func (w *window) queryWindowVisualPolicy() (windowVisualPolicy, glfw.WindowVisualState, error) {
	s, err := w.viewport.QueryWindowVisuals()
	p := windowVisualPolicy{composition: s.Composition, highContrast: s.HighContrast, highContrastKnown: s.HighContrastKnown, remote: s.RemoteSession, build: s.OSBuild}
	if err == nil && p.composition && p.build == acrylicBackdropWorkaroundBuild {
		p.accentPowerAllowed = windowsAcrylicPowerAllowed()
	}
	key, keyErr := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if keyErr == nil {
		value, _, valueErr := key.GetIntegerValue("EnableTransparency")
		key.Close()
		if valueErr == nil {
			p.transparent = value != 0
			p.transparencyKnown = true
		}
	}
	if settings, ok := granio.CurrentApp().Settings().(granio.SystemAppearanceProvider); ok {
		snapshot := settings.SystemAppearance()
		if snapshot.HighContrast.Available {
			p.highContrast, p.highContrastKnown = p.highContrast || snapshot.HighContrast.Value, true
		}
		if snapshot.Transparency.Available {
			p.transparent = (!p.transparencyKnown || p.transparent) && snapshot.Transparency.Value
			p.transparencyKnown = true
		}
	}
	if theme, ok := granio.CurrentApp().Settings().Theme().(interface{ HighContrast() bool }); ok && theme.HighContrast() {
		p.highContrast = true
		p.highContrastKnown = true
	}
	return p, s, err
}

func (w *window) applyNativeWindowBackdrop(o desktop.VisualOptions, requestID uint64, caps *desktop.VisualCapabilities) (desktop.VisualFallbackReason, bool, uint32) {
	refresh := &w.presentation.acrylicRefresh
	refresh.requestID = 0
	presenter := w.presentation.dxgi
	composition := presenter != nil && presenter.composition != nil
	if w.opaqueRendering() || w.presentation.alphaUnavailable || presenter != nil && !composition {
		caps.Backend = desktop.VisualBackendOpaque
		caps.TransparentSurface, caps.AlphaSurface = false, false
		caps.OSBackdrops = 1 << desktop.BackdropOpaque
		if o.Backdrop != desktop.BackdropOpaque {
			return desktop.VisualUnsupportedBackend, false, 0
		}
		return desktop.VisualReasonNone, false, 0
	}
	policy, state, err := w.queryWindowVisualPolicy()
	w.visuals.lastPolicy = policy
	caps.OSKnown = err == nil
	caps.OSBuild = state.OSBuild
	caps.OSBackdrops = 1 << desktop.BackdropOpaque
	caps.Backend = desktop.VisualBackendWGL
	caps.TransparentSurface = state.WGLAlphaAvailable()
	caps.AlphaSurface = state.WGLAlphaAvailable() && caps.AlphaBits > 0
	if composition {
		caps.Backend = desktop.VisualBackendDirectComposition
		caps.AlphaSurface = caps.AlphaBits > 0
		caps.TransparentSurface = state.Composition && caps.AlphaSurface
	}
	caps.CompositorAvailable = state.Composition
	caps.OSBackdrops |= 1 << desktop.BackdropTransparent
	if state.OSAttributes&glfw.VisualBackdrop != 0 {
		caps.OSBackdrops |= 0xf | 1<<desktop.BackdropSystemBlur
	}
	if reason := windowBackdropRejection(o, policy, caps, err); reason != desktop.VisualReasonNone {
		return reason, false, 0
	}
	if o.Backdrop == desktop.BackdropOpaque {
		return desktop.VisualReasonNone, false, 0
	}
	backdrop, useAccent := windowsBackdropMaterial(o.Backdrop, policy, composition)
	if useAccent && refresh.failedRequestID == requestID {
		return desktop.VisualNativeFailure, false, 0
	}
	attributes := state.OSAttributes & (glfw.VisualBackdrop | glfw.VisualDarkCaption)
	// Older Windows uses GLFW's documented transparent framebuffer path. The
	// redirection alpha attribute is additional and requested only when supported.
	if !composition && state.OSAttributes&glfw.VisualPremultipliedAlpha != 0 {
		attributes |= glfw.VisualPremultipliedAlpha
	}
	native := glfw.WindowVisualOptions{
		Attributes: attributes, Backdrop: backdrop,
		DarkCaption: granio.CurrentApp().Settings().ThemeVariant() == theme.VariantDark, PremultipliedAlpha: true,
	}
	if useAccent && setWindowCompositionAttribute.Find() == nil {
		// Both effects are composited behind the same client surface. Remove
		// the transient backdrop before enabling the live Acrylic policy.
		accent := native
		accent.Backdrop = glfw.BackdropNone
		if _, err := w.viewport.ApplyWindowVisuals(accent); err == nil && w.setAcrylicAccent(true, native.DarkCaption) {
			if windowsAcrylicRefreshRequired() {
				refresh.requestID = requestID
				refresh.dark = native.DarkCaption
				refresh.next = time.Now().Add(acrylicBackdropRefreshInterval)
			}
			return desktop.VisualReasonNone, true, 0
		}
	}
	if !w.clearAcrylicAccent() {
		return desktop.VisualNativeFailure, false, 0
	}
	result, applyErr := w.viewport.ApplyWindowVisuals(native)
	if applyErr != nil {
		return desktop.VisualNativeFailure, false, windowsVisualErrorCode(result)
	}
	return desktop.VisualReasonNone, true, 0
}

func windowsVisualErrorCode(result glfw.WindowVisualResult) uint32 {
	for _, code := range result.HRESULT {
		if code != 0 {
			return code
		}
	}
	return 0
}

// Resolve restrictions before issuing any native attributes. It is shared by
// runtime application and policy tests; no test needs to change OS settings.
func windowBackdropRejection(o desktop.VisualOptions, policy windowVisualPolicy, caps *desktop.VisualCapabilities, err error) desktop.VisualFallbackReason {
	if o.Backdrop == desktop.BackdropOpaque {
		return desktop.VisualReasonNone
	}
	if err != nil || !caps.OSBackdrops.Supports(o.Backdrop) {
		return desktop.VisualUnsupportedOS
	}
	if policy.highContrast {
		return desktop.VisualHighContrast
	}
	if policy.transparencyKnown && !policy.transparent {
		return desktop.VisualTransparencyDisabled
	}
	if policy.remote {
		return desktop.VisualRemoteSession
	}
	if !policy.composition {
		return desktop.VisualUnsupportedBackend
	}
	if !caps.RendererBackdrops.Supports(o.Backdrop) {
		return desktop.VisualUnsupportedBackend
	}
	if !caps.TransparentSurface {
		return desktop.VisualSurfaceNotTransparent
	}
	return desktop.VisualReasonNone
}

func (w *window) clearNativeWindowBackdrop() (uint32, error) {
	if !w.clearAcrylicAccent() {
		return 0, errClearAcrylicAccent
	}
	state, err := w.viewport.QueryWindowVisuals()
	if err != nil {
		return 0, err
	}
	attributes := state.OSAttributes & (glfw.VisualBackdrop | glfw.VisualPremultipliedAlpha)
	if attributes == 0 {
		return 0, nil
	}
	result, err := w.viewport.ApplyWindowVisuals(glfw.WindowVisualOptions{Attributes: attributes, Backdrop: glfw.BackdropNone})
	return windowsVisualErrorCode(result), err
}

func (w *window) refreshVisualPolicy() time.Time {
	// Ordinary windows pay only an atomic load, not a mutex or system query.
	if !w.visuals.active.Load() || w.viewport == nil || w.closing {
		return time.Time{}
	}
	now := time.Now()
	if !now.Before(w.visuals.policyNext) {
		w.visuals.policyNext = now.Add(500 * time.Millisecond)
		policy, _, err := w.queryWindowVisualPolicy()
		if err == nil && policy != w.visuals.lastPolicy {
			w.applyWindowVisuals()
		}
	}
	return w.refreshAcrylicBackdrop(now)
}

func (w *window) refreshAcrylicBackdrop(now time.Time) time.Time {
	p := &w.presentation
	r := &p.acrylicRefresh
	if r.requestID == 0 || !w.visible || p.dxgi == nil || p.dxgi.composition == nil || !w.presentationReady() {
		return time.Time{}
	}
	if now.Before(r.next) {
		return r.next
	}
	if w.viewport.GetAttrib(glfw.Iconified) == glfw.True {
		return time.Time{}
	}
	_, status, enabled := w.visualRequest()
	if !enabled || status.RequestID != r.requestID || status.Phase != desktop.VisualReady || !status.SurfaceReady {
		return time.Time{}
	}
	// On 26300.9974 DWM can retain its previous backdrop despite accepting a
	// new source frame. Replacing the accent invalidates that cached sample.
	// Do both writes together without a canvas repaint or an extra Present.
	if w.setAcrylicAccent(false, false) && w.setAcrylicAccent(true, r.dark) {
		r.next = now.Add(acrylicBackdropRefreshInterval)
		return r.next
	}
	r.failedRequestID = r.requestID
	r.requestID = 0
	w.applyWindowVisuals()
	return time.Time{}
}
