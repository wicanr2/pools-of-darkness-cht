// Typed source-color list and RGB ABI, spec025/#49.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var paletteSourceControlCode = smallFingerprint(44, "eca20cc973c3b83fb0067e4db9db860c4f59a4b58d5bb9666ff03ba198374630")
var paletteSourceSampleCode = smallFingerprint(38, "96a0c67b2570207483279d9535277fb736d13334d7e9dfafde9a8aeefd3711ec")
var paletteSourceStoreCode = smallFingerprint(26, "4b78e592196fa665f748d637540e457744cd949714e581dc3264295b465806d2")

func paletteSourceAdmit(v CPUView, lo uint16, fp smallCode) bool {
	if v.CS() != 0x48a1 || v.DS() != 0x48a1 || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 {
		return false
	}
	var code [44]byte
	copyView(v, 0x48a10+uint32(lo), code[:fp.size])
	return sha256.Sum256(code[:fp.size]) == fp.sum
}
func paletteSourceStack(v CPUView, before, after uint32) (uint32, bool) {
	sp := uint32(v.SP())
	at := uint32(v.SS())*16 + sp
	return at, sp >= before && sp+after <= 65536 && at >= 0x500+before && at+after <= 0xa0000 && !(at-before < 0x50000 && at+after > 0x48a10)
}
func paletteSourceWord(p *machine.HostWorkPlan, at uint32, value uint16, step uint64, ip uint16) {
	p.Writes = append(p.Writes, machine.HostWrite{Address: at, Value: byte(value), Step: step, CS: 0x48a1, IP: ip}, machine.HostWrite{Address: at + 1, Value: byte(value >> 8), Step: step, CS: 0x48a1, IP: ip})
}

func planPaletteSourceControl(v CPUView, budget uint64) *machine.HostWorkPlan {
	ip := v.IP()
	if ip != 0x4a91 && ip != 0x4aa4 && ip != 0x4aad {
		return nil
	}
	if !paletteSourceAdmit(v, 0x4a91, paletteSourceControlCode) {
		return nil
	}
	p := imageBasePlan(v)
	if ip == 0x4a91 {
		if v.DI() < 0x1572 || v.DI() > 0x1772 || v.DI()&1 != 0 {
			return nil
		}
		count := (v.DI() - 0x1572) / 2
		p.After.R[cpu.DI] = count
		if count == 0 {
			p.After.Flags = nativeCompareFlags(v.Flags(), v.DI(), 0x1572)
			return imageTiming(v, p, budget, 3, 11, 0, 0x4a97, 0x4bc6)
		}
		stack, ok := paletteSourceStack(v, 4, 4)
		if !ok {
			return nil
		}
		p.After.Flags = nativeCompareFlags(v.Flags(), count, 0)
		imageWord(p, 0x116f, count, 4, 0x4a9c)
		paletteSourceWord(p, stack-2, 0x48a1, 5, 0x4aa0)
		paletteSourceWord(p, stack-4, 0x4aa4, 6, 0x4aa1)
		p.After.R[cpu.SP] -= 4
		return imageTiming(v, p, budget, 6, 32, 0, 0x4aa1, 0x2a90)
	}
	count := v.Read16(0x48a10 + 0x116f)
	if count == 0 || count > 256 {
		return nil
	}
	if ip == 0x4aa4 {
		p.After.R[cpu.SI] = count * 2
		p.After.R[cpu.DI] = 0x1872
		p.After.Flags = nativeCompareFlags(v.Flags(), count*2, 0)
		return imageTiming(v, p, budget, 3, 10, 0, 0x4aaa, 0x4aad)
	}
	if v.SI()&1 != 0 || v.SI() > count*2 || uint32(v.DI()) != 0x1872+3*(uint32(count)-uint32(v.SI()/2)) {
		return nil
	}
	p.After.R[cpu.SI] -= 2
	if v.SI() == 0 {
		p.After.Flags = nativeCompareFlags(v.Flags(), 0, 2)
		return imageTiming(v, p, budget, 2, 9, 0, 0x4ab0, 0x4af6)
	}
	mask := p.After.R[cpu.SI] & 0x20
	p.After.R[cpu.AX] = mask
	p.After.Flags = nativeCompareFlags(v.Flags(), mask, 0)
	if mask != 0 {
		return imageTiming(v, p, budget, 5, 15, 0, 0x4ab7, 0x4abd)
	}
	stack, ok := paletteSourceStack(v, 4, 4)
	if !ok {
		return nil
	}
	paletteSourceWord(p, stack-2, 0x48a1, 6, 0x4ab9)
	paletteSourceWord(p, stack-4, 0x4abd, 7, 0x4aba)
	p.After.R[cpu.SP] -= 4
	return imageTiming(v, p, budget, 7, 25, 0, 0x4aba, 0x2a90)
}

