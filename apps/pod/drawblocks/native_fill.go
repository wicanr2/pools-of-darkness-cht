package drawblocks

// Spec025/#49: rectangle pixels are produced by a Go span algorithm.
import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"math/bits"
)

var rectangleCode = smallFingerprint(260, "c7db9b7b119f14464a7e76d35125922a3559d10fef36624f79f5120dd0d02f00")

type rectangleSurface struct{ seg, extra, left, top, right, bottom, flags uint16 }
type rectangleFill struct {
	selected, root rectangleSurface
	color, y       uint16
	links          int
}

func nativeCompareFlags(before, a, b uint16) uint16 {
	r := a - b
	f := before &^ uint16(cpu.CF|cpu.PF|cpu.AF|cpu.ZF|cpu.SF|cpu.OF|cpu.DF)
	if a < b {
		f |= cpu.CF
	}
	if (a^b^r)&16 != 0 {
		f |= cpu.AF
	}
	if r == 0 {
		f |= cpu.ZF
	}
	if r&0x8000 != 0 {
		f |= cpu.SF
	}
	if ((a^b)&(a^r))&0x8000 != 0 {
		f |= cpu.OF
	}
	if bits.OnesCount8(byte(r))%2 == 0 {
		f |= cpu.PF
	}
	return f
}

