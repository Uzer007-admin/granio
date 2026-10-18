// Package cache configures the framework's bounded CPU caches.
//
// Renderers, live canvas images and GPU textures have separate lifetimes and are
// not counted here. Applications should release images they no longer display.
package cache

import internal "github.com/Uzer007-admin/granio/v2/internal/cache"

// Limits bounds text metrics and rasterized SVGs by entry count and retained
// bytes. A zero limit disables that cache. Byte accounting excludes Go allocator
// and map overhead, which is additionally bounded by the entry limits.
type Limits = internal.Limits

// Usage reports the retained data in the caches governed by Limits.
type Usage = internal.Usage

// GetLimits returns the active limits. Defaults are 8192 text entries / 8 MiB
// and 256 SVG entries / 64 MiB.
func GetLimits() Limits { return internal.GetLimits() }

// SetLimits changes limits and immediately evicts least recently used entries.
// Negative limits return an error without changing the current limits.
func SetLimits(limits Limits) error { return internal.SetLimits(limits) }

// GetUsage reports the current cache sizes.
func GetUsage() Usage { return internal.GetUsage() }

// Clear discards text metrics and rasterized SVGs. Live images remain usable.
func Clear() { internal.ResetThemeCaches() }
