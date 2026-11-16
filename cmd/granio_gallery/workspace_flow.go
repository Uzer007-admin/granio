package main

import (
	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/container"
	"github.com/Uzer007-admin/granio/v3/widget"
)

// workspaceFlow lays out scroll content at the viewport width. At a fixed
// width only new items need measurement, even when the viewport gets shorter.
type workspaceFlow struct {
	widget.BaseWidget
	objects     []granio.CanvasObject
	inset, gap  float32
	width, end  float32
	bottomInset float32
	laidOut     int
}

func newWorkspaceFlow(inset, gap float32, objects ...granio.CanvasObject) *workspaceFlow {
	f := &workspaceFlow{inset: inset, gap: gap, objects: objects, width: -1}
	f.ExtendBaseWidget(f)
	return f
}

func (f *workspaceFlow) add(o granio.CanvasObject) {
	f.objects = append(f.objects, o)
	canvas.Refresh(f)
}

func (f *workspaceFlow) replace(objects []granio.CanvasObject) {
	f.objects = objects
	f.width = -1
	canvas.Refresh(f)
}

func (f *workspaceFlow) MinSize() granio.Size {
	return f.MinSizeForWidth(f.Size().Width)
}

func (f *workspaceFlow) MinSizeForWidth(width float32) granio.Size {
	if width <= 2*f.inset {
		return granio.Size{}
	}
	if f.width != width {
		f.width, f.laidOut, f.end = width, 0, 3
	}
	available := width - 2*f.inset
	for _, object := range f.objects[f.laidOut:] {
		height := workspaceMeasureHeight(object, available)
		workspacePlace(object, f.inset, f.end, available, height)
		f.end += height + f.gap
	}
	f.laidOut = len(f.objects)
	return granio.NewSize(0, f.end+5+f.bottomInset)
}

func (f *workspaceFlow) CreateRenderer() granio.WidgetRenderer { return workspaceFlowRenderer{f} }

type workspaceFlowRenderer struct{ flow *workspaceFlow }

func (r workspaceFlowRenderer) Layout(s granio.Size)           { r.flow.MinSizeForWidth(s.Width) }
func (r workspaceFlowRenderer) MinSize() granio.Size           { return r.flow.MinSize() }
func (r workspaceFlowRenderer) Objects() []granio.CanvasObject { return r.flow.objects }
func (r workspaceFlowRenderer) Destroy()                       {}
func (r workspaceFlowRenderer) Refresh() {
	r.flow.width = -1
	r.Layout(r.flow.Size())
	canvas.Refresh(r.flow)
}

func workspaceMeasureHeight(object granio.CanvasObject, width float32) float32 {
	switch o := object.(type) {
	case *widget.RichText:
		o.Resize(granio.NewSize(width, o.Size().Height))
	case *workspaceBlock:
		return o.measure(width)
	}
	return object.MinSize().Height
}

// workspaceBlock groups objects whose positions and height are both determined
// by the available width, such as the two cells of a document row.
type workspaceBlock struct {
	widget.BaseWidget
	objects       []granio.CanvasObject
	width, height float32
	arrange       func(float32) float32
}

func newWorkspaceBlock(arrange func(float32) float32, objects ...granio.CanvasObject) *workspaceBlock {
	b := &workspaceBlock{objects: objects, width: -1, arrange: arrange}
	b.ExtendBaseWidget(b)
	return b
}

func (b *workspaceBlock) measure(width float32) float32 {
	if b.width != width {
		b.width, b.height = width, b.arrange(width)
	}
	return b.height
}

func (b *workspaceBlock) CreateRenderer() granio.WidgetRenderer { return workspaceBlockRenderer{b} }

type workspaceBlockRenderer struct{ block *workspaceBlock }

func (r workspaceBlockRenderer) Layout(s granio.Size)           { r.block.measure(s.Width) }
func (r workspaceBlockRenderer) MinSize() granio.Size           { return granio.NewSize(0, r.block.height) }
func (r workspaceBlockRenderer) Objects() []granio.CanvasObject { return r.block.objects }
func (r workspaceBlockRenderer) Destroy()                       {}
func (r workspaceBlockRenderer) Refresh() {
	r.block.width = -1
	r.Layout(r.block.Size())
	canvas.Refresh(r.block)
}

func workspaceDocumentRow(label, value string) *workspaceBlock {
	left, right := workspaceParagraph(label), workspaceParagraph(value)
	leftClip, rightClip := container.NewClip(left), container.NewClip(right)
	rule := canvas.NewRectangle(workspaceBorder)
	return newWorkspaceBlock(func(width float32) float32 {
		const pad, gap = float32(4), float32(10)
		available := max(0, width-2*pad)
		// A narrow document has more useful line lengths when labels sit above
		// their values. At full width keep the label and value columns aligned.
		if width < 260 {
			leftHeight := workspaceMeasureHeight(left, available)
			rightHeight := workspaceMeasureHeight(right, available)
			height := pad + leftHeight + 2 + rightHeight + pad
			workspacePlace(leftClip, pad, pad, available, leftHeight)
			workspacePlace(rightClip, pad, pad+leftHeight+2, available, rightHeight)
			workspacePlace(rule, 0, height-1, width, 1)
			return height
		}
		leftWidth := (available - gap) * 0.42
		rightWidth := available - leftWidth - gap
		leftHeight := workspaceMeasureHeight(left, leftWidth)
		rightHeight := workspaceMeasureHeight(right, rightWidth)
		height := 2*pad + max(leftHeight, rightHeight)
		workspacePlace(leftClip, pad, pad, leftWidth, leftHeight)
		workspacePlace(rightClip, pad+leftWidth+gap, pad, rightWidth, rightHeight)
		workspacePlace(rule, 0, height-1, width, 1)
		return height
	}, leftClip, rightClip, rule)
}

func workspaceParagraph(value string) *widget.RichText {
	text := widget.NewRichText(&widget.TextSegment{Text: value, Style: widget.RichTextStyleParagraph})
	text.Wrapping = granio.TextWrapWord
	return text
}
