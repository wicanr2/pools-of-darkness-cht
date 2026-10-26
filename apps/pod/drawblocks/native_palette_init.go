// Typed fade source surface and work-array setup, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var paletteInitEntryCode = smallFingerprint(45, "3d100b4573d6a562c44b7b75771459a8c9237ab5cffcfc10b6f597adc0716fe8")
var paletteInitSurfaceCode = smallFingerprint(192, "72223da417d7c23f23a9ce6d16c7f5e8ae325a2de06b3be04648ef937e9bb14e")

func planPaletteInitialize(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || (v.IP() != 0x4944 && v.IP() != 0x4971) || v.Read8(0x449) != 0x13 {
		return nil
	}
	lo, fp := uint16(0x4944), paletteInitEntryCode
	if v.IP() == 0x4971 {
		lo, fp = 0x4971, paletteInitSurfaceCode
		if v.DS() != 0x48a1 {
			return nil
		}
	}
	var code [192]byte
	copyView(v, 0x48a10+uint32(lo), code[:fp.size])
	if sha256.Sum256(code[:fp.size]) != fp.sum {
		return nil
	}
	before, after := uint32(20), uint32(8)
	if v.IP() == 0x4971 {
		before, after = 0, 22
	}
	stack, ok := paletteSourceStack(v, before, after)
	if !ok {
		return nil
	}
	p := imageBasePlan(v)
	if v.IP() == 0x4944 {
		target := v.Read16(stack + 6)
		if target > 255 {
			return nil
		}
		for _, w := range []struct {
			off   uint32
			value uint16
			step  uint64
			ip    uint16
		}{{2, v.BP(), 1, 0x4944}, {4, v.DS(), 3, 0x4947}, {6, v.SI(), 4, 0x4948}, {8, v.DI(), 5, 0x4949}, {10, v.CS(), 6, 0x494a}} {
			paletteSourceWord(p, stack-w.off, w.value, w.step, w.ip)
		}
		for i := uint32(0); i < 3; i++ {
			n, ip := uint64(8+i*5), uint16(0x494c+i*10)
			paletteSourceWord(p, stack-10-2*i, v.AX(), n, ip)
			paletteSourceWord(p, stack-12-2*i, v.SP()-2, n+1, ip+1)
			paletteSourceWord(p, stack-10-2*i, uint16(0x117d-2*i), n+3, ip+4)
		}
		paletteSourceWord(p, stack-16, target, 23, 0x496a)
		paletteSourceWord(p, stack-18, v.CS(), 24, 0x496d)
		paletteSourceWord(p, stack-20, 0x4971, 25, 0x496e)
		p.After.R[cpu.BP] = v.SP() - 2
		p.After.R[cpu.SP] -= 20
		p.After.Seg[cpu.DS] = v.CS()
		return imageTiming(v, p, budget, 25, 109, 0, 0x496e, 0x2aeb)
	}
	if v.BP() != v.SP()+14 {
		return nil
	}
	slot := v.Read16(uint32(v.SS())*16 + uint32(v.BP()) + 6)
	if slot > 255 {
		return nil
	}
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	descriptor := func(off uint16) rectangleSurface {
		return rectangleSurface{word(0xbd + off), word(0x2bd + off), word(0x4bd + off), word(0x6bd + off), word(0x8bd + off), word(0xabd + off), word(0xcbd + off)}
	}
	s := descriptor(slot * 2)
	root := s
	links := uint64(0)
	rootSlot := slot * 2
	for root.flags&64 != 0 {
		if links >= 8 || root.seg > 510 || root.seg&1 != 0 {
			return nil
		}
		rootSlot = root.seg
		root = descriptor(rootSlot)
		links++
	}
	if root.seg != 0xa000 || root.left > root.right || root.right >= 320 || root.top > root.bottom || root.bottom >= 200 || s.left < root.left || s.left > s.right || s.right > root.right || s.top < root.top || s.top > s.bottom || s.bottom > root.bottom {
		return nil
	}
	for _, w := range []struct {
		off, value uint16
		step       uint64
		ip         uint16
	}{{0xebd, s.seg, 5, 0x497d}, {0xebf, s.extra, 7, 0x4984}, {0xec1, s.left, 9, 0x498b}, {0xec3, s.top, 11, 0x4992}, {0xec5, s.right, 13, 0x4999}, {0xecb, s.right - s.left + 1, 16, 0x49a1}, {0xec7, s.bottom, 18, 0x49a8}, {0xecd, s.bottom - s.top + 1, 21, 0x49b0}, {0xec9, s.flags, 23, 0x49b7}} {
		imageWord(p, w.off, w.value, w.step, w.ip)
	}
	for _, w := range []struct {
		off, value uint16
		step       uint64
		ip         uint16
	}{{0xedb, root.flags, 26, 0x49cc}, {0xecf, root.seg, 28, 0x49d3}, {0xed1, root.extra, 30, 0x49da}, {0xed5, root.top, 32, 0x49e1}, {0xed9, root.bottom, 34, 0x49e8}, {0xedf, root.bottom - root.top + 1, 37, 0x49f0}, {0xed3, root.left, 39, 0x49f7}, {0xed7, root.right, 41, 0x49fe}, {0xedd, root.right - root.left + 1, 44, 0x4a06}, {0x1133, s.bottom - s.top + 1, 47, 0x4a0d}, {0x1135, s.right - s.left + 1, 49, 0x4a13}} {
		imageWord(p, w.off, w.value, w.step+4*links, w.ip)
	}
	paletteSourceWord(p, stack+6, 0x48a1, 50+4*links, 0x4a16)
	p.Transfers = []machine.HostTransfer{{Address: 0x48a10 + 0x1772, Count: 128, Width: 2, Step: 55 + 4*links, CS: 0x48a1, IP: 0x4a20}, {Address: 0x48a10 + 0x1572, Count: 256, Width: 2, Step: 58 + 4*links, CS: 0x48a1, IP: 0x4a28}}
	paletteSourceWord(p, stack+6, v.BP(), 59+4*links, 0x4a2a)
	p.After.R[cpu.SP] += 6
	p.After.Seg[cpu.ES] = 0x48a1
	p.After.R[cpu.AX] = 0
	p.After.R[cpu.CX] = 0
	p.After.R[cpu.DI] = 0x1772
	p.After.R[cpu.BX] = rootSlot & 255
	p.After.R[cpu.BP] = s.top
	p.After.Flags = nativeCompareFlags(v.Flags(), 0, 0)
	cost := uint64(2131)
	if links > 0 {
		cost += 21*links - 10
	}
	return imageTiming(v, p, budget, 61+4*links, cost, 384, 0x4a2d, 0x4a31)
}
