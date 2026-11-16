package software

import (
	"container/list"
	"image"
	"image/color"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/internal/painter"
)

const (
	textCacheEntries = 256
	textCacheBytes   = 4 << 20
)

type textBitmapKey struct {
	text          string
	font          *painter.FontCacheItem
	style         granio.TextStyle
	color         color.RGBA64
	width, height int
	scale, size   float32
}

type textBitmap struct {
	key    textBitmapKey
	pixels *image.RGBA
	bytes  int
}

// textBitmapCache is local to the window's painter. Reusing rasterized labels
// also avoids keeping references to widgets that have left the render tree.
type textBitmapCache struct {
	order list.List
	items map[textBitmapKey]*list.Element
	bytes int
}

func (c *textBitmapCache) get(key textBitmapKey) *image.RGBA {
	if hit := c.items[key]; hit != nil {
		c.order.MoveToFront(hit)
		return hit.Value.(textBitmap).pixels
	}
	return nil
}

func (c *textBitmapCache) add(key textBitmapKey, pixels *image.RGBA) {
	bytes := len(pixels.Pix) + len(key.text)
	if bytes > textCacheBytes {
		return
	}
	if c.items == nil {
		c.items = make(map[textBitmapKey]*list.Element)
	}
	for len(c.items) >= textCacheEntries || c.bytes+bytes > textCacheBytes {
		tail := c.order.Back()
		old := c.order.Remove(tail).(textBitmap)
		delete(c.items, old.key)
		c.bytes -= old.bytes
	}
	c.items[key] = c.order.PushFront(textBitmap{key: key, pixels: pixels, bytes: bytes})
	c.bytes += bytes
}
