// Spec025/#49: budgeted Go memory/image spans.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"math/bits"
)

type boundedRLELayout struct {
	entry, header uint16
	prefix        [5]uint16
	literal       [7]uint16
	repeat        [9]uint16
	code          smallCode
}

func boundedRLELayouts() []boundedRLELayout {
	return []boundedRLELayout{
		{0x68f4, 0x689e, [5]uint16{0x68f4, 0x68f5, 0x68f7, 0x68f9, 0x68fb}, [7]uint16{0x68fd, 0x68ff, 0x6901, 0x6903, 0x6905, 0x6907, 0x6909}, [9]uint16{0x690b, 0x690c, 0x690e, 0x6910, 0x6912, 0x6914, 0x6916, 0x6918, 0x691a}, rleCode},
		{0x6b33, 0x6abe, [5]uint16{0x6b33, 0x6b34, 0x6b36, 0x6b38, 0x6b3a}, [7]uint16{0x6b3c, 0x6b3e, 0x6b40, 0x6b42, 0x6b44, 0x6b46, 0x6b48}, [9]uint16{0x6b4b, 0x6b4c, 0x6b4e, 0x6b50, 0x6b52, 0x6b54, 0x6b56, 0x6b58, 0x6b5a}, flippedRLECode},
	}
}

func boundedRLEShiftFlags(flags, value uint16, width uint8) uint16 {
	mask, sign := uint16(0xffff), uint16(0x8000)
	if width == 8 {
		mask, sign = 255, 128
	}
	value &= mask
	out := value / 2
	f := flags &^ uint16(cpu.CF|cpu.PF|cpu.AF|cpu.ZF|cpu.SF|cpu.OF)
	if value&1 != 0 {
		f |= cpu.CF
	}
	if value&sign != 0 {
		f |= cpu.OF
	}
	if out == 0 {
		f |= cpu.ZF
	}
	if out&sign != 0 {
		f |= cpu.SF
	}
	if bits.OnesCount8(byte(out))%2 == 0 {
		f |= cpu.PF
	}
	return f
}
func boundedRLEIncrementFlags(flags uint16, value byte) uint16 {
	out := value + 1
	f := flags &^ uint16(cpu.PF|cpu.AF|cpu.ZF|cpu.SF|cpu.OF)
	if value == 127 {
		f |= cpu.OF
	}
	if value&15 == 15 {
		f |= cpu.AF
	}
	if out == 0 {
		f |= cpu.ZF
	}
	if out&128 != 0 {
		f |= cpu.SF
	}
	if bits.OnesCount8(out)%2 == 0 {
		f |= cpu.PF
	}
	return f
}

// A parsed image packet owns COPY or FILL bytes. Guest instruction locations
// identify observable ABI checkpoints, never the pixel-generation algorithm.
func planBoundedRLE(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x48a1 || v.ES() != 0xa000 || v.Flags()&cpu.DF != 0 || v.CX()&0xff00 != 0 || v.Read8(0x449) != 0x13 {
		return nil
	}
	for _, layout := range boundedRLELayouts() {
		known := false
		for _, ip := range layout.prefix {
			known = known || v.IP() == ip
		}
		for _, ip := range layout.literal {
			known = known || v.IP() == ip
		}
		for _, ip := range layout.repeat {
			known = known || v.IP() == ip
		}
		if !known {
			continue
		}
		var raw [42]byte
		copyView(v, 0x48a10+uint32(layout.entry), raw[:layout.code.size])
		if sha256.Sum256(raw[:layout.code.size]) != layout.code.sum {
			return nil
		}
		base, end := uint32(v.DS())*16, uint32(v.BX())
		if base < 0x500 || base+end > 0xa0000 {
			return nil
		}
		si, n := int(v.SI()), int(byte(v.AX()))+1
		starts := []int{si, si - 1, si - 2, si - 1 - n, si - 1 - n/2*2}
		for _, start := range starts {
			if start < 0 || uint32(start) >= end {
				continue
			}
			control := v.Read8(base + uint32(start))
			count := uint16(control>>1) + 1
			repeat := control&1 != 0
			used := uint16(1) + count
			if repeat {
				used = 2
			}
			if uint32(start)+uint32(used) > end {
				continue
			}
			for _, emitted := range []uint16{0, count / 2 * 2, count} {
				if v.DI() < emitted {
					continue
				}
				dest := v.DI() - emitted
				if uint32(dest)+uint32(count) > 64000 {
					continue
				}
				if p := boundedRLEPacket(v, budget, layout, uint16(start), dest, control, count, used, repeat); p != nil {
					return p
				}
			}
		}
		return nil
	}
	return nil
}

