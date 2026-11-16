//go:build !wasm && !test_web_driver && !android && !ios && !mobile

package gl

import (
	"testing"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/stretchr/testify/require"
)

type bindingCounter struct {
	context
	programs, buffers, blends, enables, formats int
}

type clipCounter struct {
	context
	boxes             [][4]int32
	enables, disables int
}

func (c *clipCounter) Scissor(x, y, w, h int32) { c.boxes = append(c.boxes, [4]int32{x, y, w, h}) }
func (c *clipCounter) Enable(uint32)            { c.enables++ }
func (c *clipCounter) Disable(uint32)           { c.disables++ }
func (*clipCounter) GetError() uint32           { return 0 }

func TestScissorAlignsFractionalEdgesAndReusesState(t *testing.T) {
	ctx := &clipCounter{}
	p := &painter{ctx: ctx, fbHeight: 120, pixScale: 1.25}
	start := func() { p.StartClipping(granio.NewPos(3, 5), granio.NewSize(8.6, 20)) }
	start()
	start()
	require.Equal(t, [][4]int32{{4, 89, 11, 25}}, ctx.boxes)
	p.StartClipping(granio.NewPos(5, 6), granio.NewSize(4, 17))
	start()
	require.Equal(t, [][4]int32{{4, 89, 11, 25}, {6, 91, 5, 21}, {4, 89, 11, 25}}, ctx.boxes)
	require.Equal(t, 1, ctx.enables)
	p.StopClipping()
	p.StopClipping()
	start()
	require.Len(t, ctx.boxes, 3)
	require.Equal(t, 2, ctx.enables)
	require.Equal(t, 1, ctx.disables)
}

func (c *bindingCounter) UseProgram(Program)                { c.programs++ }
func (c *bindingCounter) BindBuffer(uint32, Buffer)         { c.buffers++ }
func (c *bindingCounter) BlendFunc(uint32, uint32)          { c.blends++ }
func (c *bindingCounter) EnableVertexAttribArray(Attribute) { c.enables++ }
func (c *bindingCounter) VertexAttribPointerWithOffset(Attribute, int, uint32, bool, int, int) {
	c.formats++
}

func TestDrawBindingsReuseAndRestoreAfterExternalPass(t *testing.T) {
	ctx := &bindingCounter{}
	p := &painter{ctx: ctx}
	bind := func(buffer Buffer) {
		p.bindProgram(8)
		p.bindArrayBuffer(buffer)
		p.blendFunc(one, oneMinusSrcAlpha)
		p.setVertexFormat(0, 2, 16, 0)
	}

	bind(6)
	bind(6)
	require.Equal(t, [5]int{1, 1, 1, 1, 1}, [5]int{ctx.programs, ctx.buffers, ctx.blends, ctx.enables, ctx.formats})
	bind(7)
	require.Equal(t, [5]int{1, 2, 1, 1, 2}, [5]int{ctx.programs, ctx.buffers, ctx.blends, ctx.enables, ctx.formats})

	p.bindings.reset()
	bind(7)
	require.Equal(t, [5]int{2, 3, 2, 2, 3}, [5]int{ctx.programs, ctx.buffers, ctx.blends, ctx.enables, ctx.formats})
}
