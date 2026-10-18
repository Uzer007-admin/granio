package driver_test

import (
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/internal/driver"
	internal_widget "github.com/Uzer007-admin/granio/v2/internal/widget"
	_ "github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/widget"

	"github.com/stretchr/testify/assert"
)

func TestAbsolutePositionForObject(t *testing.T) {
	t1r1c1 := widget.NewLabel("row 1 col 1")
	t1r1c2 := widget.NewLabel("row 1 col 2")
	t1r2c1 := widget.NewLabel("row 2 col 1")
	t1r2c2 := widget.NewLabel("row 2 col 2")
	t1r2c2.Hide()
	t1r1 := container.NewWithoutLayout(t1r1c1, t1r1c2)
	t1r2 := container.NewWithoutLayout(t1r2c1, t1r2c2)
	tree1 := container.NewWithoutLayout(t1r1, t1r2)

	t1r1c1.Move(granio.NewPos(111, 111))
	t1r1c2.Move(granio.NewPos(112, 112))
	t1r2c1.Move(granio.NewPos(121, 121))
	t1r2c2.Move(granio.NewPos(122, 122))
	t1r1.Move(granio.NewPos(11, 11))
	t1r2.Move(granio.NewPos(12, 12))
	tree1.Move(granio.NewPos(1, 1))

	t2r1c1 := widget.NewLabel("row 1 col 1")
	t2r1c2 := widget.NewLabel("row 1 col 2")
	t2r2c1 := widget.NewLabel("row 2 col 1")
	t2r2c2 := widget.NewLabel("row 2 col 2")
	t2r1 := container.NewWithoutLayout(t2r1c1, t2r1c2)
	t2r2 := container.NewWithoutLayout(t2r2c1, t2r2c2)
	tree2 := container.NewWithoutLayout(t2r1, t2r2)

	t2r1c1.Move(granio.NewPos(211, 211))
	t2r1c2.Move(granio.NewPos(212, 212))
	t2r2c1.Move(granio.NewPos(221, 221))
	t2r2c2.Move(granio.NewPos(222, 222))
	t2r1.Move(granio.NewPos(21, 21))
	t2r2.Move(granio.NewPos(22, 22))
	tree2.Move(granio.NewPos(2, 2))

	t3r1 := widget.NewLabel("row 1")
	t3r2 := widget.NewLabel("row 2")
	tree3 := container.NewWithoutLayout(t3r1, t3r2)

	t3r1.Move(granio.NewPos(31, 31))
	t3r2.Move(granio.NewPos(32, 32))
	tree3.Move(granio.NewPos(3, 3))

	trees := []granio.CanvasObject{tree1, tree2, tree3}

	outsideTrees := widget.NewLabel("outside trees")
	outsideTrees.Move(granio.NewPos(10, 10))

	tests := map[string]struct {
		object granio.CanvasObject
		want   granio.Position
	}{
		"tree 1: a cell": {
			object: t1r1c2,
			want:   granio.NewPos(124, 124), // 1 (root) + 11 (row 1) + 112 (cell 2)
		},
		"tree 1: a row": {
			object: t1r2,
			want:   granio.NewPos(13, 13), // 1 (root) + 12 (row 2)
		},
		"tree 1: root": {
			object: tree1,
			want:   granio.NewPos(1, 1),
		},
		"tree 1: a hidden element": {
			object: t1r2c2,
			want:   granio.NewPos(0, 0),
		},

		"tree 2: a row": {
			object: t2r2,
			want:   granio.NewPos(24, 24), // 2 (root) + 22 (row 2)
		},
		"tree 2: root": {
			object: tree2,
			want:   granio.NewPos(2, 2),
		},

		"tree 3: a row": {
			object: t3r2,
			want:   granio.NewPos(35, 35), // 3 (root) + 32 (row 2)
		},
		"tree 3: root": {
			object: tree3,
			want:   granio.NewPos(3, 3),
		},

		"an object not inside any tree": {
			object: outsideTrees,
			want:   granio.NewPos(0, 0),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, driver.AbsolutePositionForObject(tt.object, trees))
		})
	}
}

