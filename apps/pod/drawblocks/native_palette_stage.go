// Typed palette job boundaries and caller ABI, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var paletteStageCountCode = smallFingerprint(14, "320e832f6cf212d800d90c83a975615f45e0e5059bfe13b5126d584cb1023e77")
var paletteStageOutputCode = smallFingerprint(13, "b3465793806ae3107480c4e82cdb2273d0930e98bf06b4f2e0be250914f2c81f")

func planPaletteStage(v CPUView, budget uint64) *machine.HostWorkPlan {
	ip := v.IP()
	p := imageBasePlan(v)
	if ip == 0x4bc6 {
		if v.CS() != 0x48a1 || v.Read8(0x449) != 0x13 {
			return nil
		}
		for i, b := range []byte{0x5f, 0x5e, 0x1f, 0x5d, 0xcb} {
			if v.Read8(0x48a10+0x4bc6+uint32(i)) != b {
				return nil
			}
		}
		stack, ok := paletteSourceStack(v, 0, 12)
		if !ok {
			return nil
		}
		p.After.R[cpu.DI] = v.Read16(stack)
		p.After.R[cpu.SI] = v.Read16(stack + 2)
		p.After.Seg[cpu.DS] = v.Read16(stack + 4)
		p.After.R[cpu.BP] = v.Read16(stack + 6)
		p.After.IP = v.Read16(stack + 8)
		p.After.Seg[cpu.CS] = v.Read16(stack + 10)
		p.After.R[cpu.SP] += 12
		return imageTimingReturn(v, p, budget, 5, 40, 0x4bca)
	}
	if ip != 0x4b0b && ip != 0x4b8d && ip != 0x4b91 {
		return nil
	}
	lo, fp := uint16(0x4b0b), paletteStageCountCode
	if ip != 0x4b0b {
		lo, fp = 0x4b8d, paletteStageOutputCode
	}
	if !paletteSourceAdmit(v, lo, fp) {
		return nil
	}
	count := v.Read16(0x48a10 + 0x116f)
	if count < 1 || count > 256 {
		return nil
	}
	if ip == 0x4b0b {
		p.After.R[cpu.BX] &= 0xff00
		p.After.R[cpu.CX] = count * 3
		p.After.R[cpu.SI] = 0
		p.After.Flags = nativeCompareFlags(v.Flags(), 0, 0)
		return imageTiming(v, p, budget, 5, 18, 0, 0x4b17, 0x4b19)
	}
	if ip == 0x4b91 {
		p.After.R[cpu.SI] = 0x1872
		p.After.R[cpu.DI] = count * 2
		p.After.Flags = nativeCompareFlags(v.Flags(), count*2, 0)
		return imageTiming(v, p, budget, 3, 10, 0, 0x4b98, 0x4b9a)
	}
	stack, ok := paletteSourceStack(v, 4, 4)
	if !ok {
		return nil
	}
	paletteSourceWord(p, stack-2, 0x48a1, 1, 0x4b8d)
	paletteSourceWord(p, stack-4, 0x4b91, 2, 0x4b8e)
	p.After.R[cpu.SP] -= 4
	return imageTiming(v, p, budget, 2, 15, 0, 0x4b8e, 0x2a90)
}
