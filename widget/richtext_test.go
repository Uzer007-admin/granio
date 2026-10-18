package widget

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"
)

func richTextRenderTexts(rich granio.Widget) []*canvas.Text {
	renderer := cache.Renderer(rich).(*textRenderer)
	texts := make([]*canvas.Text, len(renderer.Objects()))
	for i, obj := range renderer.Objects() {
		texts[i] = obj.(*canvas.Text)
	}
	return texts
}

// richTextRowTexts returns the text shown on each row, including any ellipsis.
func richTextRowTexts(rich *RichText) []string {
	rows := make([]string, len(rich.rowBounds))
	for i := range rich.rowBounds {
		bound := &rich.rowBounds[i]
		for j := range bound.segments {
			rows[i] += string(rowSegmentRunes(bound, j))
		}
		if bound.ellipsis {
			rows[i] += ellipsisChar
		}
	}
	return rows
}

func trailingBoldErrorSegment() *TextSegment {
	return &TextSegment{Style: RichTextStyle{
		Alignment: granio.TextAlignTrailing,
		ColorName: theme.ColorNameError,
		TextStyle: granio.TextStyle{Bold: true},
	}}
}

func TestRichText_Hyperlink_Endline(t *testing.T) {
	u, _ := url.Parse("https://github.com/fyne-io/fyne")
	r := NewRichText(
		&TextSegment{Text: "Text", Style: RichTextStyleInline},
		&HyperlinkSegment{Text: "Link", URL: u},
	)
	r.Resize(r.MinSize())
	view := cache.Renderer(r)

	assert.Len(t, view.Objects(), 2)
	assert.Equal(t, view.Objects()[0].Position().Y, view.Objects()[1].Position().Y)   // same baseline
	assert.Greater(t, view.Objects()[1].Position().X, view.Objects()[0].Position().X) // to the right
}

func TestText_Alignment(t *testing.T) {
	seg := trailingBoldErrorSegment()
	seg.Text = "Test"
	text := NewRichText(seg)
	assert.Equal(t, granio.TextAlignTrailing, test.TempWidgetRenderer(t, text).Objects()[0].(*canvas.Text).Alignment)
}

func TestText_AlignmentInline(t *testing.T) {
	seg1 := &TextSegment{
		Style: RichTextStyle{
			Alignment: granio.TextAlignTrailing,
		},
		Text: "Thing 1",
	}
	seg2 := &TextSegment{
		Style: RichTextStyle{
			Alignment: granio.TextAlignTrailing,
			Inline:    true,
		},
		Text: "Thing 2",
	}
	seg3 := &TextSegment{
		Style: RichTextStyle{
			Alignment: granio.TextAlignTrailing,
			Inline:    true,
		},
		Text: "Thing 3",
	}
	text := NewRichText(seg1, seg2, seg3)
	text.Resize(granio.NewSize(150, 50))

	r := test.TempWidgetRenderer(t, text)
	obj1 := r.Objects()[0].(*canvas.Text)
	right1 := obj1.Position().X + obj1.Size().Width
	assert.Equal(t, granio.TextAlignTrailing, obj1.Alignment)
	obj3 := r.Objects()[2].(*canvas.Text)
	right3 := obj3.Position().X + obj3.Size().Width
	assert.Equal(t, granio.TextAlignTrailing, obj3.Alignment)

	assert.Equal(t, right1, right3)
}

func TestText_AlignmentInlineCenter(t *testing.T) {
	seg1 := &TextSegment{
		Style: RichTextStyle{
			Alignment: granio.TextAlignCenter,
		},
		Text: "MMM",
	}
	seg2 := &TextSegment{
		Style: RichTextStyle{
			Alignment: granio.TextAlignCenter,
			Inline:    true,
		},
		Text: "MMM",
	}
	seg3 := &TextSegment{
		Style: RichTextStyle{
			Alignment: granio.TextAlignCenter,
			Inline:    true,
		},
		Text: "MMM",
	}
	seg4 := &TextSegment{
		Style: RichTextStyle{
			Alignment: granio.TextAlignCenter,
			Inline:    true,
		},
		Text: "MMM",
	}
	text := NewRichText(seg1, seg2, seg3, seg4)
	text.Resize(granio.NewSize(150, 50))

	r := test.TempWidgetRenderer(t, text)
	obj1 := r.Objects()[0].(*canvas.Text)
	left1 := obj1.Position().X
	assert.Equal(t, granio.TextAlignCenter, obj1.Alignment)
	obj3 := r.Objects()[2].(*canvas.Text)
	left3 := obj3.Position().X
	assert.Equal(t, granio.TextAlignCenter, obj3.Alignment)

	assert.Equal(t, left1, left3)
}

func TestText_Row(t *testing.T) {
	text := NewRichTextWithText("")
	text.Segments[0].(*TextSegment).Text = "test"
	text.Refresh()

	assert.Nil(t, text.row(-1))
	assert.Nil(t, text.row(1))

	assert.Equal(t, []rune("test"), text.row(0))
}

