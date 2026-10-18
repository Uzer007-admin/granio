package app

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLifecycleCloseDrainsAndRejectsLateEvents(t *testing.T) {
	var lifecycle Lifecycle
	lifecycle.InitEventQueue()
	var got []int
	go lifecycle.RunEventQueue(func(fn func(), _ bool) { fn() })
	<-lifecycle.EventQueueReady()
	for i := 0; i < 4096; i++ {
		value := i
		lifecycle.QueueEvent(func() { got = append(got, value) })
	}
	lifecycle.DestroyEventQueue()
	lifecycle.WaitForEvents()
	lifecycle.DestroyEventQueue()
	require.Len(t, got, 4096)
	for i, value := range got {
		require.Equal(t, i, value)
	}
	lifecycle.QueueEvent(func() { t.Error("accepted event after lifecycle close") })
}
func TestLifecycleCloseConcurrentSenders(t *testing.T) {
	var lifecycle Lifecycle
	lifecycle.InitEventQueue()
	go lifecycle.RunEventQueue(func(fn func(), _ bool) { fn() })
	<-lifecycle.EventQueueReady()
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for j := 0; j < 1000; j++ {
				lifecycle.QueueEvent(func() {})
			}
		}()
	}
	lifecycle.DestroyEventQueue()
	workers.Wait()
	lifecycle.WaitForEvents()
}
