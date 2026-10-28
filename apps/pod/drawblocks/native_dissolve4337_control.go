// Go像素轉場，規格025/#49；VGA.DRV 0x4337，保留已驗收語意。
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

type dissolve4337ControlView struct {
	CPUView
	ds, es uint16
}

func (v dissolve4337ControlView) DS() uint16 { return v.ds }
func (v dissolve4337ControlView) ES() uint16 { return v.es }
func planDissolve4337Pixel(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x48a1 || (v.IP() != 0x4569 && v.IP() != 0x45bb) || v.DS() != 0xa000 || v.ES() != 0xa000 || v.AX() != v.CS() {
		return nil
	}
	f, ok := readDissolve4337Frame(dissolve4337ControlView{v, v.CS(), 0xa000})
	if !ok {
		return nil
	}
	x, y := uint16(0), uint16(0)
	if v.IP() == 0x45bb {
		x, y = uint16(f.state>>f.hbits), uint16(f.state&f.mask)
		if x >= f.width || y >= f.height || v.CX() != 0 || v.BX() != uint16(f.state>>(f.hbits+16)) {
			return nil
		}
	} else {
		if f.state != 1 || f.remaining != f.width*f.height || v.CX() != 0 || v.BX() != f.hbits+uint16(dissolve4337BitsLen16(f.width)) {
			return nil
		}
	}
	src, dest := (f.s.top+y)*320+f.s.left, (f.d.top+y)*320+f.d.left
	if v.Read16(0x48a10+0x1252+uint32(y)*2) != src || v.Read16(0x48a10+0x13e2+uint32(y)*2) != dest || v.SI() != src+x || v.DI() != dest+x {
		return nil
	}
	p := imageBasePlan(v)
	p.Instructions, p.Cycles, p.LastIP = 1, 7, v.IP()
	if v.DOSBoxCost() {
		p.Cycles = 2
	}
	p.Transfers = []machine.HostTransfer{{Address: 0xa0000 + uint32(dest+x), Data: []byte{v.Read8(0xa0000 + uint32(src+x))}, Count: 1, Width: 1, Backward: v.Flags()&cpu.DF != 0, Step: 1, CS: 0x48a1, IP: v.IP()}}
	direction := uint16(1)
	if v.Flags()&cpu.DF != 0 {
		direction = 0xffff
	}
	p.After.R[cpu.SI] += direction
	p.After.R[cpu.DI] += direction
	p.After.IP++
	return p
}
func dissolve4337BitsLen16(v uint16) int {
	n := 0
	for v != 0 {
		n++
		v >>= 1
	}
	return n
}
func planDissolve4337Control(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget < 2 || v.CS() != 0x48a1 || (v.IP() != 0x45c4 && v.IP() != 0x45ce) {
		return nil
	}
	f, ok := readDissolve4337Frame(dissolve4337ControlView{v, v.DS(), 0xa000})
	if !ok {
		return nil
	}
	p := imageBasePlan(v)
	if v.IP() == 0x45c4 {
		counter := v.Read16(0x48a10 + 0x1177)
		imageWord(p, 0x1177, counter-1, 1, 0x45c4)
		p.Instructions, p.Cycles, p.LastIP = 2, 13, 0x45c8
		p.After.Flags = dissolve4337CompareFlags(v.Flags(), counter, 1)&^cpu.CF | v.Flags()&cpu.CF
		p.After.IP = 0x456c
		if counter == 1 {
			p.Cycles = 8
			p.After.IP = 0x45ca
		}
		if v.DOSBoxCost() {
			p.Cycles = 2
		}
		return p
	}
	if budget < 4 || f.period == 0 {
		return nil
	}
	p.Instructions, p.Cycles, p.LastIP = 4, 36, 0x45d9
	if v.DOSBoxCost() {
		p.Cycles = 4
	}
	imageWord(p, 0x1177, f.period, 3, 0x45d6)
	p.After.R[cpu.AX] = f.period
	p.After.Seg[cpu.ES] = 0xa000
	p.After.IP = 0x456c
	return p
}
