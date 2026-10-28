// Go像素轉場，規格025/#49；VGA.DRV 0x4337，保留已驗收語意。
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/machine"
)

func dissolve4337FixedPixels(writes []machine.HostWrite, transfers []machine.HostTransfer) bool {
	for _, w := range writes {
		if w.CS == 0x48a1 && (w.IP == 0x4569 || w.IP == 0x45bb) {
			return true
		}
	}
	for _, s := range transfers {
		if s.Count > 0 && s.CS == 0x48a1 && (s.IP == 0x4569 || s.IP == 0x45bb) {
			return true
		}
	}
	return false
}
func dissolve4337(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() == 0x4337 {
		p := planDissolve4337Prefix(v)
		if p != nil && p.Instructions <= budget {
			return p
		}
	}
	if p := planDissolve4337Batch(v, budget); p != nil {
		return p
	}
	if p := planDissolve4337Cursor(v, budget); p != nil {
		return p
	}
	if p := planDissolve4337Control(v, budget); p != nil {
		return p
	}
	return planDissolve4337Pixel(v, budget)
}
