// Go像素轉場，規格025/#49；VGA.DRV 0x4337，保留已驗收語意。
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

// A bounded LFSR raster job. Source/shape guards run once; every virtual
// iteration still retains ordered state notices and its model cost.
func planDissolve4337Batch(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() != 0x456c {
		return nil
	}
	f, ok := readDissolve4337Frame(v)
	if !ok {
		return nil
	}
	window := uint64(32 + 3*f.hbits)
	if budget < window {
		return nil
	}
	p := imageBasePlan(v)
	state, remaining, counter := f.state, f.remaining, v.Read16(0x48a10+0x1177)
	shadow := make(map[uint32]byte)
	var checked [200]bool
	direction := uint16(1)
	if v.Flags()&cpu.DF != 0 {
		direction = 0xffff
	}
	charge := func(n, c, it uint64) {
		p.Instructions += n
		if v.DOSBoxCost() {
			p.Cycles += n + it
		} else {
			p.Cycles += c
		}
	}
	for p.Instructions+window <= budget {
		low := state & 1
		state >>= 1
		if low != 0 {
			state ^= f.poly
			charge(7, 26, 0)
		} else {
			charge(5, 19, 0)
		}
		imageWord(p, 0x1167, uint16(state), p.Instructions+1, 0x4581)
		imageWord(p, 0x1169, uint16(state>>16), p.Instructions+2, 0x4584)
		charge(3, 14, 0)
		charge(3*uint64(f.hbits), 6*uint64(f.hbits), 0)
		shifted := state >> f.hbits
		x, y := uint16(shifted), uint16(state&f.mask)
		p.After.R[cpu.AX] = x
		p.After.R[cpu.BX] = uint16(shifted >> 16)
		p.After.R[cpu.CX] = 0
		p.After.IP = 0x456c
		if x >= f.width {
			charge(2, 13, 0)
			p.After.Flags = dissolve4337CompareFlags(v.Flags(), x, f.width)
			p.LastIP = 0x4596
			continue
		}
		charge(2, 8, 0)
		p.After.R[cpu.SI] = y
		if y >= f.height {
			charge(4, 25, 0)
			p.After.Flags = dissolve4337CompareFlags(v.Flags(), y, f.height)
			p.LastIP = 0x45a4
			continue
		}
		charge(4, 20, 0)
		src, dest := (f.s.top+y)*320+f.s.left, (f.d.top+y)*320+f.d.left
		if !checked[y] {
			if v.Read16(0x48a10+0x1252+uint32(y)*2) != src || v.Read16(0x48a10+0x13e2+uint32(y)*2) != dest {
				return nil
			}
			checked[y] = true
		}
		src += x
		dest += x
		sourceAt, destAt := uint32(0xa0000)+uint32(src), uint32(0xa0000)+uint32(dest)
		value, found := shadow[sourceAt]
		if !found {
			value = v.Read8(sourceAt)
		}
		p.Transfers = append(p.Transfers, machine.HostTransfer{Address: destAt, Data: []byte{value}, Count: 1, Width: 1, Backward: v.Flags()&cpu.DF != 0, Step: p.Instructions + 8, CS: 0x48a1, IP: 0x45bb})
		shadow[destAt] = value
		charge(9, 72, 1)
		imageWord(p, 0x1171, remaining-1, p.Instructions+1, 0x45be)
		p.After.R[cpu.AX] = v.CS()
		p.After.R[cpu.SI] = src + direction
		p.After.R[cpu.DI] = dest + direction
		p.After.Flags = dissolve4337CompareFlags(v.Flags(), remaining, 1) &^ cpu.CF
		p.After.IP = 0x45c4
		p.LastIP = 0x45c2
		if remaining == 1 {
			charge(7, 53, 0)
			p.After.R[cpu.DI] = v.Read16(f.stack)
			p.After.R[cpu.SI] = v.Read16(f.stack + 2)
			p.After.Seg[cpu.DS] = v.Read16(f.stack + 4)
			p.After.R[cpu.BP] = v.Read16(f.stack + 6)
			p.After.R[cpu.SP] += 12
			p.After.IP = v.Read16(f.stack + 8)
			p.After.Seg[cpu.CS] = v.Read16(f.stack + 10)
			p.LastIP = 0x45df
			return p
		}
		charge(2, 8, 0)
		remaining--
		imageWord(p, 0x1177, counter-1, p.Instructions+1, 0x45c4)
		p.After.Flags = dissolve4337CompareFlags(v.Flags(), counter, 1) &^ cpu.CF
		counter--
		p.LastIP = 0x45c8
		p.After.IP = 0x456c
		if counter == 0 {
			charge(2, 8, 0)
			p.After.IP = 0x45ca
			return p
		}
		charge(2, 13, 0)
	}
	return p
}
