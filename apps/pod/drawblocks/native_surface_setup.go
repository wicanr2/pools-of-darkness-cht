package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var surfaceRootSetupCode = smallFingerprint(210, "253f27203f42adc840d848642ff4a1214f0353b26c033ae79a219dcdb4520c14")
var surfaceChildSetupCode = smallFingerprint(124, "7af20b90b4bf497b5a4011bedcdc823fd594851237ea59e0faa6fd6b1c0db228")
var surfaceFreeSlotCode = smallFingerprint(20, "d2e4e2cfc6bf845297c8f699fc5ba9d738f12b64722f250b8f333a0b06512d8c")

func surfaceSetupCode(v CPUView, ip uint16, fp smallCode) bool {
	var raw [210]byte
	copyView(v, 0x48a10+uint32(ip), raw[:fp.size])
	return sha256.Sum256(raw[:fp.size]) == fp.sum
}
func surfaceSetupStack(v CPUView, before, after uint32) (uint32, bool) {
	sp := uint32(v.SP())
	at := uint32(v.SS())*16 + sp
	return at, sp >= before && sp+after <= 65536 && at >= 0x500+before && at+after <= 0xa0000 && !(at-before < 0x50000 && at+after > 0x48a10)
}
func planSurfaceSetup(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x48a1 || v.Read8(0x449) != 0x13 || (v.IP() != 0x2786 && v.IP() != 0x28d5) {
		return nil
	}
	model, ok := v.(interface{ CPUModel() (cpu.Model, bool) })
	if !ok {
		return nil
	}
	m, known := model.CPUModel()
	if !known || m != cpu.Model80386 {
		return nil
	}
	if v.IP() == 0x2786 {
		return planSurfaceRootSetup(v, budget)
	}
	return planSurfaceChildSetup(v, budget)
}

// Derive both root descriptors and unused-slot marks directly from the call
// arguments. ABI writes preserve the guest's ordered observation boundaries.
func planSurfaceRootSetup(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget < 837 || !surfaceSetupCode(v, 0x2786, surfaceRootSetupCode) {
		return nil
	}
	stack, ok := surfaceSetupStack(v, 12, 12)
	if !ok {
		return nil
	}
	pointer := uint32(v.Read16(stack+6))*16 + uint32(v.Read16(stack+4))
	if pointer < 0x500 || pointer >= 0xa0000 {
		return nil
	}
	length := uint32(v.Read16(stack+8)) | uint32(v.Read16(stack+10))<<16
	first := uint16(pointer>>4) + 1
	limit := uint16(pointer>>4) + uint16(length>>4)
	next := first + 0xfa0
	p := imageBasePlan(v)
	emit := func(a uint32, value uint16, step uint64, ip uint16) { paletteSourceWord(p, a, value, step, ip) }
	for _, w := range []struct {
		depth, value uint16
		step         uint64
		ip           uint16
	}{{2, v.BP(), 1, 0x2786}, {4, v.DS(), 3, 0x2789}, {6, v.SI(), 4, 0x278a}, {8, v.DI(), 5, 0x278b}, {10, v.AX(), 18, 0x27a6}, {12, v.BX(), 19, 0x27a7}, {10, 0x48a1, 48, 0x27df}} {
		emit(stack-uint32(w.depth), w.value, w.step, w.ip)
	}
	imageWord(p, 0xf05, limit, 50, 0x27e1)
	for _, w := range []struct {
		off, value uint16
		step       uint64
		ip         uint16
	}{{0xbd, 0xa000, 51, 0x27e5}, {0x2bd, 0xfa0, 52, 0x27eb}, {0x4bd, 0, 53, 0x27f1}, {0x6bd, 0, 54, 0x27f7}, {0x8bd, 319, 55, 0x27fd}, {0xabd, 199, 56, 0x2803}, {0xcbd, 0x20, 57, 0x2809}, {0xbf, first, 58, 0x280f}, {0x2bf, 0xfa0, 59, 0x2812}, {0x4bf, 0, 60, 0x2818}, {0x6bf, 0, 61, 0x281e}, {0x8bf, 319, 62, 0x2824}, {0xabf, 199, 63, 0x282a}, {0xcbf, 0, 64, 0x2830}, {0xf07, next, 66, 0x2839}} {
		imageWord(p, w.off, w.value, w.step, w.ip)
	}
	for slot := uint16(2); slot < 256; slot++ {
		p.Writes = append(p.Writes, machine.HostWrite{Address: 0x48a10 + 0xcbd + uint32(slot)*2, Value: 0x80, Step: 70 + uint64(slot-2)*3, CS: 0x48a1, IP: 0x2844})
	}
	imageWord(p, 0x1109, 0, 832, 0x284d)
	p.After.R[cpu.AX] = next&0xff00 | 0x80
	p.After.R[cpu.BX] = 512
	p.After.R[cpu.CX] = 0
	p.After.R[cpu.DX] = uint16(length >> 20)
	p.After.R[cpu.SP] += 4
	p.After.IP = v.Read16(stack)
	p.After.Seg[cpu.CS] = v.Read16(stack + 2)
	p.After.Flags = imageRowAddFlags(v.Flags(), 510, 2)
	return imageTimingReturn(v, p, budget, 837, 2848, 0x2857)
}

