package painter

import (
	"time"

	"github.com/Uzer007-admin/granio/v3/canvas"
)

const MaxSoftwareBlurPixels = 384000
const maxSoftwareBlurDimension = 8192

// SoftwareBlurBudget controls one software blur per painted frame.
type SoftwareBlurBudget struct {
	owner   *canvas.Material
	quality uint8
	slow    uint8
	used    bool
	last    time.Time
}

func (b *SoftwareBlurBudget) ResetFrame() { b.used = false }

func (b *SoftwareBlurBudget) Begin(owner *canvas.Material, width, height int, initial uint8, now time.Time) uint8 {
	if b.used || !owner.SoftwareBlur || width <= 0 || height <= 0 || width > maxSoftwareBlurDimension || height > maxSoftwareBlurDimension || width > MaxSoftwareBlurPixels/height {
		return 0
	}
	b.used = true
	if b.owner != owner || now.Sub(b.last) > time.Second {
		b.owner, b.quality, b.slow = owner, initial, 0
	}
	b.last = now
	return b.quality
}

func (b *SoftwareBlurBudget) End(elapsed time.Duration) {
	if elapsed <= 2*time.Millisecond {
		b.slow = 0
		return
	}
	b.slow++
	if b.slow == 3 {
		b.slow = 0
		if b.quality == 4 {
			b.quality = 8
		} else {
			b.quality = 0
		}
	}
}