func TestText_Rows(t *testing.T) {
	text := NewRichTextWithText("test")
	assert.Equal(t, 1, text.rows())
	textSeg := text.Segments[0].(*TextSegment)

	textSeg.Text = "test\ntest"
	text.Refresh()
	assert.Equal(t, 2, text.rows())

	textSeg.Text = "test\ntest\ntest"
	text.Refresh()
	assert.Equal(t, 3, text.rows())

	textSeg.Text = "test\ntest\ntest\n"
	text.Refresh()
	assert.Equal(t, 4, text.rows())

	textSeg.Text = "\n"
	text.Refresh()
	assert.Equal(t, 2, text.rows())
}

func TestText_RowLength(t *testing.T) {
	text := NewRichTextWithText("test")

	rl := text.rowLength(0)
	assert.Equal(t, 4, rl)
	textSeg := text.Segments[0].(*TextSegment)

	textSeg.Text = "test\ntèsts"
	text.Refresh()
	rl = text.rowLength(0)
	assert.Equal(t, 4, rl)

	rl = text.rowLength(1)
	assert.Equal(t, 5, rl)

	textSeg.Text = ""
	text.Refresh()
	rl = text.rowLength(0)
	assert.Equal(t, 0, rl)

	textSeg.Text = "\nhello"
	text.Refresh()
	rl = text.rowLength(0)
	assert.Equal(t, 0, rl)

	rl = text.rowLength(1)
	assert.Equal(t, 5, rl)
}

func TestText_Scroll(t *testing.T) {
	text1 := NewRichTextWithText("test1\ntest2")
	text2 := NewRichTextWithText("test1\ntest2")
	text2.Scroll = widget.ScrollBoth

	assert.Less(t, text2.MinSize().Width, text1.MinSize().Width)
	assert.Less(t, text2.MinSize().Height, text1.MinSize().Height)

	text3 := NewRichTextWithText("test1\ntest2")
	text3.Scroll = widget.ScrollVerticalOnly
	assert.Equal(t, text3.MinSize().Width, text1.MinSize().Width)
	assert.Less(t, text3.MinSize().Height, text1.MinSize().Height)

	text4 := NewRichTextWithText("test1\ntest2")
	text4.Scroll = widget.ScrollVerticalOnly
	text4.Wrapping = granio.TextWrapBreak

	text3.Resize(granio.NewSize(32, 32))
	text4.Resize(granio.NewSize(32, 32))
	assert.Less(t, text4.MinSize().Width, text3.MinSize().Width)
	assert.Equal(t, text4.MinSize().Height, text3.MinSize().Height)

	content3 := test.TempWidgetRenderer(t, text3).Objects()[0].(*widget.Scroll).Content
	content4 := test.TempWidgetRenderer(t, text4).Objects()[0].(*widget.Scroll).Content
	assert.Less(t, content4.MinSize().Width, content3.MinSize().Width)
	assert.Greater(t, content4.MinSize().Height, content3.MinSize().Height)
}

func TestText_InsertAt(t *testing.T) {
	type fields struct {
		buffer string
	}
	type args struct {
		pos   int
		runes string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantBuffer string
	}{
		{
			name:   "case_1",
			fields: fields{buffer: "A\n1"},
			args: args{
				pos:   0,
				runes: "\n",
			},
			wantBuffer: "\nA\n1",
		},
		{
			name:   "case_2",
			fields: fields{buffer: "hello\nèé+^#"},
			args: args{
				pos:   5,
				runes: "\naddme",
			},
			wantBuffer: "hello\naddme\nèé+^#",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := NewRichTextWithText(tt.fields.buffer)
			text.insertAt(tt.args.pos, []rune(tt.args.runes))
			assert.Equal(t, tt.wantBuffer, text.String())
		})
	}
}

func TestText_Insert(t *testing.T) {
	text := NewRichTextWithText("")
	text.insertAt(0, []rune("a"))
	assert.Equal(t, "a", text.String())
	text.insertAt(1, []rune("\n"))
	assert.Equal(t, "a\n", text.String())
	text.insertAt(2, []rune("b"))
	assert.Equal(t, "a\nb", text.String())
}

func TestText_DeleteFromTo(t *testing.T) {
	type fields struct {
		buffer string
	}
	type args struct {
		lowBound  int
		highBound int
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       string
		wantBuffer string
	}{
		{
			name:   "case_1",
			fields: fields{buffer: "A\n1"},
			args: args{
				lowBound:  0,
				highBound: 1,
			},
			want:       "A",
			wantBuffer: "\n1",
		},
		{
			name:   "case_2",
			fields: fields{buffer: "A\n1"},
			args: args{
				lowBound:  1,
				highBound: 2,
			},
			want:       "\n",
			wantBuffer: "A1",
		},
		{
			name:   "case_3",
			fields: fields{buffer: "A\nè1"},
			args: args{
				lowBound:  1,
				highBound: 3,
			},
			want:       "\nè",
			wantBuffer: "A1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := NewRichTextWithText(tt.fields.buffer)
			got := text.deleteFromTo(tt.args.lowBound, tt.args.highBound)
			assert.Equal(t, tt.want, string(got))
			assert.Equal(t, tt.wantBuffer, text.String())
		})
	}
}

