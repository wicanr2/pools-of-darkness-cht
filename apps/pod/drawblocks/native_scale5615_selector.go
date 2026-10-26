// Typed Go integer surface scaling and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/machine"
)

func scale5615(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() == 0x5615 {
		p := planScale5615Prefix(v)
		if p != nil && p.Instructions <= budget {
			return p
		}
	}
	if p := planScale5615Prepare(v, budget); p != nil {
		return p
	}
	if p := planScale5615Output(v, budget); p != nil {
		return p
	}
	return planScale5615Span(v, budget)
}

func scale5615FixedRaster(writes []machine.HostWrite, transfers []machine.HostTransfer) bool {
	target := func(cs, ip uint16) bool {
		return cs == 0x48a1 && (ip == 0x5813 || ip == 0x5879 || ip == 0x587d || ip == 0x5856 || ip == 0x585a)
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
