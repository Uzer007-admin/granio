//go:build accessibility && windows && !no_glfw && !mobile

package glfw

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/require"
)

func TestWindowsAccessibilityProviderLifetime(t *testing.T) {
	runOnMain(func() {
		if !windowsAccessibilityContractTest() {
			t.Error("UIA ownership, Value pattern, or independent windows failed")
		}
	})
}

func TestWindowsAccessibilityOffscreenHierarchy(t *testing.T) {
	w := createWindowWithResizeCallback("UIA complete logical hierarchy")
	defer w.Close()
	valid := false
	var selected []int
	runOnMain(func() {
		list := widget.NewList(func() int { return 100000 }, func() granio.CanvasObject { return widget.NewLabel("") }, func(id widget.ListItemID, object granio.CanvasObject) {
			object.(*widget.Label).SetText(fmt.Sprintf("Item %d", id))
		})
		list.SelectionMode = widget.SelectionMultiple
		after := widget.NewEntry()
		after.PlaceHolder = "After collection"
		w.window.SetContent(container.NewVBox(widget.NewLabel("Before collection"), list, after))
		w.window.Resize(granio.NewSize(360, 180))
		w.window.Show()
		list.SetSelection([]int{90000})
		w.window.updateAccessibility()
		hwnd := uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
		stored, _ := winAccessibleWindows.Load(hwnd)
		state := stored.(*winAccessibility)
		state.mu.RLock()
		id := state.ids[accessibleKey{object: list}]
		state.mu.RUnlock()
		valid = windowsAccessibilityVirtualContractTest(hwnd, id)
		selected = list.Selected()
	})
	require.True(t, valid, "real COM logical navigation / offscreen selection / virtual pattern")
	require.Equal(t, []int{90000}, selected)
}

func TestWindowsAccessibilityRecycledSelection(t *testing.T) {
	w := createWindowWithResizeCallback("UIA virtual selection fixture")
	defer w.Close()
	var list *widget.List
	runOnMain(func() {
		list = widget.NewList(func() int { return 1000 }, func() granio.CanvasObject { return widget.NewLabel("") },
			func(id widget.ListItemID, object granio.CanvasObject) {
				object.(*widget.Label).SetText(fmt.Sprintf("Item %d", id))
			})
		list.SelectionMode = widget.SelectionMultiple
		w.window.SetContent(list)
		w.window.Resize(granio.NewSize(360, 180))
		w.window.Show()
		list.ScrollTo(90)
		list.SetSelection([]widget.ListItemID{90})
	})
	valid := false
	require.Eventually(t, func() bool {
		runOnMain(func() {
			w.window.updateAccessibility()
			hwnd := uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
			stored, _ := winAccessibleWindows.Load(hwnd)
			state := stored.(*winAccessibility)
			state.mu.RLock()
			id, child := state.ids[accessibleKey{object: list}], uint64(0)
			for key, node := range state.nodes {
				if node.selectionOwner == list && node.selectionIndex == 90 {
					child = key
					break
				}
			}
			state.mu.RUnlock()
			valid = child != 0 && windowsAccessibilitySelectionContractTest(hwnd, id, child)
		})
		return valid
	}, time.Second, 10*time.Millisecond)
}

func TestWindowsAccessibilityTextRangeABI(t *testing.T) {
	success := false
	runOnMain(func() { success = windowsAccessibilityTextContractTest() })
	if !success {
		t.Fatal("UIA Text/Range Unicode movement or detached lifetime failed")
	}
}

func TestWindowsAccessibilityEventABI(t *testing.T) {
	success := false
	runOnMain(func() { success = windowsAccessibilityEventsContractTest() })
	require.True(t, success, "UIA property VARIANT types or reentrant provider reads failed")
}

func TestWindowsAccessibilityScrollPreservesSelection(t *testing.T) {
	w := createWindowWithResizeCallback("UIA text scroll fixture")
	defer w.Close()
	var entry *widget.Entry
	var hwnd uintptr
	var id uint64
	runOnMain(func() {
		entry = widget.NewMultiLineEntry()
		entry.SetText(strings.Repeat("alpha 日本\n", 80))
		w.window.SetContent(entry)
		w.window.Resize(granio.NewSize(360, 180))
		w.window.Show()
		entry.AccessibilitySetSelection(1, 3)
		w.window.updateAccessibility()
		hwnd = uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
		stored, _ := winAccessibleWindows.Load(hwnd)
		state := stored.(*winAccessibility)
		state.mu.RLock()
		id = state.ids[accessibleKey{object: entry}]
		state.mu.RUnlock()
	})
	require.NotZero(t, id)
	applied := false
	runOnMain(func() { applied = windowsAccessibilityScrollRangeTest(hwnd, id, 180, 200) })
	require.True(t, applied)
	require.Eventually(t, func() bool {
		visible := false
		runOnMain(func() { visible = len(entry.AccessibilityTextRectangles(180, 200)) > 0 })
		return visible
	}, time.Second, time.Millisecond)
	var start, end int
	runOnMain(func() { _, start, end = entry.AccessibilityTextSelection() })
	require.Equal(t, []int{1, 3}, []int{start, end})
}
