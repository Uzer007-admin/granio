//go:build android || ios || mobile || wasm || test_web_driver

package gl

func newCaptureBackend(*painter) (captureBackend, error) { return nil, nil }
