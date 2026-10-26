package drawblocks

import "github.com/wicanr2/dosgolem/internal/machine"

// Dispatch a typed line by its coordinate distances, not guest instructions.
func planLine(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || v.Read8(0x449) != 0x13 {
		return nil
	}
	switch v.IP() {
	case 0x309c, 0x30b9, 0x3206, 0x321c:
		return planLineCursor(v, budget)
	}
	if v.IP() != 0x2cfc {
		return nil
	}
	stack, ok := paletteSourceStack(v, 10, 16)
	if !ok {
		return nil
	}
	x0, y0 := v.Read16(stack+6), v.Read16(stack+8)
	x1, y1 := v.Read16(stack+10), v.Read16(stack+12)
	if x0 == x1 || y0 == y1 {
		return planAxisLine(v, budget)
	}
	dx := int32(int16(x1)) - int32(int16(x0))
	dy := int32(int16(y1)) - int32(int16(y0))
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}

	var plans *linePlans
	if dy > dx {
		plans = buildSteepLine(v)
	} else {
		plans = buildShallowLine(v)
	}
	if plans != nil {
		return chooseLinePlan(plans, budget)
	}
	if p := planClippedLinePrefix(v); p != nil && p.Instructions <= budget {
		return p
	}
	return nil
}