func TestText_DeleteFromTo_Segments(t *testing.T) {
	type args struct {
		lowBound  int
		highBound int
	}
	tests := []struct {
		name         string
		segments     []RichTextSegment
		args         args
		want         string
		wantSegments []RichTextSegment
	}{
		{
			name: "remove begin",
			segments: []RichTextSegment{
				&TextSegment{Text: "A\n"},
				&TextSegment{Text: "1"},
			},
			args: args{
				lowBound:  0,
				highBound: 1,
			},
			want: "A",
			wantSegments: []RichTextSegment{
				&TextSegment{Text: "\n"},
				&TextSegment{Text: "1"},
			},
		},
		{
			name: "remove end",
			segments: []RichTextSegment{
				&TextSegment{Text: "A"},
				&TextSegment{Text: "\n1"},
			},
			args: args{
				lowBound:  1,
				highBound: 2,
			},
			want: "\n",
			wantSegments: []RichTextSegment{
				&TextSegment{Text: "A"},
				&TextSegment{Text: "1"},
			},
		},
		{
			name: "remove both",
			segments: []RichTextSegment{
				&TextSegment{Text: "A\n"},
				&TextSegment{Text: "è1"},
			},
			args: args{
				lowBound:  1,
				highBound: 3,
			},
			want: "\nè",
			wantSegments: []RichTextSegment{
				&TextSegment{Text: "A"},
				&TextSegment{Text: "1"},
			},
		},
		{
			name: "remove nontext",
			segments: []RichTextSegment{
				&TextSegment{Text: "A\n"},
				&SeparatorSegment{},
				&TextSegment{Text: "B1"},
			},
			args: args{
				lowBound:  1,
				highBound: 3,
			},
			want: "\nB",
			wantSegments: []RichTextSegment{
				&TextSegment{Text: "A"},
				&TextSegment{Text: "1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := NewRichText(tt.segments...)
			got := text.deleteFromTo(tt.args.lowBound, tt.args.highBound)
			assert.Equal(t, tt.want, string(got))
			for _, s := range tt.wantSegments {
				if txt, ok := s.(*TextSegment); ok {
					txt.parent = text
				}
			}
			assert.Equal(t, tt.wantSegments, text.Segments)
		})
	}
}

func TestText_Multiline(t *testing.T) {
	text := NewRichText(
		&TextSegment{Text: "line1\nli", Style: RichTextStyleStrong},
		&TextSegment{Text: "ne2\nline3", Style: RichTextStyleInline},
	)

	w := test.NewTempWindow(t, text)
	w.Resize(granio.NewSize(64, 90))
	test.AssertImageMatches(t, "richtext/richtext_multiline.png", w.Canvas().Capture())
}

func TestText_Color(t *testing.T) {
	text := NewRichText(trailingBoldErrorSegment())

	assert.Equal(t, theme.Color(theme.ColorNameError), richTextRenderTexts(text)[0].Color)
}

func TestTextRenderer_ApplyTheme(t *testing.T) {
	label := NewLabel("Test\nLine2")
	texts := labelTextRenderTexts(label)

	text1 := texts[0]
	text2 := texts[1]
	textSize1 := text1.TextSize
	textSize2 := text2.TextSize
	customTextSize1 := textSize1
	customTextSize2 := textSize2
	test.WithTestTheme(t, func() {
		label.Refresh()
		text1 := texts[0]
		text2 := texts[1]
		customTextSize1 = text1.TextSize
		customTextSize2 = text2.TextSize
	})

	assert.NotEqual(t, textSize1, customTextSize1)
	assert.NotEqual(t, textSize2, customTextSize2)
}

func TestTextProvider_LineSizeToColumn(t *testing.T) {
	label := NewLabel("Test")
	label.CreateRenderer() // TODO make this a simple refresh call once it's in
	provider := label.provider

	inPad := theme.InnerPadding()
	textSize := theme.TextSize()
	fullSize := provider.lineSizeToColumn(4, 0, textSize, inPad)
	assert.Equal(t, fullSize, provider.lineSizeToColumn(10, 0, textSize, inPad))
	assert.Greater(t, fullSize.Width, provider.lineSizeToColumn(2, 0, textSize, inPad).Width)

	out := provider.lineSizeToColumn(-1, -1, textSize, inPad)
	assert.Equal(t, out, provider.lineSizeToColumn(0, 0, textSize, inPad))
}

func TestText_splitLines(t *testing.T) {
	tests := []struct {
		name string
		text string
		want [][2]int
	}{
		{
			name: "Empty",
			text: "",
			want: [][2]int{
				{0, 0},
			},
		},
		{
			name: "Single",
			text: "foo",
			want: [][2]int{
				{0, 3},
			},
		},
		{
			name: "Multiple",
			text: "foo\nbar",
			want: [][2]int{
				{0, 3},
				{4, 7},
			},
		},
		{
			name: "Multiple_Trailing",
			text: "foo\nbar\n",
			want: [][2]int{
				{0, 3},
				{4, 7},
				{8, 8},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitLines(&TextSegment{Text: tt.text})
			for i, wantRow := range tt.want {
				assert.Equal(t, wantRow[0], got[i].segBegin)
				assert.Equal(t, wantRow[1], got[i].segEnd)
			}
		})
	}
}

func TestText_lineBounds(t *testing.T) {
	measurer := func(text []rune) granio.Size {
		return granio.MeasureText(string(text), 14, granio.TextStyle{})
	}
	tests := []struct {
		name     string
		text     string
		wrap     granio.TextWrap
		trunc    granio.TextTruncation
		want     [][2]int
		ellipses int
	}{
		{
			name: "Empty_WrapOff",
			text: "",
			wrap: granio.TextWrapOff,
			want: [][2]int{
				{0, 0},
			},
		},
		{
			name: "Empty_Truncate",
			text: "",
			wrap: granio.TextWrap(granio.TextTruncateClip),
			want: [][2]int{
				{0, 0},
			},
		},
		{
			name:  "Empty_TruncateClip",
			text:  "",
			trunc: granio.TextTruncateClip,
			want: [][2]int{
				{0, 0},
			},
		},
		{
			name: "Empty_WrapBreak",
			text: "",
			wrap: granio.TextWrapBreak,
			want: [][2]int{
				{0, 0},
			},
		},
		{
			name: "Empty_WrapWord",
			text: "",
			wrap: granio.TextWrapWord,
			want: [][2]int{
				{0, 0},
			},
		},
		{
			name: "Single_Short_WrapOff",
			text: "foobar",
			wrap: granio.TextWrapOff,
			want: [][2]int{
				{0, 6},
			},
		},
		{
			name: "Single_Short_Truncate",
			text: "foobar",
			wrap: granio.TextWrap(granio.TextTruncateClip),
			want: [][2]int{
				{0, 6},
			},
		},
		{
			name:  "Single_Short_TruncateClip",
			text:  "foobar",
			trunc: granio.TextTruncateClip,
			want: [][2]int{
				{0, 6},
			},
		},
		{
			name:  "Single_Short_TruncateEllipsis",
			text:  "foobar",
			trunc: granio.TextTruncateEllipsis,
			want: [][2]int{
				{0, 6},
			},
			ellipses: 0, // too short to cut
		},
		{
			name: "Single_Short_WrapBreak",
			text: "foobar",
			wrap: granio.TextWrapBreak,
			want: [][2]int{
				{0, 6},
			},
		},
		{
			name: "Single_Short_WrapWord",
			text: "foobar",
			wrap: granio.TextWrapWord,
			want: [][2]int{
				{0, 6},
			},
		},
		{
			name: "Single_Long_WrapOff",
			text: "foobar foobar",
			wrap: granio.TextWrapOff,
			want: [][2]int{
				{0, 13},
			},
		},
		{
			name: "Single_Long_Truncate",
			text: "foobar foobar",
			wrap: granio.TextWrap(granio.TextTruncateClip),
			want: [][2]int{
				{0, 10},
			},
		},
		{
			name:  "Single_Long_TruncateClip",
			text:  "foobar foobar",
			trunc: granio.TextTruncateClip,
			want: [][2]int{
				{0, 10},
			},
		},
		{
			name:  "Single_Long_TruncateEllipsis",
			text:  "foobar foobar",
			trunc: granio.TextTruncateEllipsis,
			want: [][2]int{
				{0, 9},
			},
			ellipses: 1,
		},
		{
			name: "Single_Long_WrapBreak",
			text: "foobar foobar",
			wrap: granio.TextWrapBreak,
			want: [][2]int{
				{0, 10},
				{10, 13},
			},
		},
		{
			name: "Single_Long_WrapWord",
			text: "foobar foobar",
			wrap: granio.TextWrapWord,
			want: [][2]int{
				{0, 6},
				{7, 13},
			},
		},
		{
			name: "Multiple_Short_WrapOff",
			text: "foo\nbar",
			wrap: granio.TextWrapOff,
			want: [][2]int{
				{0, 3},
				{4, 7},
			},
		},
		{
			name: "Multiple_Short_Truncate",
			text: "foo\nbar",
			wrap: granio.TextWrap(granio.TextTruncateClip),
			want: [][2]int{
				{0, 3},
				{4, 7},
			},
		},
		{
			name:  "Multiple_Short_TruncateClip",
			text:  "foo\nbar",
			trunc: granio.TextTruncateClip,
			want: [][2]int{
				{0, 3},
				{4, 7},
			},
		},
		{
			name:  "Multiple_Short_TruncateEllipsis",
			text:  "foo\nbar",
			trunc: granio.TextTruncateEllipsis,
			want: [][2]int{
				{0, 3},
				{4, 7},
			},
			ellipses: 0, // too wide
		},
		{
			name: "Multiple_Short_WrapBreak",
			text: "foo\nbar",
			wrap: granio.TextWrapBreak,
			want: [][2]int{
				{0, 3},
				{4, 7},
			},
		},
		{
			name: "Multiple_Short_WrapWord",
			text: "foo\nbar",
			wrap: granio.TextWrapWord,
			want: [][2]int{
				{0, 3},
				{4, 7},
			},
		},
		{
			name: "Multiple_Long_WrapOff",
			text: "foobar\nfoobar foobar foobar\nfoobar foobar",
			wrap: granio.TextWrapOff,
			want: [][2]int{
				{0, 6},
				{7, 27},
				{28, 41},
			},
		},
		{
			name: "Multiple_Long_Truncate",
			text: "foobar\nfoobar foobar foobar\nfoobar foobar",
			wrap: granio.TextWrap(granio.TextTruncateClip),
			want: [][2]int{
				{0, 6},
				{7, 17},
				{28, 38},
			},
		},
		{
			name:  "Multiple_Long_TruncateClip",
			text:  "foobar\nfoobar foobar foobar\nfoobar foobar",
			trunc: granio.TextTruncateClip,
			want: [][2]int{
				{0, 6},
				{7, 17},
				{28, 38},
			},
		},
		{
			name:  "Multiple_Long_TruncateEllipsis",
			text:  "foobar\nfoobar foobar foobar\nfoobar foobar",
			trunc: granio.TextTruncateEllipsis,
			want: [][2]int{
				{0, 6},
				{7, 16},
				{28, 37},
			},
			ellipses: 2,
		},
		{
			name: "Multiple_Long_WrapBreak",
			text: "foobar\nfoobar foobar foobar\nfoobar foobar",
			wrap: granio.TextWrapBreak,
			want: [][2]int{
				{0, 6},
				{7, 17},
				{17, 27},
				{28, 38},
				{38, 41},
			},
		},
		{
			name: "Multiple_Long_WrapWord",
			text: "foobar\nfoobar foobar foobar\nfoobar foobar",
			wrap: granio.TextWrapWord,
			want: [][2]int{
				{0, 6},
				{7, 13},
				{14, 20},
				{21, 27},
				{28, 34},
				{35, 41},
			},
		},
		{
			name: "Multiple_Contiguous_Long_WrapOff",
			text: "foobar\nfoobarfoobarfoobar\nfoobarfoobar\n",
			wrap: granio.TextWrapOff,
			want: [][2]int{
				{0, 6},
				{7, 25},
				{26, 38},
				{39, 39},
			},
		},
		{
			name: "Multiple_Contiguous_Long_Truncate",
			text: "foobar\nfoobarfoobarfoobar\nfoobarfoobar\n",
			wrap: granio.TextWrap(granio.TextTruncateClip),
			want: [][2]int{
				{0, 6},
				{7, 17},
				{26, 36},
				{39, 39},
			},
		},
		{
			name:  "Multiple_Contiguous_Long_TruncateClip",
			text:  "foobar\nfoobarfoobarfoobar\nfoobarfoobar\n",
			trunc: granio.TextTruncateClip,
			want: [][2]int{
				{0, 6},
				{7, 17},
				{26, 36},
				{39, 39},
			},
		},
		{
			name:  "Multiple_Contiguous_Long_TruncateEllipsis",
			text:  "foobar\nfoobarfoobarfoobar\nfoobarfoobar\n",
			trunc: granio.TextTruncateEllipsis,
			want: [][2]int{
				{0, 6},
				{7, 15},
				{26, 34},
				{39, 39},
			},
			ellipses: 2,
		},
		{
			name: "Multiple_Contiguous_Long_WrapBreak",
			text: "foobar\nfoobarfoobarfoobar\nfoobarfoobar\n",
			wrap: granio.TextWrapBreak,
			want: [][2]int{
				{0, 6},
				{7, 17},
				{17, 25},
				{26, 36},
				{36, 38},
				{39, 39},
			},
		},
		{
			name: "Multiple_Contiguous_Long_WrapWord",
			text: "foobar\nfoobarfoobarfoobar\nfoobarfoobar\n",
			wrap: granio.TextWrapWord,
			want: [][2]int{
				{0, 6},
				{7, 17},
				{17, 25},
				{26, 36},
				{36, 38},
				{39, 39},
			},
		},
		{
			name: "Multiple_Trailing_Short_WrapOff",
			text: "foo\nbar\n",
			wrap: granio.TextWrapOff,
			want: [][2]int{
				{0, 3},
				{4, 7},
				{8, 8},
			},
		},
		{
			name: "Multiple_Trailing_Short_Truncate",
			text: "foo\nbar\n",
			wrap: granio.TextWrap(granio.TextTruncateClip),
			want: [][2]int{
				{0, 3},
				{4, 7},
				{8, 8},
			},
		},
		{
			name:  "Multiple_Trailing_Short_TruncateClip",
			text:  "foo\nbar\n",
			trunc: granio.TextTruncateClip,
			want: [][2]int{
				{0, 3},
				{4, 7},
				{8, 8},
			},
		},
		{
			name:  "Multiple_Trailing_Short_TruncateEllipsis",
			text:  "foo\nbar\n",
			trunc: granio.TextTruncateEllipsis,
			want: [][2]int{
				{0, 3},
				{4, 7},
				{8, 8},
			},
		},
		{
			name: "Multiple_Trailing_Short_WrapBreak",
			text: "foo\nbar\n",
			wrap: granio.TextWrapBreak,
			want: [][2]int{
				{0, 3},
				{4, 7},
				{8, 8},
			},
		},
		{
			name: "Multiple_Trailing_Short_WrapWord",
			text: "foo\nbar\n",
			wrap: granio.TextWrapWord,
			want: [][2]int{
				{0, 3},
				{4, 7},
				{8, 8},
			},
		},
		{
			name:  "Multiple_Trailing_Short_WrapTruncateClip",
			text:  "foo\nbar\nbaz",
			wrap:  granio.TextWrapBreak,
			trunc: granio.TextTruncateClip,
			want: [][2]int{
				{0, 3},
				{4, 7},
			},
		},
		{
			name:  "Multiple_Trailing_Short_WrapTruncateEllipsis",
			text:  "foo\nbar\nbaz",
			wrap:  granio.TextWrapBreak,
			trunc: granio.TextTruncateEllipsis,
			want: [][2]int{
				{0, 3},
				{4, 7},
			},
			ellipses: 0,
		},
		{
			name: "Multiple_Trailing_Long_WrapOff",
			text: "foobar\nfoobar foobar foobar\nfoobar foobar\n",
			wrap: granio.TextWrapOff,
			want: [][2]int{
				{0, 6},
				{7, 27},
				{28, 41},
				{42, 42},
			},
		},
		{
			name: "Multiple_Trailing_Long_Truncate",
			text: "foobar\nfoobar foobar foobar\nfoobar foobar\n",
			wrap: granio.TextWrap(granio.TextTruncateClip),
			want: [][2]int{
				{0, 6},
				{7, 17},
				{28, 38},
				{42, 42},
			},
		},
		{
			name:  "Multiple_Trailing_Long_TruncateClip",
			text:  "foobar\nfoobar foobar foobar\nfoobar foobar\n",
			trunc: granio.TextTruncateClip,
			want: [][2]int{
				{0, 6},
				{7, 17},
				{28, 38},
				{42, 42},
			},
		},
		{
			name:  "Multiple_Trailing_Long_TruncateEllipsis",
			text:  "foobar\nfoobar foobar foobar\nfoobar foobar\n",
			trunc: granio.TextTruncateEllipsis,
			want: [][2]int{
				{0, 6},
				{7, 16},
				{28, 37},
				{42, 42},
			},
			ellipses: 2,
		},
		{
			name: "Multiple_Trailing_Long_WrapBreak",
			text: "foobar\nfoobar foobar foobar\nfoobar foobar\n",
			wrap: granio.TextWrapBreak,
			want: [][2]int{
				{0, 6},
				{7, 17},
				{17, 27},
				{28, 38},
				{38, 41},
				{42, 42},
			},
		},
		{
			name: "Multiple_Trailing_Long_WrapWord",
			text: "foobar\nfoobar foobar foobar\nfoobar foobar\n",
			wrap: granio.TextWrapWord,
			want: [][2]int{
				{0, 6},
				{7, 13},
				{14, 20},
				{21, 27},
				{28, 34},
				{35, 41},
				{42, 42},
			},
		},
		{
			name:  "Multiple_Trailing_Long_WrapTruncateClip",
			text:  "foo\nfoobar foobar foobar\nbaz",
			wrap:  granio.TextWrapBreak,
			trunc: granio.TextTruncateClip,
			want: [][2]int{
				{0, 3},
				{4, 14},
			},
		},
		{
			name:  "Multiple_Trailing_Long_WrapTruncateEllipsis",
			text:  "foo\nfoobar foobar foobar\nbazzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			wrap:  granio.TextWrapBreak,
			trunc: granio.TextTruncateEllipsis,
			want: [][2]int{
				{0, 3},
				{4, 14},
				{14, 23},
			},
			ellipses: 1,
		},
		{
			name:  "Multiple_Trailing_Long_WrapWordTruncateEllipsis",
			text:  "foo\nfoobar foobar foobar\nbaz",
			wrap:  granio.TextWrapWord,
			trunc: granio.TextTruncateEllipsis,
			want: [][2]int{
				{0, 3},
				{4, 10},
				{11, 17},
			},
			ellipses: 1,
		},
		{
			name:  "Multi_byte_ellipsis_not_truncated",
			text:  "🪃 234",
			trunc: granio.TextTruncateEllipsis,
			wrap:  granio.TextWrapOff,
			want: [][2]int{
				{0, 5},
			},
			ellipses: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ellipses := 0
			richText := NewRichTextWithText(tt.text)
			richText.Wrapping = tt.wrap
			richText.Truncation = tt.trunc
			got, _ := lineBounds(richText, richText.Segments[0], 76, granio.NewSize(76, 64), measurer)
			for i, wantRow := range tt.want {
				assert.Equal(t, wantRow[0], got[i].segBegin)
				assert.Equal(t, wantRow[1], got[i].segEnd)

				if got[i].ellipsis {
					ellipses++
				}
			}
			assert.Equal(t, tt.ellipses, ellipses)
		})
	}
}

