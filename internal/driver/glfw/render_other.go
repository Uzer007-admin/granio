//go:build !windows || wasm || test_web_driver

package glfw

func (*window) preparePresentationWindow(bool, bool) {}
func (*window) resetPresentationWindowHint()         {}
func (*window) retryPresentationWindow(bool) bool    { return false }

func (*window) preparePresentationFirstFrame() {}

func (*window) preparePresentation()                {}
func (*window) initPresentation()                   {}
func (w *window) opaqueRendering() bool             { return w.softwareRendering() }
func (*window) presentationReady() bool             { return true }
func (*window) presentationFailure() (uint32, bool) { return 0, false }
func (*window) requestPresentationCheck()           {}

func (*window) presentationMaxTextureSize() int { return 0 }
