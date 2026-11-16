//go:build wasm

package gl

import "github.com/Uzer007-admin/granio/deps/gl-js"

func sameProgram(a, b Program) bool { return gl.Program(a).Equal(gl.Program(b).Value) }

func sameBuffer(a, b Buffer) bool { return gl.Buffer(a).Equal(gl.Buffer(b).Value) }