func TestFindObjectAtPositionMatching(t *testing.T) {
	col1cell1 := &objectTree{
		pos:  granio.NewPos(10, 10),
		size: granio.NewSize(15, 15),
	}
	col1cell2 := &objectTree{
		pos:  granio.NewPos(10, 35),
		size: granio.NewSize(15, 15),
	}
	col1cell3 := &objectTree{
		pos:  granio.NewPos(10, 60),
		size: granio.NewSize(15, 15),
	}
	col1 := &objectTree{
		children: []granio.CanvasObject{col1cell1, col1cell2, col1cell3},
		pos:      granio.NewPos(10, 10),
		size:     granio.NewSize(35, 80),
	}
	col2cell1 := &objectTree{
		pos:  granio.NewPos(10, 10),
		size: granio.NewSize(15, 15),
	}
	col2cell2 := &objectTree{
		pos:  granio.NewPos(10, 35),
		size: granio.NewSize(15, 15),
	}
	col2cell3 := &objectTree{
		pos:  granio.NewPos(10, 60),
		size: granio.NewSize(15, 15),
	}
	col2 := &objectTree{
		children: []granio.CanvasObject{col2cell1, col2cell2, col2cell3},
		pos:      granio.NewPos(55, 10),
		size:     granio.NewSize(35, 80),
	}
	colTree := &objectTree{
		children: []granio.CanvasObject{col1, col2},
		pos:      granio.NewPos(10, 10),
		size:     granio.NewSize(100, 100),
	}
	row1cell1 := &objectTree{
		pos:  granio.NewPos(10, 10),
		size: granio.NewSize(15, 15),
	}
	row1cell2 := &objectTree{
		pos:  granio.NewPos(35, 10),
		size: granio.NewSize(15, 15),
	}
	row1cell3 := &objectTree{
		pos:  granio.NewPos(60, 10),
		size: granio.NewSize(15, 15),
	}
	row1 := &objectTree{
		children: []granio.CanvasObject{row1cell1, row1cell2, row1cell3},
		pos:      granio.NewPos(10, 10),
		size:     granio.NewSize(80, 35),
	}
	row2cell1 := &objectTree{
		pos:  granio.NewPos(10, 10),
		size: granio.NewSize(15, 15),
	}
	row2cell2 := &objectTree{
		pos:  granio.NewPos(35, 10),
		size: granio.NewSize(15, 15),
	}
	row2cell3 := &objectTree{
		pos:  granio.NewPos(60, 10),
		size: granio.NewSize(15, 15),
	}
	row2 := &objectTree{
		children: []granio.CanvasObject{row2cell1, row2cell2, row2cell3},
		pos:      granio.NewPos(10, 55),
		size:     granio.NewSize(80, 35),
	}
	rowTree := &objectTree{
		children: []granio.CanvasObject{row1, row2},
		pos:      granio.NewPos(10, 10),
		size:     granio.NewSize(100, 100),
	}
	tree1 := &objectTree{
		pos:  granio.NewPos(100, 100),
		size: granio.NewSize(5, 5),
	}
	tree2 := &objectTree{
		pos:  granio.NewPos(0, 0),
		size: granio.NewSize(5, 5),
	}
	tree3 := &objectTree{
		pos:  granio.NewPos(50, 50),
		size: granio.NewSize(5, 5),
	}
	for name, tt := range map[string]struct {
		matcher    func(object granio.CanvasObject) bool
		overlay    granio.CanvasObject
		pos        granio.Position
		roots      []granio.CanvasObject
		wantObject granio.CanvasObject
		wantPos    granio.Position
		wantLayer  int
	}{
		"match in overlay and roots": {
			matcher:    func(o granio.CanvasObject) bool { return o.Size().Width == 15 },
			overlay:    colTree,
			pos:        granio.NewPos(35, 60),
			roots:      []granio.CanvasObject{rowTree},
			wantObject: col1cell2,
			wantPos:    granio.NewPos(5, 5),
			wantLayer:  0,
		},
		"match in root but overlay without match present": {
			matcher:    func(o granio.CanvasObject) bool { return o.Size().Width == 15 },
			overlay:    tree1,
			pos:        granio.NewPos(35, 60),
			roots:      []granio.CanvasObject{colTree, rowTree},
			wantObject: nil,
			wantPos:    granio.Position{},
			wantLayer:  0,
		},
		"match in multiple roots without overlay": {
			matcher:    func(o granio.CanvasObject) bool { return o.Size().Width == 15 },
			overlay:    nil,
			pos:        granio.NewPos(83, 83),
			roots:      []granio.CanvasObject{tree1, rowTree, tree2, colTree},
			wantObject: row2cell3,
			wantPos:    granio.NewPos(3, 8),
			wantLayer:  2,
		},
		"no match in roots without overlay": {
			matcher:    func(granio.CanvasObject) bool { return true },
			overlay:    nil,
			pos:        granio.NewPos(66, 66),
			roots:      []granio.CanvasObject{tree1, tree2, tree3},
			wantObject: nil,
			wantPos:    granio.Position{},
			wantLayer:  3,
		},
		"no overlay and no roots": {
			matcher:    func(granio.CanvasObject) bool { return true },
			overlay:    nil,
			pos:        granio.NewPos(66, 66),
			roots:      nil,
			wantObject: nil,
			wantPos:    granio.Position{},
			wantLayer:  0,
		},
	} {
		t.Run(name, func(t *testing.T) {
			o, p, l := driver.FindObjectAtPositionMatching(tt.pos, tt.matcher, tt.overlay, tt.roots...)
			assert.Equal(t, tt.wantObject, o, "found object")
			assert.Equal(t, tt.wantPos, p, "position of found object")
			assert.Equal(t, tt.wantLayer, l, "layer of found object (0 - overlay, 1, 2, 3… - roots")
		})
	}
}

