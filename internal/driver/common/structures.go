package common

import (
	"sync"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/internal"
)

const retainedRefreshObjects = 4096

type deduplicatedObjectQueue struct {
	mu      sync.Mutex
	pending []granio.CanvasObject
	spare   []granio.CanvasObject
	dedup   map[granio.CanvasObject]struct{}
}

// In adds an object to the queue if it is not already present.
func (q *deduplicatedObjectQueue) In(obj granio.CanvasObject) {
	if obj == nil {
		return
	}
	q.mu.Lock()
	if _, exists := q.dedup[obj]; !exists {
		if q.dedup == nil {
			q.dedup = make(map[granio.CanvasObject]struct{})
		}
		q.dedup[obj] = struct{}{}
		q.pending = append(q.pending, obj)
	}
	q.mu.Unlock()
}

// drain takes the current batch before calling into the painter. Objects
// refreshed in the meantime enter the next batch, including ones just drained.
func (q *deduplicatedObjectQueue) drain(consume func(granio.CanvasObject)) uint64 {
	q.mu.Lock()
	batch := q.pending
	if len(batch) == 0 {
		q.mu.Unlock()
		return 0
	}
	q.pending, q.spare = q.spare[:0], nil
	if len(batch) > retainedRefreshObjects {
		q.dedup = nil
	} else {
		clear(q.dedup)
	}
	q.mu.Unlock()

	for _, object := range batch {
		consume(object)
	}
	if cap(batch) <= retainedRefreshObjects {
		clear(batch)
		q.mu.Lock()
		q.spare = batch[:0]
		q.mu.Unlock()
	}
	return uint64(len(batch))
}

type renderCacheTree struct {
	root *RenderCacheNode
}

// RenderCacheNode represents a node in a render cache tree.
type RenderCacheNode struct {
	// structural data
	firstChild  *RenderCacheNode
	nextSibling *RenderCacheNode
	obj         granio.CanvasObject
	parent      *RenderCacheNode
	// cache data
	minSize granio.Size
}

// Obj returns the node object.
func (r *RenderCacheNode) Obj() granio.CanvasObject {
	return r.obj
}

type overlayStack struct {
	internal.OverlayStack

	renderCaches []*renderCacheTree
}

func (o *overlayStack) Add(overlay granio.CanvasObject) {
	if overlay == nil {
		return
	}

	o.renderCaches = append(o.renderCaches, &renderCacheTree{root: &RenderCacheNode{obj: overlay}})
	o.OverlayStack.Add(overlay)
}

func (o *overlayStack) Remove(overlay granio.CanvasObject) {
	if overlay == nil || len(o.List()) == 0 {
		return
	}

	o.OverlayStack.Remove(overlay)
	overlayCount := len(o.List())

	clear(o.renderCaches[overlayCount:])
	o.renderCaches = o.renderCaches[:overlayCount]
}
