// Typed Go integer surface scaling and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

type scale5615SpanView struct {
	CPUView
	sp, dx uint16
}

func (v scale5615SpanView) SP() uint16 { return v.sp }
func (v scale5615SpanView) DX() uint16 { return v.dx }
func scale5615PushedFrame(v CPUView) (f scale5615Frame, row uint16, ok bool) {
	stack, good := paletteSourceStack(v, 0, 26)
	if !good || v.SP() > 65533 {
		return f, 0, false
	}
	factor, height := v.Read16(stack+24), v.Read16(0x48a10+0xef1)
	if factor == 0 || factor > 8 || height == 0 || height > 200 {
		return f, 0, false
	}
	cy := int32(int16(v.Read16(stack + 20)))
	y := int32(int16(v.Read16(0x48a10 + 0x1171)))
	startY := cy - int32((height-1)/2/factor)
	if y < startY || y-startY > int32((height-1)/factor) {
		return f, 0, false
	}
	destY := v.Read16(0x48a10+0xee7) + uint16(y-startY)*factor
	f, good = readScale5615Frame(scale5615SpanView{v, v.SP() + 2, destY})
	if !good {
		return f, 0, false
	}
	row = v.Read16(stack)
	if row < f.destY || row > f.d.bottom || row-f.destY >= f.factor {
		return f, 0, false
	}
	return f, row, true
}
func planScale5615Span(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x48a1 {
		return nil
	}
	p := imageBasePlan(v)
	var count uint16
	var width uint8
	var at uint32
	var data []byte
	var value uint16
	copying := false
	switch v.IP() {
	case 0x5813:
		f, row, ok := scale5615PushedFrame(v)
		if !ok || row != f.destY || v.ES() != v.CS() || f.y < int32(f.s.top) || f.y > int32(f.s.bottom) {
			return nil
		}
		x := int32(int16(v.DX()))
		sample := x - f.x
		if sample < 0 || sample > int32(f.samples) || v.BX() != f.samples-uint16(sample) || v.CX() != f.factor || v.SI() != uint16(f.y*320+f.x+sample+1) || v.DI() != 0x1252+uint16(sample)*f.factor {
			return nil
		}
		expected := f.color
		if x >= int32(f.s.left) && x <= int32(f.s.right) {
			high := uint16(0x100)
			if f.x < int32(f.s.left) {
				high = f.color & 0xff00
			}
			expected = high | uint16(v.Read8(0xa0000+uint32(f.y*320+x)))
		}
		if v.AX() != expected {
			return nil
		}
		count, width, at, value = f.factor, 1, 0x48a10+uint32(v.DI()), expected&255
	case 0x5879, 0x587d:
		f, ok := readScale5615Frame(v)
		if !ok || v.ES() != v.CS() || (f.y >= int32(f.s.top) && f.y <= int32(f.s.bottom)) || v.SI() != uint16(f.y) || v.AX() != (f.color&255)*257 {
			return nil
		}
		count, width, at, value = f.samples*f.factor*2, 2, 0x48a10+0x1252, (f.color&255)*257
		if v.IP() == 0x587d {
			at += uint32(count) * 2
			count, width = 0, 1
		}
		if uint32(v.DI())+0x48a10 != at || v.CX() != count {
			return nil
		}
	case 0x5856, 0x585a:
		f, row, ok := scale5615PushedFrame(v)
		if !ok || v.ES() != 0xa000 || v.AX() != 320 || v.DX() != row*64 {
			return nil
		}
		done := uint16(0)
		count, width = f.width/2, 2
		if v.IP() == 0x585a {
			done = count * 2
			count, width = f.width%2, 1
		}
		at = 0xa0000 + uint32(row)*320 + uint32(f.d.left+done)
		if v.SI() != 0x1252+done || uint32(v.DI())+0xa0000 != at || v.CX() != count {
			return nil
		}
		copying = true
		data = make([]byte, int(count)*int(width))
		copyView(v, 0x48a10+0x1252+uint32(done), data)
	default:
		return nil
	}
	p.Instructions, p.Cycles, p.LastIP = 1, 3+uint64(count)*5, v.IP()
	if v.DOSBoxCost() {
		p.Cycles = 1 + uint64(count)
	}
	if count > 0 {
		p.Transfers = []machine.HostTransfer{{Address: at, Data: data, Value: value, Count: count, Width: width, Step: 1, CS: 0x48a1, IP: v.IP()}}
	}
	p.After.R[cpu.CX] = 0
	p.After.R[cpu.DI] += count * uint16(width)
	if copying {
		p.After.R[cpu.SI] += count * uint16(width)
	}
	p.After.IP += 2
	return p
}