func TestReverseWalkVisibleObjectTree(t *testing.T) {
	child1 := canvas.NewRectangle(color.White)
	child1.SetMinSize(granio.NewSize(100, 100))
	child2 := canvas.NewRectangle(color.Black)
	child2.Hide()
	child3 := canvas.NewRectangle(color.White)
	base := container.NewHBox(child1, child2, child3)

	var walked []granio.CanvasObject
	driver.ReverseWalkVisibleObjectTree(
		base,
		func(object granio.CanvasObject, _ granio.Position, _ granio.Position, _ granio.Size) bool {
			walked = append(walked, object)
			return false
		},
		nil,
	)

	assert.Equal(t, []granio.CanvasObject{base, child3, child1}, walked)
}

func TestReverseWalkVisibleObjectTree_Clip(t *testing.T) {
	rect := canvas.NewRectangle(color.White)
	rect.SetMinSize(granio.NewSize(100, 100))
	child := canvas.NewRectangle(color.Black)
	base := container.NewGridWithColumns(
		1,
		rect,
		internal_widget.NewScroll(child),
		container.NewGridWithColumns(
			2,
			canvas.NewCircle(color.White),
			canvas.NewCircle(color.White),
			canvas.NewCircle(color.White),
			&scrollable{},
		),
	)

	// We are not in a window. Resize the container to a default size.
	base.Resize(base.MinSize())

	var scClipPos, scrollableClipPos granio.Position
	var scClipSize, scrollableClipSize granio.Size

	driver.ReverseWalkVisibleObjectTree(base, func(object granio.CanvasObject, _ granio.Position, clippingPos granio.Position, clippingSize granio.Size) bool {
		if _, ok := object.(*internal_widget.Scroll); ok {
			scClipPos = clippingPos
			scClipSize = clippingSize
		} else if _, ok = object.(granio.Scrollable); ok {
			scrollableClipPos = clippingPos
			scrollableClipSize = clippingSize
		}
		return false
	}, nil)

	// layout:
	// +-------------------------------+
	// | 0,0: rect 100x100             |
	// +-------------------------------+
	// |            padding            |
	// +-------------------------------+
	// | 0,104: scroller 100x100       |
	// +-------------------------------+
	// |            padding            |
	// +--------------+-+--------------+
	// | circle 48x48 |p| circle 48x48 |
	// +--------------+-+--------------+
	// |            padding            |
	// +--------------+-+--------------+
	// | circle 48x48 |p| scrollable   |
	// +--------------+-+--------------+
	assert.Equal(t, granio.NewPos(0, 104), scClipPos)
	assert.Equal(t, granio.NewSize(100, 100), scClipSize)
	assert.Equal(t, granio.NewPos(52, 260), scrollableClipPos)
	assert.Equal(t, granio.NewSize(48, 48), scrollableClipSize)
}

func TestWalkVisibleObjectTree(t *testing.T) {
	child1 := canvas.NewRectangle(color.White)
	child1.SetMinSize(granio.NewSize(100, 100))
	child2 := canvas.NewRectangle(color.Black)
	child2.Hide()
	child3 := canvas.NewRectangle(color.White)
	base := container.NewHBox(child1, child2, child3)

	var walked []granio.CanvasObject
	driver.WalkVisibleObjectTree(base, func(object granio.CanvasObject, _ granio.Position, _ granio.Position, _ granio.Size) bool {
		walked = append(walked, object)
		return false
	}, nil)

	assert.Equal(t, []granio.CanvasObject{base, child1, child3}, walked)
}