func planPaletteSourceSample(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() != 0x4abd || !paletteSourceAdmit(v, 0x4abd, paletteSourceSampleCode) {
		return nil
	}
	count := v.Read16(0x48a10 + 0x116f)
	if count == 0 || count > 256 || v.SI()&1 != 0 || v.SI() >= count*2 || uint32(v.DI()) != 0x1872+3*(uint32(count)-uint32(v.SI()/2)-1) {
		return nil
	}
	index := v.Read16(0x48a10 + 0x1572 + uint32(v.SI()))
	if index > 255 {
		return nil
	}
	stack, ok := paletteSourceStack(v, 12, 4)
	if !ok {
		return nil
	}
	p := imageBasePlan(v)
	for i := uint32(0); i < 3; i++ {
		n, ip := uint64(i)*5, uint16(0x4abd+i*10)
		paletteSourceWord(p, stack-2*(i+1), v.AX(), n+1, ip)
		paletteSourceWord(p, stack-2*(i+2), v.BP(), n+2, ip+1)
		paletteSourceWord(p, stack-2*(i+1), uint16(0x1183-i*2), n+4, ip+4)
	}
	paletteSourceWord(p, stack-8, index, 16, 0x4adb)
	paletteSourceWord(p, stack-10, 0x48a1, 17, 0x4adf)
	paletteSourceWord(p, stack-12, 0x4ae3, 18, 0x4ae0)
	p.After.R[cpu.SP] -= 12
	return imageTiming(v, p, budget, 18, 83, 0, 0x4ae0, 0x2aeb)
}

func planPaletteSourceStore(v CPUView, budget uint64) *machine.HostWorkPlan {
	ip := v.IP()
	if ip != 0x4ae3 && ip != 0x4af6 {
		return nil
	}
	if !paletteSourceAdmit(v, 0x4ae3, paletteSourceStoreCode) {
		return nil
	}
	count := v.Read16(0x48a10 + 0x116f)
	if count == 0 || count > 256 {
		return nil
	}
	p := imageBasePlan(v)
	if ip == 0x4af6 {
		p.After.R[cpu.CX] = count
		p.After.R[cpu.DI] = 0x1b72
		return imageTiming(v, p, budget, 2, 8, 0, 0x4afa, 0x4afd)
	}
	if v.SI()&1 != 0 || v.SI() >= count*2 || uint32(v.DI()) != 0x1872+3*(uint32(count)-uint32(v.SI()/2)-1) {
		return nil
	}
	stack, ok := paletteSourceStack(v, 0, 8)
	if !ok {
		return nil
	}
	rgb := [3]uint16{v.Read16(0x48a10 + 0x117f), v.Read16(0x48a10 + 0x1181), v.Read16(0x48a10 + 0x1183)}
	for _, color := range rgb {
		if color > 63 {
			return nil
		}
	}
	paletteSourceWord(p, stack+6, 0x48a1, 2, 0x4ae6)
	for i, color := range rgb {
		p.Writes = append(p.Writes, machine.HostWrite{Address: 0x48a10 + uint32(v.DI()) + uint32(i), Value: byte(color), Step: uint64(5 + i*2), CS: 0x48a1, IP: uint16(0x4aeb + i*4)})
	}
	p.After.R[cpu.SP] += 8
	p.After.Seg[cpu.ES] = 0x48a1
	p.After.R[cpu.DI] += 3
	p.After.R[cpu.AX] = v.AX()&0xff00 | rgb[2]
	p.After.Flags = imageRowAddFlags(v.Flags(), v.SP(), 8)
	return imageTiming(v, p, budget, 10, 44, 3, 0x4af4, 0x4aad)
}
