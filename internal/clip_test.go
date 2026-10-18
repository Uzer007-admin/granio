package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v2"
)

func TestClipStack_Intersect(t *testing.T) {
	p1 := granio.NewPos(5, 25)
	s1 := granio.NewSize(100, 100)
	c := &ClipStack{
		clips: []*ClipItem{
			{p1, s1},
		},
	}

	p2 := granio.NewPos(25, 0)
	s2 := granio.NewSize(50, 50)
	i := c.Push(p2, s2)

	assert.Equal(t, granio.NewPos(25, 25), i.pos)
	assert.Equal(t, granio.NewSize(50, 25), i.size)
	assert.Len(t, c.clips, 2)

	_ = c.Pop()
	p2 = granio.NewPos(50, 50)
	s2 = granio.NewSize(150, 50)
	i = c.Push(p2, s2)

	assert.Equal(t, granio.NewPos(50, 50), i.pos)
	assert.Equal(t, granio.NewSize(55, 50), i.size)
	assert.Len(t, c.clips, 2)
}

func TestClipStack_Pop(t *testing.T) {
	p := granio.NewPos(5, 5)
	s := granio.NewSize(100, 100)
	c := &ClipStack{
		clips: []*ClipItem{
			{p, s},
		},
	}

	i := c.Pop()
	assert.Equal(t, p, i.pos)
	assert.Equal(t, s, i.size)
	assert.Empty(t, c.clips)
}

func TestClipStack_Push(t *testing.T) {
	c := &ClipStack{}
	p := granio.NewPos(5, 5)
	s := granio.NewSize(100, 100)

	i := c.Push(p, s)
	assert.Equal(t, p, i.pos)
	assert.Equal(t, s, i.size)
	assert.Len(t, c.clips, 1)
}