func boundedRLEPacket(v CPUView, budget uint64, l boundedRLELayout, start, dest uint16, control byte, count, used uint16, repeat bool) *machine.HostWorkPlan {
	current := boundedMemoryRegisters(v)
	r := current
	r.IP = l.entry
	r.R[cpu.SI] = start
	r.R[cpu.DI] = dest
	steps := []boundedMemoryCheckpoint{}
	add := func(next uint16, native, dos uint64, tr *machine.HostTransfer, change func()) {
		before := r
		change()
		r.IP = next
		steps = append(steps, boundedMemoryCheckpoint{before: before, after: r, nativeCycles: native, dosCycles: dos, transfer: tr})
	}
	none := func() {}
	base := uint32(v.DS()) * 16
	add(l.prefix[1], 7, 2, nil, func() { r.R[cpu.AX] = r.R[cpu.AX]&0xff00 | uint16(control); r.R[cpu.SI]++ })
	add(l.prefix[2], 2, 1, nil, func() {
		r.Flags = boundedRLEShiftFlags(r.Flags, uint16(control), 8)
		r.R[cpu.AX] = r.R[cpu.AX]&0xff00 | uint16(control>>1)
	})
	add(l.prefix[3], 2, 1, nil, func() { r.R[cpu.CX] = uint16(control >> 1) })
	add(l.prefix[4], 2, 1, nil, func() { r.Flags = boundedRLEIncrementFlags(r.Flags, byte(r.R[cpu.CX])); r.R[cpu.CX] = count })
	body := l.literal[:]
	if repeat {
		body = l.repeat[:]
	}
	branch := uint64(2)
	if repeat {
		branch = 7
	}
	add(body[0], branch, 1, nil, none)
	var data []byte
	value := uint16(0)
	shiftIndex := 0
	if repeat {
		value = uint16(v.Read8(base+uint32(start)+1)) * 257
		add(body[1], 7, 2, nil, func() { r.R[cpu.AX] = r.R[cpu.AX]&0xff00 | value&255; r.R[cpu.SI]++ })
		add(body[2], 2, 1, nil, func() { r.R[cpu.AX] = value })
		shiftIndex = 2
	} else {
		data = make([]byte, int(count))
		copyView(v, base+uint32(start)+1, data)
	}
	add(body[shiftIndex+1], 2, 1, nil, func() { r.Flags = boundedRLEShiftFlags(r.Flags, count, 16); r.R[cpu.CX] = count / 2 })
	words := count / 2
	tail := count % 2
	var wordTransfer *machine.HostTransfer
	if words != 0 {
		wordTransfer = &machine.HostTransfer{Address: 0xa0000 + uint32(dest), Value: value, Count: words, Width: 2, CS: 0x48a1, IP: body[shiftIndex+1]}
		if data != nil {
			wordTransfer.Data = append([]byte(nil), data[:int(words)*2]...)
		}
	}
	add(body[shiftIndex+2], 3+5*uint64(words), 1+uint64(words), wordTransfer, func() {
		r.R[cpu.CX] = 0
		r.R[cpu.DI] += words * 2
		if !repeat {
			r.R[cpu.SI] += words * 2
		}
	})
	add(body[shiftIndex+3], 2, 1, nil, func() { r.R[cpu.CX] = tail; r.Flags &^= uint16(cpu.CF | cpu.OF) })
	var byteTransfer *machine.HostTransfer
	if tail != 0 {
		byteTransfer = &machine.HostTransfer{Address: 0xa0000 + uint32(dest) + uint32(words)*2, Value: value, Count: 1, Width: 1, CS: 0x48a1, IP: body[shiftIndex+3]}
		if data != nil {
			byteTransfer.Data = append([]byte(nil), data[len(data)-1:]...)
		}
	}
	add(body[shiftIndex+4], 3+5*uint64(tail), 1+uint64(tail), byteTransfer, func() {
		r.R[cpu.CX] = 0
		r.R[cpu.DI] += tail
		if !repeat {
			r.R[cpu.SI] += tail
		}
	})
	add(body[shiftIndex+5], 2, 1, nil, func() { r.Flags = nativeCompareFlags(r.Flags, start+used, v.BX()) })
	if start+used == v.BX() {
		add(body[shiftIndex+6], 2, 1, nil, none)
		add(l.header, 7, 1, nil, none)
	} else {
		add(l.entry, 7, 1, nil, none)
	}
	first := -1
	for i, s := range steps {
		if s.before == current {
			first = i
			break
		}
	}
	if first < 0 {
		return nil
	}
	n := len(steps) - first
	if uint64(n) > budget {
		n = int(budget)
	}
	p := &machine.HostWorkPlan{Algorithmic: true, Instructions: uint64(n), After: steps[first+n-1].after, LastCS: 0x48a1, LastIP: steps[first+n-1].before.IP}
	for i := 0; i < n; i++ {
		s := steps[first+i]
		if v.DOSBoxCost() {
			p.Cycles += s.dosCycles
		} else {
			p.Cycles += s.nativeCycles
		}
		if s.transfer != nil {
			tr := *s.transfer
			tr.Step = uint64(i + 1)
			p.Transfers = append(p.Transfers, tr)
		}
	}
	return p
}
