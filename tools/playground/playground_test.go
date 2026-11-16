package playground

import (
	"encoding/base64"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/driver/software"
	"github.com/Uzer007-admin/granio/v3/internal/test"
	_ "github.com/Uzer007-admin/granio/v3/test" // for test app initialization
	"github.com/Uzer007-admin/granio/v3/theme"
)

func TestRender(t *testing.T) {
	obj := canvas.NewRectangle(color.Black)
	obj.SetMinSize(granio.NewSquareSize(10))

	img := software.Render(obj, test.DarkTheme(theme.DefaultTheme()))
	assert.NotNil(t, img)

	enc, err := encodeImage(img)
	assert.NoError(t, err)
	assert.Equal(t, encodedRect, enc)

	bytes, err := base64.StdEncoding.DecodeString(enc)
	assert.NoError(t, err)
	assert.Equal(t, "PNG", string(bytes)[1:4])
}
