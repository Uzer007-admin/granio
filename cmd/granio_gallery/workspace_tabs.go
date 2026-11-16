package main

import (
	"math"
	"slices"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/container"
	"github.com/Uzer007-admin/granio/v3/driver/desktop"
	"github.com/Uzer007-admin/granio/v3/widget"
)

var workspaceDocNames = []string{"QA.md", "ОТЧЁТ.md", "README.md"}

type workspaceDocTab struct {
	widget.Button
	view  *workspaceView
	id    int
	frame *container.Clip
}

func (t *workspaceDocTab) Dragged(e *granio.DragEvent) { t.view.dragDocument(t.id, e.Dragged.DX) }
func (t *workspaceDocTab) DragEnd()                    { t.view.finishDocumentDrag() }
func (t *workspaceDocTab) Cursor() desktop.Cursor      { return desktop.DefaultCursor }

func (v *workspaceView) documentTabs() *granio.Container {
	v.docTabs = make([]*workspaceDocTab, len(workspaceDocNames))
	items := make([]granio.CanvasObject, len(v.docTabs))
	for i, name := range workspaceDocNames {
		tab := &workspaceDocTab{view: v, id: i, Button: widget.Button{
			Style: widget.ButtonStyleToolbar, Alignment: widget.ButtonAlignLeading,
			AccessibleName: name, Icon: workspaceIcon("document", workspaceSubtle),
			OnTapped: func() { v.selectDocument(i) },
		}}
		tab.ExtendBaseWidget(tab)
		label := widget.NewLabel(name)
		label.Truncation = granio.TextTruncateEllipsis
		label.Importance = widget.LowImportance
		tab.frame = container.NewClip(workspaceArrange(granio.Size{}, func(s granio.Size) {
			workspacePlace(tab, 0, 0, s.Width, s.Height)
			workspacePlace(label, 20, 0, s.Width-22, s.Height)
			workspaceShow(label, s.Width >= 50)
		}, tab, label))
		v.docTabs[i], items[i] = tab, tab.frame
	}
	v.tabHighlight = workspaceSurface(workspaceColor(0xc3d8ee), 1)
	v.tabStrip = workspaceArrange(granio.Size{}, func(s granio.Size) {
		width := v.documentTabWidth(s.Width)
		for slot, id := range v.docOrder {
			x := float32(slot) * width
			if id == v.draggedTab {
				x = min(max(0, v.tabDragX), width*float32(len(v.docOrder)-1))
			} else if id == v.slidingTab {
				x += v.tabSlide
			}
			workspacePlace(v.docTabs[id].frame, x, 0, width-3, s.Height)
		}
		x := v.tabHighlightAt * width
		if v.draggedTab == v.doc || v.slidingTab == v.doc {
			x = v.docTabs[v.doc].frame.Position().X
		}
		workspacePlace(v.tabHighlight, x+6, s.Height-3, width-15, 2)
	}, items...)
	v.tabStrip.Objects = append(v.tabStrip.Objects, v.tabHighlight)
	return v.tabStrip
}

func (v *workspaceView) moveTabHighlight(index int) {
	if previous := v.highlightMotion; previous != nil {
		previous.Stop()
		v.highlightMotion = nil
	}
	target := float32(slices.Index(v.docOrder, index))
	from := v.tabHighlightAt
	if v.draggedTab == index || from == target || !v.documentPane.PanelOpen() {
		v.tabHighlightAt = target
		v.layoutDocumentTabs()
		return
	}
	a := workspaceAnimation(v.tabStrip, granio.MotionNavigation, func(p float32) {
		v.tabHighlightAt = from + (target-from)*p
		v.layoutDocumentTabs()
	})
	a.OnFinished = func() {
		if v.highlightMotion == a {
			v.highlightMotion = nil
		}
	}
	v.highlightMotion = a
	a.Start()
}

func (v *workspaceView) documentTabWidth(available float32) float32 {
	return min(105, available/float32(len(v.docTabs)))
}

func (v *workspaceView) layoutDocumentTabs() {
	v.tabStrip.Layout.Layout(v.tabStrip.Objects, v.tabStrip.Size())
}

func (v *workspaceView) moveDocumentTab(id, to int) {
	from := slices.Index(v.docOrder, id)
	if to < 0 || to >= len(v.docOrder) || from == to {
		return
	}
	if from < to {
		copy(v.docOrder[from:to], v.docOrder[from+1:to+1])
	} else {
		copy(v.docOrder[to+1:from+1], v.docOrder[to:from])
	}
	v.docOrder[to] = id
}

func (v *workspaceView) dragDocument(id int, delta float32) {
	width := v.documentTabWidth(v.tabStrip.Size().Width)
	if width <= 0 {
		return
	}
	if v.draggedTab != id {
		if v.tabAnimation != nil {
			v.tabAnimation.Stop()
			v.tabAnimation = nil
		}
		v.draggedTab, v.slidingTab = id, -1
		v.tabDragX = float32(slices.Index(v.docOrder, id)) * width
		// Place the tab above its neighbours while it follows the pointer.
		frames := v.tabStrip.Objects[:len(v.docTabs)]
		at := slices.Index(frames, granio.CanvasObject(v.docTabs[id].frame))
		copy(frames[at:], frames[at+1:])
		frames[len(frames)-1] = v.docTabs[id].frame
		v.selectDocument(id)
	}
	v.tabDragX = min(max(0, v.tabDragX+delta), width*float32(len(v.docOrder)-1))
	v.moveDocumentTab(id, int(math.Round(float64(v.tabDragX/width))))
	v.tabHighlightAt = float32(slices.Index(v.docOrder, id))
	v.layoutDocumentTabs()
}

func (v *workspaceView) finishDocumentDrag() {
	if v.draggedTab < 0 {
		return
	}
	id := v.draggedTab
	v.draggedTab = -1
	target := float32(slices.Index(v.docOrder, id)) * v.documentTabWidth(v.tabStrip.Size().Width)
	offset := v.tabDragX - target
	if offset == 0 {
		v.layoutDocumentTabs()
		return
	}
	v.slidingTab, v.tabSlide = id, offset
	a := workspaceAnimation(v.tabStrip, granio.MotionNavigation, func(p float32) {
		v.tabSlide = offset * (1 - p)
		v.layoutDocumentTabs()
	})
	a.OnFinished = func() {
		v.tabAnimation, v.slidingTab = nil, -1
	}
	v.tabAnimation = a
	a.Start()
}
