package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func axisSpanPixelIP(ip uint16) bool {
	return ip == 0x2e6d || ip == 0x2e71 || (ip >= 0x2ed4 && ip <= 0x2f21 && (ip-0x2ed4)%11 == 0)
}

// Coordinates and the selected/root surfaces determine the span. The ten
// original positions identify ABI checkpoints, never an opcode dispatcher.
func planAxisSpan(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x48a1 || !axisSpanPixelIP(v.IP()) || v.DS() != 0x48a1 || v.ES() != 0xa000 || v.Read8(0x449) != 0x13 || v.Flags()&cpu.DF != 0 {
		return nil
	}
	mv, ok := v.(interface{ CPUModel() (cpu.Model, bool) })
	if !ok {
		return nil
	}
	model, known := mv.CPUModel()
	if !known || model != cpu.Model80386 {
		return nil
	}
	var code [1338]byte
	copyView(v, 0x48a10+0x2cfc, code[:])
	if sha256.Sum256(code[:]) != axisLineCode.sum {
		return nil
	}
	stack, ok := paletteSourceStack(v, 0, 24)
	if !ok {
		return nil
	}
	slot := v.Read16(stack + 12)
	if slot > 255 {
		return nil
	}
	word := func(off uint16) uint16 { return v.Read16(0x48a10 + uint32(off)) }
	descriptor := func(index uint16) rectangleSurface {
		return rectangleSurface{word(0xbd + index), word(0x2bd + index), word(0x4bd + index), word(0x6bd + index), word(0x8bd + index), word(0xabd + index), word(0xcbd + index)}
	}
	s := descriptor(slot * 2)
	root := s
	links := 0
	for root.flags&64 != 0 {
		if links >= 8 || root.seg > 510 || root.seg&1 != 0 {
			return nil
		}
		root = descriptor(root.seg)
		links++
	}
	if root.seg != 0xa000 || root.left > root.right || root.right >= 320 || root.top > root.bottom || root.bottom >= 200 || s.left < root.left || s.left > s.right || s.right > root.right || s.top < root.top || s.top > s.bottom || s.bottom > root.bottom {
		return nil
	}
	x0, y0, x1, y1, color := v.Read16(stack+14), v.Read16(stack+16), v.Read16(stack+18), v.Read16(stack+20), v.Read16(stack+22)
	if x0 != x1 && y0 != y1 {
		return nil
	}
	distance := func(a, b uint16) int32 {
		d := int32(int16(a)) - int32(int16(b))
		if d < 0 {
			d = -d
		}
		return d
	}
	if distance(x0, x1) > 32767 || distance(y0, y1) > 32767 {
		return nil
	}
	if int16(y0) >= int16(y1) {
		x0, x1, y0, y1 = x1, x0, y1, y0
	}
	if v.BP() != y1-y0 {
		return nil
	}
	horizontal := y0 == y1
	firstX, lastX := x0, x1
	firstY, lastY := y0, y1
	if horizontal {
		if int16(firstX) > int16(lastX) {
			firstX, lastX = lastX, firstX
		}
		if int16(firstY) < int16(s.top) || int16(firstY) > int16(s.bottom) || int16(lastX) < int16(s.left) || int16(firstX) > int16(s.right) {
			return nil
		}
		if int16(firstX) < int16(s.left) {
			firstX = s.left
		}
		if int16(lastX) > int16(s.right) {
			lastX = s.right
		}
	} else {
		if int16(x0) < int16(s.left) || int16(x0) > int16(s.right) || int16(lastY) < int16(s.top) || int16(firstY) > int16(s.bottom) {
			return nil
		}
		if int16(firstY) < int16(s.top) {
			firstY = s.top
		}
		if int16(lastY) > int16(s.bottom) {
			lastY = s.bottom
		}
	}
	stride := root.right - root.left + 1
	wantY0, wantY1 := y0, y1
	if !horizontal {
		wantY0, wantY1 = firstY, lastY
	}
	for _, pair := range [][2]uint16{{0xebd, s.seg}, {0xebf, s.extra}, {0xec1, s.left}, {0xec3, s.top}, {0xec5, s.right}, {0xec7, s.bottom}, {0xec9, s.flags}, {0xecb, s.right - s.left + 1}, {0xecd, s.bottom - s.top + 1}, {0xecf, root.seg}, {0xed1, root.extra}, {0xed3, root.left}, {0xed5, root.top}, {0xed7, root.right}, {0xed9, root.bottom}, {0xedb, root.flags}, {0xedd, stride}, {0xedf, root.bottom - root.top + 1}, {0x1129, color}, {0x112b, x0}, {0x112d, x1}, {0x112f, wantY0}, {0x1131, wantY1}, {0x1149, x0 - x1}, {0x114d, uint16(distance(x0, x1))}, {0x1151, y1 - y0}} {
		if word(pair[0]) != pair[1] {
			return nil
		}
	}
	at := uint32(firstY-root.top)*uint32(stride) + uint32(firstX-root.left)
	p := &machine.HostWorkPlan{Algorithmic: true, Instructions: 1, After: boundedMemoryRegisters(v), LastCS: 0x48a1, LastIP: v.IP()}
	if horizontal {
		if v.IP() != 0x2e6d && v.IP() != 0x2e71 {
			return nil
		}
		if word(0x113b) != firstX || word(0x113d) != lastX {
			return nil
		}
		width := lastX - firstX + 1
		count, unit := width/2, uint16(2)
		if v.IP() == 0x2e71 {
			at += uint32(width/2) * 2
			count, unit = width%2, 1
			if v.Flags()&cpu.CF != 0 {
				return nil
			}
		} else if v.Flags()&cpu.CF != width%2 {
			return nil
		}
		pixel := color & 255
		value := pixel | pixel<<8
		if v.AX() != value || v.CX() != count || v.DI() != uint16(at) || at+uint32(count)*uint32(unit) > 64000 {
			return nil
		}
		p.Cycles = 3 + 5*uint64(count)
		if v.DOSBoxCost() {
			p.Cycles = 1 + uint64(count)
		}
		p.After.IP += 2
		p.After.R[cpu.CX] = 0
		p.After.R[cpu.DI] += count * unit
		if count > 0 {
			p.Transfers = []machine.HostTransfer{{Address: 0xa0000 + at, Value: value, Count: count, Width: uint8(unit), Step: 1, CS: 0x48a1, IP: v.IP()}}
		}
	} else {
		if v.IP() < 0x2ed4 || v.CX() == 0 || v.CX() > lastY-firstY+1 || v.AX() != color {
			return nil
		}
		done := lastY - firstY + 1 - v.CX()
		if uint16((v.IP()-0x2ed4)/11) != done%8 {
			return nil
		}
		at += uint32(done) * uint32(stride)
		if at >= 64000 || v.DI() != uint16(at) {
			return nil
		}
		p.Cycles = 7
		if v.DOSBoxCost() {
			p.Cycles = 1
		}
		p.After.IP += 3
		p.Writes = []machine.HostWrite{{Address: 0xa0000 + at, Value: byte(color), Step: 1, CS: 0x48a1, IP: v.IP()}}
	}
	return p
}

func axisSpanFixedPixels(writes []machine.HostWrite, transfers []machine.HostTransfer) bool {
	for _, x := range writes {
		if x.CS == 0x48a1 && axisSpanPixelIP(x.IP) {
			return true
		}
	}
	for _, x := range transfers {
		if x.CS == 0x48a1 && axisSpanPixelIP(x.IP) {
			return true
		}
	}
	return false
}
