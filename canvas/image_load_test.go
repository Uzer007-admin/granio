package canvas_test

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/stretchr/testify/require"
)

func TestLoadImageLimitsAndHTTPFailures(t *testing.T) {
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, image.NewNRGBA(image.Rect(0, 0, 8, 8))))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/failed" {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write(encoded.Bytes())
	}))
	defer server.Close()
	uri, err := storage.ParseURI(server.URL + "/image.png")
	require.NoError(t, err)
	loaded, err := canvas.LoadImage(context.Background(), uri)
	require.NoError(t, err)
	require.Equal(t, image.Rect(0, 0, 8, 8), loaded.Image.Bounds())
	_, err = canvas.LoadImageWithOptions(context.Background(), uri, canvas.ImageLoadOptions{MaxEncodedBytes: 10})
	require.Error(t, err)
	_, err = canvas.LoadImageWithOptions(context.Background(), uri, canvas.ImageLoadOptions{MaxPixels: 63})
	require.Error(t, err)
	failed, _ := storage.ParseURI(server.URL + "/failed")
	_, err = canvas.LoadImage(context.Background(), failed)
	require.ErrorContains(t, err, "503")
}

func TestLoadImageCancelsPendingHTTP(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	defer server.Close()
	uri, _ := storage.ParseURI(server.URL + "/image.png")
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { _, err := canvas.LoadImage(ctx, uri); result <- err }()
	<-started
	cancel()
	select {
	case err := <-result:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("image request ignored cancellation")
	}
}
