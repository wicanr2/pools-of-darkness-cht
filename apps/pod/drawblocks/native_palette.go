package drawblocks

import (
	"crypto/sha256"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var paletteScanCode = smallFingerprint(11, "3c8934166cf6cd8d1ecdf79eac51b0cb33bb1bb3e545b883578dcb85971e1b5a")

// planPaletteScan marks the colors present in one linear framebuffer row.
// It traverses pixel data directly; palette I/O and fading remain outside it.
func planPaletteScan(v CPUView, budget uint64) *machine.HostWorkPlan {
	entry, next, store, last := uint16(0x4a67), uint16(0x4a72), uint16(0x4a6a), uint16(0x4a70)
	if v.IP() == 0x47ba {
		entry, next, store, last = 0x47ba, 0x47c5, 0x47bd, 0x47c3
	}

	if v.CS() != 0x48a1 || v.IP() != entry || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 || v.BX()&0xff00 != 0 || v.CX() == 0 || v.CX() > 320 || budget < 4 {
		return nil
	}
	var code [11]byte
	copyView(v, 0x48a10+uint32(entry), code[:])
	if sha256.Sum256(code[:]) != paletteScanCode.sum {
		return nil
	}
	base, si, count := uint32(v.DS())*16, uint32(v.SI()), uint32(v.CX())
	if base != 0xa0000 || si+count > 64000 || si%320+count > 320 {
		return nil
	}
	if uint64(count) > budget/4 {
		count = uint32(budget / 4)
	}
	p := &machine.HostWorkPlan{Algorithmic: true, After: machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: entry, Flags: v.Flags()}, LastCS: 0x48a1, LastIP: last}
	var color byte
	for x := uint32(0); x < count; x++ {
		color = v.Read8(base + si + x)
		p.Writes = append(p.Writes, machine.HostWrite{Address: 0x48a10 + 0x1772 + uint32(color), Value: 1, Step: uint64(x)*4 + 3, CS: 0x48a1, IP: store})
	}
	p.Instructions = uint64(count) * 4
	p.Cycles = uint64(count) * 18
	if v.DOSBoxCost() {
		p.Cycles = uint64(count) * 5
	}
	p.After.R[cpu.AX] = v.AX()&0xff00 | uint16(color)
	p.After.R[cpu.BX] = uint16(color)
	p.After.R[cpu.SI] += uint16(count)
	p.After.R[cpu.CX] -= uint16(count)
	if p.After.R[cpu.CX] == 0 {
		p.After.IP = next
	}
	return p
}
