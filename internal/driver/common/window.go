package common

import (
	"github.com/Uzer007-admin/granio/v3/internal/async"
)

var DonePool = async.Pool[chan struct{}]{
	New: func() chan struct{} {
		return make(chan struct{})
	},
}