func TestText_lineBounds_hyperlinks(t *testing.T) {
	measurer := func(text []rune) granio.Size {
		return granio.MeasureText(string(text), 14, granio.TextStyle{})
	}
	tests := []struct {
		name     string
		text     string
		wrap     granio.TextWrap
		trunc    granio.TextTruncation
		want     [][2]int
		ellipses int
	}{
		{
			name: "Hyperlink_Truncate",
			text: "this is a long link",
			wrap: granio.TextWrap(granio.TextTruncateClip),
			want: [][2]int{
				{0, 9},
			},
		},
		{
			name:  "Hyperlink_TruncateClip",
			text:  "this is a long link",
			trunc: granio.TextTruncateClip,
			want: [][2]int{
				{0, 9},
			},
		},
		{
			name:  "Hyperlink_TruncateEllipsis",
			text:  "this is a long link",
			trunc: granio.TextTruncateEllipsis,
			want: [][2]int{
				{0, 8},
			},
			ellipses: 1,
		},
		{
			name: "Hyperlink_WrapBreak",
			text: "this is a long link",
			wrap: granio.TextWrapBreak,
			want: [][2]int{
				{0, 9},
				{9, 17},
			},
		},
		{
			name: "Hyperlink_WrapWord",
			text: "this is a long link",
			wrap: granio.TextWrapWord,
			want: [][2]int{
				{0, 9},
				{10, 14},
				{15, 19},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ellipses := 0
			u, _ := url.Parse("https://fyne.io/")
			richText := NewRichText(&HyperlinkSegment{Text: tt.text, URL: u})
			richText.Wrapping = tt.wrap
			richText.Truncation = tt.trunc
			got, _ := lineBounds(richText, richText.Segments[0], 50, granio.NewSize(50, 64), measurer)
			for i, wantRow := range tt.want {
				assert.Equal(t, wantRow[0], got[i].segBegin)
				assert.Equal(t, wantRow[1], got[i].segEnd)

				if got[i].ellipsis {
					ellipses++
				}
			}
			assert.Equal(t, tt.ellipses, ellipses)
		})
	}
}

