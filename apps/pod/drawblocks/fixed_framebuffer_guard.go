package drawblocks

import "github.com/wicanr2/dosgolem/internal/machine"

// Fixed compatibility work must never be mistaken for a Go raster algorithm.
// Per-kernel guards additionally protect known offscreen raster destinations.
func fixedFramebufferWrites(writes []machine.HostWrite, transfers []machine.HostTransfer) bool {
	const lo, hi = int64(0xa0000), int64(0xa0000 + 64000)
	for _, w := range writes {
		if a := int64(w.Address); lo <= a && a < hi {
			return true
		}
	}
	for _, t := range transfers {
		if t.Count == 0 || t.Width == 0 {
			continue
		}
		first := int64(t.Address)
		end := first + int64(t.Width)*int64(t.Count)
		if t.Backward {
			first -= int64(t.Width) * (int64(t.Count) - 1)
			end = int64(t.Address) + int64(t.Width)
		}
		if first < hi && lo < end {
			return true
		}
	}
	return false
}
