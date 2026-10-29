package pod

import (
	"github.com/wicanr2/dosgolem/apps/pod/drawblocks"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func DrawBlockRanges() []machine.HostCodeRange {
	return []machine.HostCodeRange{{CS: 0x1b2e, Lo: 0x1c4ca, Hi: 0x1c4cb}, {CS: 0x1b2e, Lo: 0x1c4d1, Hi: 0x1c4d2}, {CS: 0x48a1, Lo: 0x48a10 + 0x2772, Hi: 0x48a10 + 0x757f}, {CS: 0x1ed4, Lo: 0x1ed40 + 0x1a5c, Hi: 0x1ed40 + 0x1a93}, {CS: 0x1d1f, Lo: 0x1d1f0 + 0x46e, Hi: 0x1d1f0 + 0x487}}
}
func (l *Live) PrepareHostBlock(v HostCallCPU, budget uint64) (*machine.HostWorkPlan, error) {
	if l.Inventory {
		return nil, nil
	}
	// GAME sampling remains a boundary; VACE starts before GAME's ready marker.
	if l.ready && l.Map != nil {
		if l.nextMapSample <= v.Steps() {
			return nil, nil
		}
		if left := l.nextMapSample - v.Steps(); left < budget {
			budget = left
		}
	}
	return drawblocks.Plan(v, budget), nil
}
