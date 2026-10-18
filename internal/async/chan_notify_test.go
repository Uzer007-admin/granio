package async_test

import (
	"github.com/Uzer007-admin/granio/v2/internal/async"
	"testing"
	"time"
)

func TestUnboundedNotificationPublishesReadyWork(t *testing.T) {
	ready := make(chan struct{}, 1)
	queue := async.NewUnboundedChanWithNotify[int](func() {
		select {
		case ready <- struct{}{}:
		default:
		}
	})
	defer queue.Close()
	queue.In() <- 42
	select {
	case <-ready:
	case <-time.After(time.Second):
		t.Fatal("output did not notify its waiting consumer")
	}
	select {
	case got := <-queue.Out():
		if got != 42 {
			t.Fatalf("got %d, want 42", got)
		}
	default:
		t.Fatal("notification preceded publishing the output")
	}
}

func TestUnboundedCloseDrainsLargeQueue(t *testing.T) {
	queue := async.NewUnboundedChan[int]()
	const count = 4096
	for i := 0; i < count; i++ {
		queue.In() <- i
	}
	queue.Close()
	queue.Close() // Closing twice must not block or race a closed signal channel.
	received := 0
	for got := range queue.Out() {
		if got != received {
			t.Fatalf("FIFO after close: got %d, want %d", got, received)
		}
		received++
	}
	if received != count {
		t.Fatalf("close discarded accepted work: got %d, want %d", received, count)
	}
}