func TestText_lineBounds_variable_char_width(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		wrap  granio.TextWrap
		trunc granio.TextTruncation
		want  [][2]int
	}{
		{
			name: "IM_WrapOff",
			text: "iiiiiiiiiimmmmmmmmmm",
			wrap: granio.TextWrapOff,
			want: [][2]int{
				{0, 20},
			},
		},
		{
			name: "IM_Truncate",
			text: "iiiiiiiiiimmmmmmmmmm",
			wrap: granio.TextWrap(granio.TextTruncateClip),
			want: [][2]int{
				{0, 12},
			},
		},
		{
			name:  "IM_TruncateClip",
			text:  "iiiiiiiiiimmmmmmmmmm",
			trunc: granio.TextTruncateClip,
			want: [][2]int{
				{0, 12},
			},
		},
		{
			name:  "IM_TruncateEllipsis",
			text:  "iiiiiiiiiimmmmmmmmmm",
			trunc: granio.TextTruncateEllipsis,
			want: [][2]int{
				{0, 9},
			},
		},
		{
			name: "IM_WrapBreak",
			text: "iiiiiiiiiimmmmmmmmmm",
			wrap: granio.TextWrapBreak,
			want: [][2]int{
				{0, 12},
				{12, 16},
				{16, 20},
			},
		},
		{
			name: "IM_WrapWord",
			text: "iiiiiiiiiimmmmmmmmmm",
			wrap: granio.TextWrapWord,
			want: [][2]int{
				{0, 12},
				{12, 16},
				{16, 20},
			},
		},
	}
	textSize := float32(10)
	textStyle := granio.TextStyle{}
	measurer := func(text []rune) granio.Size {
		return granio.MeasureText(string(text), textSize, textStyle)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			richText := NewRichTextWithText(tt.text)
			richText.Wrapping = tt.wrap
			richText.Truncation = tt.trunc
			got, _ := lineBounds(richText, richText.Segments[0], 46, granio.NewSize(46, 184), measurer)
			for i, wantRow := range tt.want {
				assert.Equal(t, wantRow[0], got[i].segBegin)
				assert.Equal(t, wantRow[1], got[i].segEnd)
			}
		})
	}
}

