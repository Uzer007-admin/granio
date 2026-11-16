package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/dialog"
	"github.com/Uzer007-admin/granio/v3/driver"
)

const workspaceRendererPreference = "workspace-renderer"

func configureGalleryRenderer(a granio.App, flag string, workspace bool) error {
	value := flag
	if value == "" && workspace && strings.TrimSpace(os.Getenv("GRANIO_RENDERER")) == "" {
		value = a.Preferences().String(workspaceRendererPreference)
	}
	if value == "" {
		return nil
	}
	var mode driver.RenderMode
	switch value {
	case "auto":
		mode = driver.RenderAuto
	case "opengl":
		mode = driver.RenderOpenGL
	case "software":
		mode = driver.RenderSoftware
	default:
		if flag != "" {
			return fmt.Errorf("неизвестный режим графики %q, допустимы auto, opengl и software", flag)
		}
		return nil
	}
	rendering, ok := a.Driver().(driver.RenderingDriver)
	if !ok {
		if flag == "" || mode == driver.RenderAuto {
			return nil
		}
		return driver.ErrRenderModeUnavailable
	}
	err := rendering.SetRenderMode(mode)
	if flag == "" && errors.Is(err, driver.ErrRenderModeUnavailable) {
		return nil
	}
	return err
}

func workspaceRenderMode(a granio.App, w granio.Window) driver.RenderMode {
	if window, ok := w.(driver.RenderingWindow); ok {
		return window.RenderMode()
	}
	if rendering, ok := a.Driver().(driver.RenderingDriver); ok {
		return rendering.RenderMode()
	}
	return driver.RenderAuto
}

func workspaceRendererName(mode driver.RenderMode) string {
	switch mode {
	case driver.RenderOpenGL:
		if runtime.GOOS == "android" || runtime.GOOS == "ios" {
			return "OpenGL ES"
		}
		return "OpenGL"
	case driver.RenderSoftware:
		return "CPU"
	default:
		return "Система"
	}
}

func (v *workspaceView) rendererMenu() {
	a := granio.CurrentApp()
	if _, ok := a.Driver().(driver.RenderingDriver); !ok {
		dialog.ShowInformation("Графика", "Режим отрисовки выбирает система.", v.w)
		return
	}
	selectMode := func(mode, caption string) {
		a.Preferences().SetString(workspaceRendererPreference, mode)
		dialog.ShowInformation("Графика", caption+" будет выбран при следующем запуске без параметров.", v.w)
	}
	items := []*granio.MenuItem{
		granio.NewMenuItem("Автоматически", func() { selectMode("auto", "Автоматический режим") }),
		granio.NewMenuItem("OpenGL", func() { selectMode("opengl", "OpenGL") }),
	}
	switch runtime.GOOS {
	case "windows", "linux", "android":
		items = append(items, granio.NewMenuItem("CPU · без аппаратного ускорения", func() { selectMode("software", "Режим CPU") }))
	}
	v.menu(v.rendererButton, items...)
}
