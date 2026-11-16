//go:build !windows || wasm || test_web_driver

package glfw

func (*window) updateClientTitleBar() {}

func (*window) changeTitleBarBounds(int, int, bool) bool { return false }

func (*window) setTitleBarLimits(int, int, int, int) bool { return false }