func TestText_lineBounds_small_firstWidth(t *testing.T) {
	measurer := func(text []rune) granio.Size {
		return granio.MeasureText(string(text), 14, granio.TextStyle{})
	}
	richText := NewRichTextWithText("foobar")
	richText.Wrapping = granio.TextWrapWord
	richText.Truncation = granio.TextTruncateOff
	got, _ := lineBounds(richText, richText.Segments[0], 0.1, granio.NewSize(64, 20), measurer)
	assert.Equal(t, 0, got[0].segBegin)
	assert.Equal(t, 0, got[0].segEnd)
	assert.Equal(t, 0, got[1].segBegin)
	assert.Equal(t, 6, got[1].segEnd)
}

func TestText_Truncate_Segments(t *testing.T) {
	for name, tt := range map[string]struct {
		trunc granio.TextTruncation
		width float32
		want  []string
	}{
		"Clip":                     {granio.TextTruncateClip, 120, []string{"Normal ", "Bold", " a", "second line of t"}},
		"ClipInEarlierSegment":     {granio.TextTruncateClip, 93, []string{"Normal ", "Bol", "second line"}},
		"Ellipsis":                 {granio.TextTruncateEllipsis, 120, []string{"Normal ", "Bold", " …", "second line of …"}},
		"EllipsisNoRoomLeft":       {granio.TextTruncateEllipsis, 105, []string{"Normal ", "Bol…", "second line …"}},
		"EllipsisInEarlierSegment": {granio.TextTruncateEllipsis, 93, []string{"Normal ", "B…", "second li…"}},
	} {
		t.Run(name, func(t *testing.T) {
			text := NewRichText(
				&TextSegment{Text: "Normal ", Style: RichTextStyleInline},
				&TextSegment{Text: "Bold", Style: RichTextStyleStrong},
				&TextSegment{Text: " and more\nsecond line of text", Style: RichTextStyleInline},
			)
			text.Truncation = tt.trunc
			text.Resize(granio.NewSize(tt.width, 100))

			var got []string
			for _, obj := range richTextRenderTexts(text) {
				got = append(got, obj.Text)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestText_Truncate_ParagraphSpacing(t *testing.T) {
	md := "Normal **Bold** and some text, which is long enough to be truncated.\n\n> A quote"
	for _, trunc := range []granio.TextTruncation{granio.TextTruncateClip, granio.TextTruncateEllipsis} {
		full := NewRichTextFromMarkdown(md)
		full.Truncation = trunc
		full.Resize(granio.NewSize(1000, 100))
		text := NewRichTextFromMarkdown(md)
		text.Truncation = trunc
		text.Resize(granio.NewSize(200, 100))

		// the quote should not move up when the paragraph above it is truncated
		assert.Len(t, text.rowBounds, len(full.rowBounds))
		wantY, _ := full.rowGeometry(1)
		gotY, _ := text.rowGeometry(1)
		assert.Equal(t, wantY, gotY, "truncation %d", trunc)
		assert.Equal(t, full.MinSize().Height, text.MinSize().Height, "truncation %d", trunc)
	}
}

func TestText_WrapTruncate_Segments(t *testing.T) {
	md := "Normal **Bold** *Italic* [Link](https://fyne.io/) and some `Code`.\n" +
		"This styled row should also wrap as expected, but only *when required*."
	full := NewRichTextFromMarkdown(md)
	full.Wrapping = granio.TextWrapWord
	full.Resize(granio.NewSize(200, 1000))
	want := richTextRowTexts(full)

	for _, trunc := range []granio.TextTruncation{granio.TextTruncateClip, granio.TextTruncateEllipsis} {
		for height := float32(0); height <= 160; height += 5 {
			text := NewRichTextFromMarkdown(md)
			text.Wrapping = granio.TextWrapWord
			text.Truncation = trunc
			if height < text.MinSize().Height {
				continue // too short to show any of the text
			}
			text.Resize(granio.NewSize(200, height))

			// the rows shown should be the start of the full text, ending where it was truncated
			got := richTextRowTexts(text)
			last := len(got) - 1
			if !assert.Positive(t, len(got)) || !assert.LessOrEqual(t, len(got), len(want)) {
				continue
			}
			assert.Equal(t, want[:last], got[:last], "truncation %d at height %v", trunc, height)
			assert.True(t, strings.HasPrefix(want[last], strings.TrimSuffix(got[last], ellipsisChar)),
				"truncation %d at height %v: %q is not the start of %q", trunc, height, got[last], want[last])
			if trunc == granio.TextTruncateEllipsis && len(got) < len(want) {
				assert.True(t, strings.HasSuffix(got[last], ellipsisChar), "truncation %d at height %v", trunc, height)
			}
		}
	}
}

func TestText_howManyRunesFit(t *testing.T) {
	maxWidth := float32(46)
	textSize := float32(10)
	textStyle := granio.TextStyle{}
	measurer := func(text []rune) granio.Size {
		return granio.MeasureText(string(text), textSize, textStyle)
	}
	charWidth := measurer([]rune("z")).Width
	for name, tt := range map[string]struct {
		text string
		want int
	}{
		"IM": {
			text: "iiiiiiiiiimmmmmmmmmm",
			want: 12,
		},
		"Single_Line": {
			text: "foobar foobar",
			want: 9,
		},
		"WH": {
			text: "wwwww hhhhhh",
			want: 6,
		},
		"DS": {
			text: "dddddd sssssss",
			want: 8,
		},
		"DI": {
			text: "dididi dididd",
			want: 10,
		},
		"XW": {
			text: "xwxwxwxw xwxw",
			want: 7,
		},
		"W": {
			text: "WWWWW",
			want: 4,
		},
		"Empty": {
			text: "",
			want: 0,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, howManyRunesFit([]rune(tt.text), maxWidth, charWidth, measurer))
		})
	}
	t.Run("Zero width", func(t *testing.T) {
		assert.Equal(t, 0, howManyRunesFit([]rune("foo"), 0, charWidth, measurer))
	})
	t.Run("Negative width", func(t *testing.T) {
		assert.Equal(t, 0, howManyRunesFit([]rune("foo"), -1, charWidth, measurer))
	})
}

func TestText_findSpaceIndex(t *testing.T) {
	for name, tt := range map[string]struct {
		text string
		want int
	}{
		"no_space_fallback": {
			text: "iiiiiiiiiimmmmmmmmmm",
			want: -1,
		},
		"single_space": {
			text: "foobar foobar",
			want: 6,
		},
		"double_space": {
			text: "ww wwww www",
			want: 7,
		},
		"many_spaces": {
			text: "ww wwww www wwwww",
			want: 11,
		},
		"space beginning": {
			text: " ww",
			want: 0,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, findSpaceIndex([]rune(tt.text), len(tt.text)-1))
		})
	}
}
