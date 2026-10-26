package drawblocks

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

type smallCode struct {
	size int
	sum  [32]byte
}

func smallFingerprint(n int, s string) smallCode {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 32 {
		panic("drawing code fingerprint")
	}
	c := smallCode{size: n}
	copy(c.sum[:], b)
	return c
}

var smallCalls = map[uint32]smallCode{
	0x1ed41a5c: smallFingerprint(35, "1941e7e4eafb9300636ac7ff777605d8ea0b96654893d1198ca3d5f0bfd93972"),
	0x1ed41a7f: smallFingerprint(20, "f1c5fade96970aa1dd1a9ffc2d876c7aea57bef8680d929d9f65d6c5190e17f7"),
	0x1d1f046e: smallFingerprint(25, "ebe07354f00d07b7c388ccbf5a4bd13a2fd9a5ba8aff103883d64463f7d936dd"),
}

// planSmallCall is the complete ABI effect of the three shared memory
// primitives. Larger or uncertain cases retain the verified block path.
func planSmallCall(v CPUView, budget uint64) *machine.HostWorkPlan {
	key := uint32(v.CS())<<16 | uint32(v.IP())
	code, ok := smallCalls[key]
	if !ok {
		return nil
	}
	var raw [35]byte
	copyView(v, uint32(v.CS())*16+uint32(v.IP()), raw[:code.size])
	if sha256.Sum256(raw[:code.size]) != code.sum {
		return nil
	}
	b := machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: v.IP(), Flags: v.Flags()}
	args := uint16(14)
	if key == 0x1ed41a7f {
		args = 12
	}
	if key == 0x1d1f046e {
		args = 10
	}
	sp := uint32(v.SP())
	sb := uint32(v.SS()) * 16
	if sp+uint32(args) > 65536 || sb+sp+uint32(args) > 0xa0000 {
		return nil
	}
	word := func(off uint32) uint16 { return v.Read16(sb + sp + off) }
	count := word(4)
	off, seg := word(6), word(8)
	value := uint16(0)
	back := false
	srcOff, srcSeg := uint16(0), uint16(0)
	n, cycles := uint64(9), uint64(74+5*uint64(count/2))
	last := uint16(0x484)
	e := effect{s: b}
	e.s.R[3] = v.SP()
	e.s.R[1] = 0
	e.s.R[4] += args
	e.s.Flags &^= cpu.DF
	switch key {
	case 0x1ed41a5c:
		srcOff, srcSeg = word(10), word(12)
		back = srcOff < off
		n = 11
		cycles = 119 + 5*uint64(count)
		if back {
			n = 16
			cycles = 124 + 5*uint64(count)
		}
		last = 0x1a7c
		e.alu(5, srcOff, off, 16)
		e.s.Flags &^= cpu.DF
		if back {
			a := e.alu(0, srcOff, count, 16)
			d := e.alu(0, off, count, 16)
			e.inc(a, 16, true)
			e.inc(d, 16, true)
			e.s.Flags |= cpu.DF
		}
		e.s.R[2] = b.Seg[3]
		e.s.R[6] = srcOff + count
		e.s.R[7] = off + count
		if back {
			e.s.R[6] = srcOff - 1
			e.s.R[7] = off - 1
		}
	case 0x1ed41a7f:
		value = word(4) & 255
		count = word(6)
		off, seg = word(8), word(10)
		n = 7
		cycles = 70 + 5*uint64(count)
		last = 0x1a90
		e.s.R[0] = b.R[0]&0xff00 | value
		e.s.R[7] = off + count
	case 0x1d1f046e:
		e.alu(6, 0, 0, 16)
		e.shift(5, count, 1, 16)
		e.s.Flags &^= cpu.DF
		e.s.R[0] = 0
		e.s.R[7] = off + count
		if count&1 != 0 {
			if count == 1 {
				n, cycles = 10, 75
			} else {
				n = 11
				cycles = 78 + 5*uint64(count/2)
			}
		}
	}
	if n > budget {
		return nil
	}
	p := &machine.HostWorkPlan{Algorithmic: true, After: e.s, Instructions: n, Cycles: cycles, LastCS: b.Seg[1], LastIP: last}
	if v.DOSBoxCost() {
		iterations := uint64(count)
		if key == 0x1d1f046e {
			iterations = uint64(count/2 + count%2)
		}
		p.Cycles = n + iterations
	}
	p.After.Seg[0] = seg
	p.After.IP = word(0)
	p.After.Seg[1] = word(2)
	if count == 0 {
		return p
	}
	base := uint32(seg) * 16
	lo, hi := uint32(off), uint32(off)+uint32(count)
	dest := base + uint32(off)
	if hi > 65536 || base+lo < 0x500 || base+hi > 0xafa00 || base+lo < 0xa0000 && base+hi > 0xa0000 || overlapCode(base+lo, base+hi) || base+lo < sb+sp+uint32(args) && sb+sp < base+hi {
		return nil
	}
	data := []byte(nil)
	if key == 0x1ed41a5c {
		srcBase := uint32(srcSeg) * 16
		sl, sh := uint32(srcOff), uint32(srcOff)+uint32(count)
		if sh > 65536 || srcBase+sh > 0xafa00 {
			return nil
		}
		dstart, sstart := base+lo, srcBase+sl
		if dstart < srcBase+sh && sstart < base+hi && (back && dstart < sstart || !back && dstart > sstart) {
			return nil
		}
		data = make([]byte, int(count))
		copyView(v, sstart, data)
		if back {
			for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
				data[i], data[j] = data[j], data[i]
			}
			dest += uint32(count) - 1
		}
	}
	transfer := func(addr uint32, width uint8, num uint16, step uint64, ip uint16, bytes []byte) {
		if num != 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: addr, Width: width, Count: num, Backward: back, Value: value, Data: bytes, Step: step, CS: b.Seg[1], IP: ip})
		}
	}
	switch key {
	case 0x1ed41a5c:
		transfer(dest, 1, count, n-2, 0x1a78, data)
	case 0x1ed41a7f:
		transfer(dest, 1, count, 6, 0x1a8e, nil)
	case 0x1d1f046e:
		if count&1 != 0 {
			transfer(dest, 1, 1, 8, 0x47f, nil)
			dest++
		}
		step := uint64(8)
		if count&1 != 0 {
			step = 10
		}
		transfer(dest, 2, count/2, step, 0x482, nil)
	}
	return p
}