type childSurfaceSetup struct {
	slot, root, free, left, top, right, bottom uint16
	links                                      uint64
}

func readChildSurfaceSetup(v CPUView, stack uint32) (childSurfaceSetup, bool) {
	j := childSurfaceSetup{slot: v.Read16(stack + 4), left: v.Read16(stack + 6), top: v.Read16(stack + 8), right: v.Read16(stack + 10), bottom: v.Read16(stack + 12)}
	if j.slot > 255 || j.left > j.right || j.right >= 320 || j.top > j.bottom || j.bottom >= 200 {
		return j, false
	}
	word := func(off, slot uint16) uint16 { return v.Read16(0x48a10 + uint32(off) + uint32(slot)*2) }
	found := false
	for slot := uint16(0); slot < 256; slot++ {
		if word(0xcbd, slot)&0x80 != 0 {
			j.free = slot
			found = true
			break
		}
	}
	if !found || j.free < 2 {
		return j, false
	}
	parent := j.slot
	for {
		if parent == j.free || word(0xcbd, parent)&0x80 != 0 {
			return j, false
		}
		if word(0xcbd, parent)&0x40 == 0 {
			j.root = parent
			break
		}
		raw := word(0xbd, parent)
		if j.links >= 8 || raw > 510 || raw&1 != 0 {
			return j, false
		}
		parent = raw / 2
		j.links++
	}
	if word(0xbd, j.root) != 0xa000 || word(0x4bd, j.root) > word(0x8bd, j.root) || word(0x8bd, j.root) >= 320 || word(0x6bd, j.root) > word(0xabd, j.root) || word(0xabd, j.root) >= 200 {
		return j, false
	}
	if j.left < word(0x4bd, j.slot) || j.right > word(0x8bd, j.slot) || j.top < word(0x6bd, j.slot) || j.bottom > word(0xabd, j.slot) {
		return j, false
	}
	return j, true
}

func planSurfaceChildSetup(v CPUView, budget uint64) *machine.HostWorkPlan {
	if !surfaceSetupCode(v, 0x28d5, surfaceChildSetupCode) || !surfaceSetupCode(v, 0x2772, surfaceFreeSlotCode) {
		return nil
	}
	stack, ok := surfaceSetupStack(v, 12, 14)
	if !ok {
		return nil
	}
	j, ok := readChildSurfaceSetup(v, stack)
	if !ok {
		return nil
	}
	n := 48 + uint64(j.free)*4 + j.links*4
	if n > budget {
		return nil
	}
	p := imageBasePlan(v)
	for _, w := range []struct {
		depth, value uint16
		step         uint64
		ip           uint16
	}{{2, v.BP(), 1, 0x28d5}, {4, v.DS(), 3, 0x28d8}, {6, v.SI(), 4, 0x28d9}, {8, v.DI(), 5, 0x28da}, {10, 0x48a1, 7, 0x28dc}, {10, 0x48a1, 9, 0x28de}, {12, 0x28e2, 10, 0x28df}} {
		paletteSourceWord(p, stack-uint32(w.depth), w.value, w.step, w.ip)
	}
	shift := uint64(j.free)*4 + j.links*4
	for _, w := range []struct {
		off, value uint16
		step       uint64
		ip         uint16
	}{{0xbd, j.root * 2, 21, 0x28f7}, {0x2bd, 0, 24, 0x2900}, {0x4bd, j.left, 28, 0x290f}, {0x6bd, j.top, 32, 0x291c}, {0x8bd, j.right, 36, 0x2929}, {0xabd, j.bottom, 40, 0x2936}, {0xcbd, 0x40, 41, 0x293a}} {
		imageWord(p, w.off+j.free*2, w.value, w.step+shift, w.ip)
	}
	p.After.R[cpu.AX] = j.free
	p.After.R[cpu.CX] = 256 - j.free
	p.After.R[cpu.SP] += 4
	p.After.IP = v.Read16(stack)
	p.After.Seg[cpu.CS] = v.Read16(stack + 2)
	p.After.Flags = nativeCompareFlags(v.Flags()&^cpu.DF, j.free, 0)
	return imageTimingReturn(v, p, budget, n, 281+12*uint64(j.free-2)+21*j.links, 0x2948)
}