func TestWalkVisibleObjectTree_Clip(t *testing.T) {
	rect := canvas.NewRectangle(color.White)
	rect.SetMinSize(granio.NewSize(100, 100))
	child := canvas.NewRectangle(color.Black)
	base := container.NewGridWithColumns(
		1,
		rect,
		internal_widget.NewScroll(child),
		container.NewGridWithColumns(
			2,
			canvas.NewCircle(color.White),
			canvas.NewCircle(color.White),
			canvas.NewCircle(color.White),
			&scrollable{},
		),
	)

	// We are not in a window. Resize the container to a default size.
	base.Resize(base.MinSize())

	var scClipPos, scrollableClipPos granio.Position
	var scClipSize, scrollableClipSize granio.Size

	driver.WalkVisibleObjectTree(base, func(object granio.CanvasObject, _ granio.Position, clippingPos granio.Position, clippingSize granio.Size) bool {
		if _, ok := object.(*internal_widget.Scroll); ok {
			scClipPos = clippingPos
			scClipSize = clippingSize
		} else if _, ok = object.(granio.Scrollable); ok {
			scrollableClipPos = clippingPos
			scrollableClipSize = clippingSize
		}
		return false
	}, nil)

	// layout:
	// +-------------------------------+
	// | 0,0: rect 100x100             |
	// +-------------------------------+
	// |            padding            |
	// +-------------------------------+
	// | 0,104: scroller 100x100       |
	// +-------------------------------+
	// |            padding            |
	// +--------------+-+--------------+
	// | circle 48x48 |p| circle 48x48 |
	// +--------------+-+--------------+
	// |            padding            |
	// +--------------+-+--------------+
	// | circle 48x48 |p| scrollable   |
	// +--------------+-+--------------+
	assert.Equal(t, granio.NewPos(0, 104), scClipPos)
	assert.Equal(t, granio.NewSize(100, 100), scClipSize)
	assert.Equal(t, granio.NewPos(52, 260), scrollableClipPos)
	assert.Equal(t, granio.NewSize(48, 48), scrollableClipSize)
}

func TestWalkWholeObjectTree(t *testing.T) {
	child1 := canvas.NewRectangle(color.White)
	child1.SetMinSize(granio.NewSize(100, 100))
	child2 := canvas.NewRectangle(color.Black)
	child2.Hide()
	child3 := canvas.NewRectangle(color.White)
	base := container.NewHBox(child1, child2, child3)

	var walked []granio.CanvasObject
	driver.WalkCompleteObjectTree(base, func(object granio.CanvasObject, _ granio.Position, _ granio.Position, _ granio.Size) bool {
		walked = append(walked, object)
		return false
	}, nil)

	assert.Equal(t, []granio.CanvasObject{base, child1, child2, child3}, walked)
}

var _ granio.Widget = (*objectTree)(nil)

type objectTree struct {
	children []granio.CanvasObject
	hidden   bool
	pos      granio.Position
	size     granio.Size
}

func (o *objectTree) Size() granio.Size {
	return o.size
}

func (o *objectTree) Resize(size granio.Size) {
	o.size = size
}

func (o *objectTree) Position() granio.Position {
	return o.pos
}

func (o *objectTree) Move(position granio.Position) {
	o.pos = position
}

func (o *objectTree) MinSize() granio.Size {
	return o.size
}

func (o objectTree) Visible() bool {
	return !o.hidden
}

func (o *objectTree) Show() {
	o.hidden = false
}

func (o *objectTree) Hide() {
	o.hidden = true
}

func (*objectTree) Refresh() {
}

func (o *objectTree) CreateRenderer() granio.WidgetRenderer {
	r := &objectTreeRenderer{}
	r.SetObjects(o.children)
	return r
}

var _ granio.WidgetRenderer = (*objectTreeRenderer)(nil)

type objectTreeRenderer struct {
	internal_widget.BaseRenderer
}

func (objectTreeRenderer) Layout(_ granio.Size) {
}

func (objectTreeRenderer) MinSize() granio.Size {
	return granio.NewSize(0, 0)
}

func (objectTreeRenderer) Refresh() {
}

type scrollable struct {
	pos  granio.Position
	size granio.Size
}

var (
	_ granio.CanvasObject = (*scrollable)(nil)
	_ granio.Scrollable   = (*scrollable)(nil)
)

func (*scrollable) Hide() {
	panic("implement me")
}

func (*scrollable) MinSize() granio.Size {
	return granio.NewSize(1, 1)
}

func (s *scrollable) Move(position granio.Position) {
	s.pos = position
}

func (s *scrollable) Position() granio.Position {
	return s.pos
}

func (*scrollable) Refresh() {
	panic("implement me")
}

func (s *scrollable) Resize(size granio.Size) {
	s.size = size
}

func (*scrollable) Scrolled(_ *granio.ScrollEvent) {
	panic("implement me")
}

func (*scrollable) Show() {
	panic("implement me")
}

func (s *scrollable) Size() granio.Size {
	return s.size
}

func (*scrollable) Visible() bool {
	return true
}
