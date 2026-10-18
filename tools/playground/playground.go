// Package playground provides tooling for running fyne applications inside the Go playground.
package playground

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/software"
	"github.com/Uzer007-admin/granio/v2/internal/test"
	"github.com/Uzer007-admin/granio/v2/theme"
)

func encodeImage(img image.Image) (string, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func imageToPlayground(img image.Image) {
	enc, err := encodeImage(img)
	if err != nil {
		granio.LogError("Failed to encode image", err)
		return
	}

	fmt.Println("IMAGE:" + enc)
}

// RenderCanvas takes a canvas and converts it into an inline image for showing in the playground
func RenderCanvas(c granio.Canvas) {
	imageToPlayground(software.RenderCanvas(c, test.DarkTheme(theme.DefaultTheme())))
}

// RenderWindow takes a window and converts its canvas into an inline image for showing in the playground
func RenderWindow(w granio.Window) {
	RenderCanvas(w.Canvas())
}

// Render takes a canvasobject and converts it into an inline image for showing in the playground
func Render(obj granio.CanvasObject) {
	imageToPlayground(software.Render(obj, test.DarkTheme(theme.DefaultTheme())))
}
