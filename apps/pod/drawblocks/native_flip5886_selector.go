// Typed Go in-place surface flips and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/machine"
)

func flip5886(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() == 0x5886 {
		p := planFlip5886Prefix(v)
		if p != nil && p.Instructions <= budget {
			return p
		}
	}
	if p := planFlip5886Address(v, budget); p != nil {
		return p
	}
	if p := planFlip5886Rows(v, budget); p != nil {
		return p
	}
	return planFlip5886Span(v, budget)
}

func flip5886FixedRaster(writes []machine.HostWrite, transfers []machine.HostTransfer) bool {
	target := func(cs, ip uint16) bool {
		return cs == 0x48a1 && (ip == 0x59c8 || ip == 0x59cf || ip == 0x59d3 || ip == 0x59f1 || ip == 0x59f8 || ip == 0x59fc || ip == 0x5a0c || ip == 0x5a10)
	}
	for _, w := range writes {
		if target(w.CS, w.IP) {
			return true
		}
	}
	for _, s := range transfers {
		if s.Count > 0 && target(s.CS, s.IP) {
			return true
		}
	}
	return false
}
