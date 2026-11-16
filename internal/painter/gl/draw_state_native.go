//go:build !wasm

package gl

func sameProgram(a, b Program) bool { return a == b }

func sameBuffer(a, b Buffer) bool { return a == b }
