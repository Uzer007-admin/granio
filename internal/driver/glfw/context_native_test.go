//go:build !no_glfw && !mobile && !wasm && !test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNestedNativeContexts(t *testing.T) {
	a, b := createWindow("context A"), createWindow("context B")
	defer a.Close()
	defer b.Close()
	valid := true
	runOnMain(func() {
		a.window.RunWithContext(func() {
			valid = valid && glfw.GetCurrentContext() == a.viewport
			b.window.RunWithContext(func() { valid = valid && glfw.GetCurrentContext() == b.viewport })
			valid = valid && glfw.GetCurrentContext() == a.viewport
			func() {
				defer func() { _ = recover() }()
				b.window.RunWithContext(func() { panic("test context restoration") })
			}()
			valid = valid && glfw.GetCurrentContext() == a.viewport
		})
		valid = valid && glfw.GetCurrentContext() == nil
	})
	require.True(t, valid, "inner render and panic must restore the outer native context")
}
