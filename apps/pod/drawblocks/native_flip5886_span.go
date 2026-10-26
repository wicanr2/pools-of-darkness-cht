// Typed Go in-place surface flips and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

type flip5886SpanView struct {
	CPUView
	sp, ds, es uint16
}

func (v flip5886SpanView) SP() uint16 { return v.sp }
func (v flip5886SpanView) DS() uint16 { return v.ds }
func (v flip5886SpanView) ES() uint16 { return v.es }
func planFlip5886Span(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x48a1 {
		return nil
	}
	pushed, scratch, reversed := false, false, false
	switch v.IP() {
	case 0x59c8:
		pushed, scratch, reversed = true, true, true
	case 0x59cf, 0x59d3:
		pushed, scratch = true, true
	case 0x59f1:
		reversed = true
	case 0x59f8, 0x59fc, 0x5a0c, 0x5a10:
	default:
		return nil
	}
	sp := v.SP()
	if pushed {
		if sp > 65533 {
			return nil
		}
		sp += 2
	}
	f, ok := readFlip5886Frame(flip5886SpanView{v, sp, v.CS(), 0xa000})
	if !ok {
		return nil
	}
	bottom := f.top
	if f.vertical {
		bottom = f.bottom
	}
	topAt, bottomAt := f.top*320+f.s.left, bottom*320+f.s.left
	if v.BX() != bottomAt || v.DX() != bottom*64 {
		return nil
	}
	if pushed && v.Read16(f.stack-2) != topAt {
		return nil
	}
	third := v.IP() == 0x5a0c || v.IP() == 0x5a10
	if third {
		if v.DS() != v.CS() || v.ES() != 0xa000 {
			return nil
		}
	} else {
		if v.DS() != 0xa000 {
			return nil
		}
		wantES := uint16(0xa000)
		if scratch {
			wantES = v.CS()
		}
		if v.ES() != wantES {
			return nil
		}
	}
	p := imageBasePlan(v)
	if reversed {
		if !f.horizontal || v.CX() == 0 || v.CX() > f.width {
			return nil
		}
		done := f.width - v.CX()
		src, dst := bottomAt, uint16(topAt)
		if scratch {
			src, dst = topAt, 0x1252
		}
		if v.SI() != src+f.width-2-done || v.DI() != dst+done {
			return nil
		}
		value := v.Read8(0xa0000 + uint32(src+f.width-1-done))
		if byte(v.AX()) != value {
			return nil
		}
		target := uint32(0xa0000) + uint32(dst+done)
		if scratch {
			target = 0x48a10 + uint32(dst+done)
		}
		p.Writes = []machine.HostWrite{{Address: target, Value: value, Step: 1, CS: 0x48a1, IP: v.IP()}}
		p.Instructions, p.Cycles, p.LastIP = 1, 7, v.IP()
		if v.DOSBoxCost() {
			p.Cycles = 2
		}
		p.After.R[cpu.DI]++
		p.After.IP++
		return p
	}
	if !third && f.horizontal {
		return nil
	}
	srcBase, dstBase := uint32(0xa0000)+uint32(bottomAt), uint32(0xa0000)+uint32(topAt)
	srcOff, dstOff := bottomAt, topAt
	if scratch {
		srcBase, dstBase = 0xa0000+uint32(topAt), 0x48a10+0x1252
		srcOff, dstOff = topAt, 0x1252
	}
	if third {
		srcBase, dstBase = 0x48a10+0x1252, 0xa0000+uint32(bottomAt)
		srcOff, dstOff = 0x1252, bottomAt
	}
	count, size, done := f.width/2, uint8(2), uint16(0)
	if v.IP() == 0x59d3 || v.IP() == 0x59fc || v.IP() == 0x5a10 {
		done = count * 2
		count, size = f.width%2, 1
	}
	if v.CX() != count || v.SI() != srcOff+done || v.DI() != dstOff+done {
		return nil
	}
	data := make([]byte, int(count)*int(size))
	copyView(v, srcBase+uint32(done), data)
	p.Instructions, p.Cycles, p.LastIP = 1, 3+5*uint64(count), v.IP()
	if v.DOSBoxCost() {
		p.Cycles = 1 + uint64(count)
	}
	if count > 0 {
		p.Transfers = []machine.HostTransfer{{Address: dstBase + uint32(done), Data: data, Count: count, Width: size, Step: 1, CS: 0x48a1, IP: v.IP()}}
	}
	p.After.R[cpu.CX] = 0
	p.After.R[cpu.SI] += count * uint16(size)
	p.After.R[cpu.DI] += count * uint16(size)
	p.After.IP += 2
	return p
}
