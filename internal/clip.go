package internal

import (
	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/internal/driver"
)

// ClipStack keeps track of the areas that should be clipped when drawing a canvas.
// If no clips are present then adding one will be added as-is.
// Subsequent items pushed will be completely within the previous clip.
type ClipStack struct {
	clips []*ClipItem
}

// Pop removes the current top clip and returns it.
func (c *ClipStack) Pop() *ClipItem {
	if len(c.clips) == 0 {
		return nil
	}

	top := len(c.clips) - 1
	ret := c.clips[top]
	c.clips[top] = nil // release memory reference
	c.clips = c.clips[:top]
	return ret
}

// Length  returns the number of items in this clip stack. 0 means no clip.
func (c *ClipStack) Length() int {
	return len(c.clips)
}

// Push a new clip onto this stack at position and size specified.
// The returned clip item is the result of calculating the intersection of the requested clip and its parent.
func (c *ClipStack) Push(p granio.Position, s granio.Size) *ClipItem {
	outer := c.Top()
	inner := outer.Intersect(p, s)

	c.clips = append(c.clips, inner)
	return inner
}

// Top returns the current clip item - it will always be within the bounds of any parent clips.
func (c *ClipStack) Top() *ClipItem {
	if len(c.clips) == 0 {
		return nil
	}

	return c.clips[len(c.clips)-1]
}

// ClipItem represents a single clip in a clip stack, denoted by a size and position.
type ClipItem struct {
	pos  granio.Position
	size granio.Size
}

// Rect returns the position and size parameters of the clip.
func (i *ClipItem) Rect() (granio.Position, granio.Size) {
	return i.pos, i.size
}

// Intersect returns a new clip item that is the intersection of the requested parameters and this clip.
func (i *ClipItem) Intersect(p granio.Position, s granio.Size) *ClipItem {
	if i != nil {
		p, s = driver.IntersectClip(i.pos, i.size, p, s)
	}
	return &ClipItem{p, s}
}