// planRectangle paints horizontal spans. CS:IP selects only an entry/resume
// boundary; neither pixel generation nor row traversal dispatches x86 opcodes.
func planRectangle(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.CS() != 0x48a1 || (v.IP() != 0x41af && v.IP() != 0x4268) || v.Read8(0x449) != 0x13 {
		return nil
	}
	var code [260]byte
	copyView(v, 0x48a10+0x41af, code[:])
	if sha256.Sum256(code[:]) != rectangleCode.sum {
		return nil
	}
	initial := v.IP() == 0x41af
	base := uint32(0x48a10)
	word := func(off uint16) uint16 { return v.Read16(base + uint32(off)) }
	sp := uint32(v.SP())
	stack := uint32(v.SS())*16 + sp
	if initial {
		if sp < 10 || sp+8 > 65536 || stack < 0x50a || stack+8 > 0xa0000 {
			return nil
		}
	} else if sp+12 > 65536 || stack < 0x500 || stack+12 > 0xa0000 || v.BP() != v.SP()+6 || v.DS() != 0x48a1 {
		return nil
	}
	if stack+12 > base && stack-10 < base+0x757f {
		return nil
	}
	descriptor := func(off uint16) rectangleSurface {
		return rectangleSurface{word(0xbd + off), word(0x2bd + off), word(0x4bd + off), word(0x6bd + off), word(0x8bd + off), word(0xabd + off), word(0xcbd + off)}
	}
	r := rectangleFill{}
	if initial {
		slot := v.Read16(stack + 4)
		if slot > 255 {
			return nil
		}
		r.selected = descriptor(slot * 2)
		r.root = r.selected
		for r.root.flags&64 != 0 {
			if r.links >= 8 || r.root.seg > 510 || r.root.seg&1 != 0 {
				return nil
			}
			r.root = descriptor(r.root.seg)
			r.links++
		}
		r.color = v.Read16(stack + 6)
		r.y = r.selected.top
	} else {
		r.selected = rectangleSurface{word(0xebd), word(0xebf), word(0xec1), word(0xec3), word(0xec5), word(0xec7), word(0xec9)}
		r.root = rectangleSurface{word(0xecf), word(0xed1), word(0xed3), word(0xed5), word(0xed7), word(0xed9), word(0xedb)}
		r.color = word(0x1129)
		r.y = v.BX()
		if v.ES() != r.root.seg || word(0x113b) != r.selected.left || word(0x113d) != r.selected.right {
			return nil
		}
	}
	s, root := r.selected, r.root
	if root.seg != 0xa000 || root.flags&64 != 0 || root.left > root.right || root.top > root.bottom || root.right >= 320 || root.bottom >= 200 || s.left < root.left || s.right > root.right || s.left > s.right || s.top < root.top || s.bottom > root.bottom || s.top > s.bottom || r.y < s.top || r.y > s.bottom {
		return nil
	}
	width, height := int(s.right-s.left+1), int(s.bottom-r.y+1)
	stride := int(root.right - root.left + 1)
	if !initial && (word(0xedd) != uint16(stride) || word(0xedf) != root.bottom-root.top+1 || word(0xecb) != uint16(width) || word(0xecd) != s.bottom-s.top+1) {
		return nil
	}
	rowSteps, rowCycles, wordStep := uint64(22), uint64(97+5*(width/2+width%2)), uint64(17)
	if stride == 320 {
		rowSteps = 24
		rowCycles -= 10
		wordStep += 2
	}
	prefixN, prefixCycles := uint64(0), uint64(0)
	if initial {
		prefixN = 59 + 4*uint64(r.links)
		prefixCycles = 238
		if r.links > 0 {
			prefixCycles += 21*uint64(r.links) - 10
		}
	}
	if budget < prefixN+rowSteps {
		return nil
	}
	rows := int((budget - prefixN) / rowSteps)
	if rows > height {
		rows = height
	}
	complete := rows == height
	if complete && prefixN+uint64(rows)*rowSteps+5 > budget {
		rows--
		complete = false
	}
	if rows == 0 {
		return nil
	}
	p := &machine.HostWorkPlan{After: machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: v.IP(), Flags: v.Flags()}, Instructions: prefixN + uint64(rows)*rowSteps, Cycles: prefixCycles + uint64(rows)*rowCycles, LastCS: 0x48a1, LastIP: 0x42ac}
	p.Algorithmic = true
	emit := func(addr uint32, value uint16, step uint64, ip uint16) {
		p.Writes = append(p.Writes, machine.HostWrite{Address: addr, Value: byte(value), Step: step, CS: 0x48a1, IP: ip}, machine.HostWrite{Address: addr + 1, Value: byte(value >> 8), Step: step, CS: 0x48a1, IP: ip})
	}
	if initial {
		// ABI save area and externally visible descriptor scratch writes.
		for _, w := range []struct {
			off   uint32
			value uint16
			step  uint64
			ip    uint16
		}{{2, v.BP(), 1, 0x41af}, {4, v.DS(), 3, 0x41b2}, {6, v.SI(), 4, 0x41b3}, {8, v.DI(), 5, 0x41b4}, {10, v.CS(), 6, 0x41b5}} {
			emit(stack-w.off, w.value, w.step, w.ip)
		}
		for _, w := range []struct {
			off, value uint16
			step       uint64
			ip         uint16
		}{
			{0xebd, s.seg, 11, 0x41c0}, {0xebf, s.extra, 13, 0x41c7}, {0xec1, s.left, 15, 0x41ce}, {0xec3, s.top, 17, 0x41d5}, {0xec5, s.right, 19, 0x41dc}, {0xecb, uint16(width), 22, 0x41e4}, {0xec7, s.bottom, 24, 0x41eb}, {0xecd, s.bottom - s.top + 1, 27, 0x41f3}, {0xec9, s.flags, 29, 0x41fa},
		} {
			emit(base+uint32(w.off), w.value, w.step, w.ip)
		}
		for _, w := range []struct {
			off, value uint16
			step       uint64
			ip         uint16
		}{
			{0xedb, root.flags, 32, 0x420f}, {0xecf, root.seg, 34, 0x4216}, {0xed1, root.extra, 36, 0x421d}, {0xed5, root.top, 38, 0x4224}, {0xed9, root.bottom, 40, 0x422b}, {0xedf, root.bottom - root.top + 1, 43, 0x4233}, {0xed3, root.left, 45, 0x423a}, {0xed7, root.right, 47, 0x4241}, {0xedd, uint16(stride), 50, 0x4249}, {0x1129, r.color, 54, 0x4255}, {0x113b, s.left, 56, 0x425b}, {0x113d, s.right, 58, 0x4261},
		} {
			emit(base+uint32(w.off), w.value, w.step+4*uint64(r.links), w.ip)
		}
		p.After.R[cpu.SP] -= 8
		p.After.R[cpu.BP] = v.SP() - 2
		p.After.Seg[cpu.DS] = 0x48a1
	}
	p.After.Seg[cpu.ES] = root.seg
	color := uint16(byte(r.color)) * 257
	for row := 0; row < rows; row++ {
		y := int(r.y) + row
		off := (y-int(root.top))*stride + int(s.left-root.left)
		if off < 0 || off+width > 64000 {
			return nil
		}
		n := prefixN + uint64(row)*rowSteps
		if width/2 > 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + uint32(off), Value: color, Count: uint16(width / 2), Width: 2, Step: n + wordStep, CS: 0x48a1, IP: 0x42a1})
		}
		if width%2 != 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + uint32(off+width/2*2), Value: color, Count: 1, Width: 1, Step: n + wordStep + 2, CS: 0x48a1, IP: 0x42a5})
		}
		p.After.R[cpu.DI] = uint16(off + width)
	}
	y := r.y + uint16(rows)
	p.After.R[cpu.AX] = color
	p.After.R[cpu.CX] = 0
	p.After.R[cpu.DX] = 0
	p.After.R[cpu.BX] = y
	if stride == 320 {
		p.After.R[cpu.DX] = (y - 1 - root.top) * 64
	}
	p.After.Flags = nativeCompareFlags(v.Flags(), y, s.bottom)
	p.After.IP = 0x4268
	if complete {
		p.Instructions += 5
		p.Cycles += 35
		p.LastIP = 0x42b2
		if initial {
			p.After.R[cpu.BP] = v.BP()
			p.After.R[cpu.SI] = v.SI()
			p.After.R[cpu.DI] = v.DI()
			p.After.Seg[cpu.DS] = v.DS()
			p.After.R[cpu.SP] = v.SP() + 4
			p.After.IP = v.Read16(stack)
			p.After.Seg[cpu.CS] = v.Read16(stack + 2)
		} else {
			p.After.R[cpu.DI] = v.Read16(stack)
			p.After.R[cpu.SI] = v.Read16(stack + 2)
			p.After.Seg[cpu.DS] = v.Read16(stack + 4)
			p.After.R[cpu.BP] = v.Read16(stack + 6)
			p.After.R[cpu.SP] = v.SP() + 12
			p.After.IP = v.Read16(stack + 8)
			p.After.Seg[cpu.CS] = v.Read16(stack + 10)
		}
	}
	if v.DOSBoxCost() {
		p.Cycles = p.Instructions + uint64(rows*(width/2+width%2))
	}
	return p
}
