package embedded_test

import (
	"image"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/embedded"
	intNoos "github.com/Uzer007-admin/granio/v2/internal/driver/embedded"
	_ "github.com/Uzer007-admin/granio/v2/test" // for test app initialization
)

func TestNoOSDriver(t *testing.T) {
	count := 0
	render := func(image.Image) {
		count++
	}
	queue := make(chan embedded.Event, 1)
	size := func() granio.Size {
		return granio.NewSquareSize(100)
	}
	d := intNoos.NewNoOSDriver(render, func(f func()) { f() }, queue, size)

	go func() {
		time.Sleep(100 * time.Millisecond)
		close(queue)
	}()
	d.Run()
	assert.Equal(t, 0, count)

	queue = make(chan embedded.Event, 1)
	d = intNoos.NewNoOSDriver(render, func(f func()) { f() }, queue, size)
	_ = d.CreateWindow("Test")
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(queue)
	}()
	d.Run()
	assert.Equal(t, 1, count)

	count = 0
	queue = make(chan embedded.Event, 1)
	d = intNoos.NewNoOSDriver(render, func(f func()) { f() }, queue, size)
	w := d.CreateWindow("Test")
	keyed := make(chan bool)
	w.Canvas().SetOnTypedKey(func(*granio.KeyEvent) {
		keyed <- true
	})
	go func() {
		queue <- &embedded.CharacterEvent{
			Rune: 'a',
		}
		queue <- &embedded.KeyEvent{
			Name: granio.KeyEscape,
		}
		<-keyed
		close(queue)
	}()
	d.Run()
	assert.Equal(t, 3, count)
}
