//go:build !wasm && !test_web_driver && !mobile

package glfw

import (
	"bytes"
	"log"
	"testing"

	"github.com/Uzer007-admin/granio/v2"

	"github.com/stretchr/testify/assert"
)

func TestWindow_SizeLimitWarning(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previous) })

	for _, size := range []granio.Size{granio.NewSize(100_000, 100), granio.NewSize(100, 100_000)} {
		w := &window{canvas: &glCanvas{scale: 1, texScale: 1}}
		output.Reset()
		w.screenSize(granio.NewSize(16384, 16384))
		assert.Empty(t, output.String())
		w.screenSize(size)
		assert.Contains(t, output.String(), "wrapping, truncation, or scrolling")
		output.Reset()
		w.screenSize(size)
		assert.Empty(t, output.String())
	}
}
