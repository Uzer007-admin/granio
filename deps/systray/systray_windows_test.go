//go:build windows

package systray

import (
	"io/ioutil"
	"runtime"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const iconFilePath = "example/icon/iconwin.ico"

func TestBaseWindowsTray(t *testing.T) {
	systrayReady = func() {}
	systrayExit = func() {}
	systrayExitCalled = false

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := wt.initInstance(); err != nil {
		t.Fatalf("initInstance failed: %s", err)
	}

	if err := wt.createMenu(); err != nil {
		t.Fatalf("createMenu failed: %s", err)
	}
	wt.initialized.Store(true)
	defer wt.initialized.Store(false)

	defer func() {
		pDestroyWindow.Call(uintptr(wt.window))
		wt.wcex.unregister()
	}()

	if err := wt.setIcon(iconFilePath); err != nil {
		t.Errorf("SetIcon failed: %s", err)
	}

	if err := wt.setTooltip("Cyrillic tooltip тест:)"); err != nil {
		t.Errorf("SetIcon failed: %s", err)
	}

	err := wt.addOrUpdateMenuItem(1, 0, "Simple enabled", "Ctrl+E", false, false)
	if err != nil {
		t.Errorf("mergeMenuItem failed: %s", err)
	}
	err = wt.addOrUpdateMenuItem(2, 0, "Simple disabled", "", true, false)
	if err != nil {
		t.Errorf("mergeMenuItem failed: %s", err)
	}
	err = wt.addSeparatorMenuItem(3, 0)
	if err != nil {
		t.Errorf("addSeparatorMenuItem failed: %s", err)
	}
	err = wt.addOrUpdateMenuItem(4, 0, "Simple checked enabled", "", false, true)
	if err != nil {
		t.Errorf("mergeMenuItem failed: %s", err)
	}
	err = wt.addOrUpdateMenuItem(5, 0, "Simple checked disabled", "", true, true)
	if err != nil {
		t.Errorf("mergeMenuItem failed: %s", err)
	}

	err = wt.hideMenuItem(1, 0)
	if err != nil {
		t.Errorf("hideMenuItem failed: %s", err)
	}
	if wt.getVisibleItemIndex(0, 1) != -1 {
		t.Error("hidden item is still marked visible")
	}
	_ = wt.hideMenuItem(100, 0)
	if wt.getVisibleItemIndex(0, 2) == -1 {
		t.Error("hiding an unknown item removed an existing entry")
	}

	err = wt.addOrUpdateMenuItem(2, 0, "Simple disabled update", "", true, false)
	if err != nil {
		t.Errorf("mergeMenuItem failed: %s", err)
	}

	time.AfterFunc(1*time.Second, quit)

	m := struct {
		WindowHandle windows.Handle
		Message      uint32
		Wparam       uintptr
		Lparam       uintptr
		Time         uint32
		Pt           point
	}{}
	for {
		ret, _, err := pGetMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		res := int32(ret)
		if res == -1 {
			t.Errorf("win32 GetMessage failed: %v", err)
			return
		} else if res == 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func TestWindowsRun(t *testing.T) {
	onReady := func() {
		b, err := ioutil.ReadFile(iconFilePath)
		if err != nil {
			t.Fatalf("Can't load icon file: %v", err)
		}
		SetIcon(b)
		SetTitle("Test title с кириллицей")

		bSomeBtn := AddMenuItem("Йа кнопко", "")
		bSomeBtn.Check()
		AddSeparator()
		bQuit := AddMenuItem("Quit", "Quit the whole app")
		go func() {
			<-bQuit.ClickedCh
			t.Log("Quit reqested")
			Quit()
		}()
		time.AfterFunc(1*time.Second, Quit)
	}

	onExit := func() {
		t.Log("Exit success")
	}

	Run(onReady, onExit)
}
