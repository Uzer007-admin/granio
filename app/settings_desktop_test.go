//go:build !android && !ios && !mobile

package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Uzer007-admin/granio/v2"
)

func TestEnsureDir(t *testing.T) {
	tmpDir := testPath("fynetest")

	ensureDirExists(tmpDir)
	if st, err := os.Stat(tmpDir); err != nil || !st.IsDir() {
		t.Error("Could not ensure directory exists")
	}

	os.Remove(tmpDir)
}

func TestWatchSettings(t *testing.T) {
	settings := &settings{}
	listener := make(chan granio.Settings, 1)
	settings.AddChangeListener(listener)

	settings.fileChanged() // simulate the settings file changing

	select {
	case <-listener:
	case <-time.After(100 * time.Millisecond):
		t.Error("Settings listener was not called")
	}
}

func TestWatchFile(t *testing.T) {
	path := testPath(t.Name() + ".txt")
	require.NoError(t, os.WriteFile(path, nil, 0o600))
	defer os.Remove(path)

	called := make(chan any, 1)
	watcher := watchFile(path, func() {
		select {
		case called <- true:
		default:
		}
	})
	require.NotNil(t, watcher)
	defer watcher.Close()
	file, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	file.WriteString(" ")
	file.Close()

	select {
	case <-called:
	case <-time.After(5 * time.Second):
		t.Error("File watcher callback was not called")
	}
}

func TestFileWatcher_FileDeleted(t *testing.T) {
	path := testPath(t.Name() + ".txt")
	require.NoError(t, os.WriteFile(path, nil, 0o600))
	defer os.Remove(path)

	called := make(chan any, 1)
	watcher := watchFile(path, func() {
		select {
		case called <- true:
		default:
		}
	})
	if watcher == nil {
		assert.Fail(t, "Could not start watcher")
		return
	}

	defer watcher.Close()
	require.NoError(t, os.Remove(path))

	select {
	case <-called:
	case <-time.After(5 * time.Second):
		t.Fatal("File watcher callback was not called after deletion")
	}
	require.NoError(t, os.WriteFile(path, []byte("replacement"), 0o600))
	select {
	case <-called:
	case <-time.After(5 * time.Second):
		t.Fatal("File watcher callback was not called after recreation")
	}
}

func TestWatchFileIgnoresSibling(t *testing.T) {
	path := testPath(t.Name() + ".txt")
	sibling := path + ".other"
	require.NoError(t, os.WriteFile(path, nil, 0o600))
	defer os.Remove(path)
	defer os.Remove(sibling)
	called := make(chan struct{}, 1)
	watcher := watchFile(path, func() {
		select {
		case called <- struct{}{}:
		default:
		}
	})
	require.NotNil(t, watcher)
	defer watcher.Close()
	require.NoError(t, os.WriteFile(sibling, []byte("unrelated"), 0o600))
	select {
	case <-called:
		t.Fatal("Unrelated file triggered the settings callback")
	case <-time.After(200 * time.Millisecond):
	}
}

func testPath(child string) string {
	// TMPDIR would be more normal but fsnotify cannot watch that on macOS...
	return filepath.Join("testdata", child)
}
