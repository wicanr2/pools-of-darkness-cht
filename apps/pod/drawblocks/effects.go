// Fixed drawing effects, spec025/#47. No guest instruction decoding occurs here.
package drawblocks

import (
	"crypto/sha256"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"math/bits"
)

type CPUView interface {
	Read8(uint32) byte
	Read16(uint32) uint16
	Steps() uint64
	AX() uint16
	CX() uint16
	DX() uint16
	BX() uint16
	SP() uint16
	BP() uint16
	SI() uint16
	DI() uint16
	ES() uint16
	CS() uint16
	SS() uint16
	DS() uint16
	IP() uint16
	Flags() uint16
	DOSBoxCost() bool
}

type effect struct {
	v              CPUView
	s              machine.HostRegisters
	p              machine.HostWorkPlan
	budget         uint64
	shadow         map[uint32]byte
	dos            bool
	spanLo, spanHi uint32
	payload        int
	checked        [(blockCount + 63) / 64]uint64
	bad            bool
}
type block struct {
	id   uint16
	cs   uint16
	n    uint64
	size uint16
	sha  [32]byte
	run  func(*effect)
}

func Plan(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.Read8(0x449) != 0x13 {
		return nil
	}
	if p := planSmallCall(v, budget); p != nil {
		return p
	}
	at := uint32(v.CS())*16 + uint32(v.IP())
	if b, ok := blocks[at]; !ok || b.cs != v.CS() {
		return nil
	}
	e := &effect{v: v, dos: v.DOSBoxCost(), budget: budget, shadow: make(map[uint32]byte), s: machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: v.IP(), Flags: v.Flags()}}
	for e.p.Instructions < budget {
		at = uint32(e.s.Seg[1])*16 + uint32(e.s.IP)
		b, ok := blocks[at]
		if !ok || b.cs != e.s.Seg[1] || b.n > budget-e.p.Instructions || len(e.p.Writes) >= 131072 || e.payload >= 131072 {
			break
		}
		mask := uint64(1) << (b.id & 63)
		if e.checked[b.id>>6]&mask == 0 {
			var code [128]byte
			if int(b.size) > len(code) {
				break
			}
			copyView(e.v, at, code[:b.size])
			if sha256.Sum256(code[:b.size]) != b.sha {
				break
			}
			e.checked[b.id>>6] |= mask
		}
		before := e.s
		instr, cycles, lastCS, lastIP := e.p.Instructions, e.p.Cycles, e.p.LastCS, e.p.LastIP
		nw, nt := len(e.p.Writes), len(e.p.Transfers)
		b.run(e)
		if e.bad {
			e.s = before
			e.p.Instructions, e.p.Cycles, e.p.LastCS, e.p.LastIP = instr, cycles, lastCS, lastIP
			e.p.Writes = e.p.Writes[:nw]
			e.p.Transfers = e.p.Transfers[:nt]
			break
		}
	}
	if e.p.Instructions == 0 {
		return nil
	}
	e.p.After = e.s
	return &e.p
}
func (e *effect) mark(ip uint16, cost uint64) {
	e.p.Instructions++
	e.p.LastCS = e.s.Seg[1]
	e.p.LastIP = ip
	if e.dos {
		e.p.Cycles++
	} else {
		e.p.Cycles += cost
	}
}
func (e *effect) charge(n uint64) {
	if !e.dos {
		e.p.Cycles += n
	}
}
func (e *effect) read(seg, off uint16, size int) uint16 {
	a := uint32(seg)*16 + uint32(off)
	if a >= 0xafa00 {
		e.bad = true
		return 0
	}
	v, ok := e.shadow[a]
	if !ok {
		v = e.v.Read8(a)
		if a >= e.spanLo && a < e.spanHi {
			for i := len(e.p.Transfers) - 1; i >= 0; i-- {
				s := e.p.Transfers[i]
				lo, hi := transferBounds(s)
				if a < lo || a >= hi {
					continue
				}
				index := int(a - s.Address)
				if s.Backward {
					d := int64(a) - int64(s.Address)
					rem := int(d % int64(s.Width))
					if rem < 0 {
						rem += int(s.Width)
					}
					index = int((-d+int64(rem))/int64(s.Width))*int(s.Width) + rem
				}
				v = byte(s.Value >> uint(8*(index%int(s.Width))))
				if s.Data != nil {
					v = s.Data[index]
				}
				break
			}
		}
	}
	if size == 8 {
		return uint16(v)
	}
	return uint16(v) | e.read(seg, off+1, 8)<<8
}
func protected(a uint32) bool {
	return a >= 0x48a10+0x2772 && a < 0x48a10+0x757f || a >= 0x1ed40+0x1a5c && a < 0x1ed40+0x1a93 || a >= 0x1d1f0+0x46e && a < 0x1d1f0+0x487
}
func (e *effect) write(seg, off, value uint16, size int) {
	a := uint32(seg)*16 + uint32(off)
	if a < 0x500 || a >= 0xafa00 || protected(a) {
		e.bad = true
		return
	}
	v := byte(value)
	e.shadow[a] = v
	e.p.Writes = append(e.p.Writes, machine.HostWrite{Address: a, Value: v, Step: e.p.Instructions, CS: e.p.LastCS, IP: e.p.LastIP})
	if size == 16 {
		e.write(seg, off+1, value>>8, 8)
	}
}
func (e *effect) reg8(n int) uint16 { return e.s.R[n&3] >> uint(8*(n/4)) & 255 }
func (e *effect) set8(n int, v uint16) {
	shift := uint(8 * (n / 4))
	e.s.R[n&3] = e.s.R[n&3]&^(255<<shift) | (v&255)<<shift
}
func (e *effect) push(v uint16) { e.charge(2); e.s.R[4] -= 2; e.write(e.s.Seg[2], e.s.R[4], v, 16) }
func (e *effect) pop() uint16 {
	e.charge(2)
	v := e.read(e.s.Seg[2], e.s.R[4], 16)
	e.s.R[4] += 2
	return v
}
func (e *effect) flag(f uint16) bool { return e.s.Flags&f != 0 }
func (e *effect) set(f uint16, on bool) {
	if on {
		e.s.Flags |= f
	} else {
		e.s.Flags &^= f
	}
}
func (e *effect) szp(v uint16, size int) {
	mask, sign := uint16(65535), uint16(32768)
	if size == 8 {
		mask, sign = 255, 128
	}
	e.set(cpu.ZF, v&mask == 0)
	e.set(cpu.SF, v&sign != 0)
	e.set(cpu.PF, bits.OnesCount8(byte(v))%2 == 0)
}
func (e *effect) alu(op int, a, b uint16, size int) uint16 {
	mask, sign := uint32(65535), uint32(32768)
	if size == 8 {
		mask, sign = 255, 128
	}
	x, y := uint32(a), uint32(b)
	var out uint32
	carry := uint32(0)
	if e.flag(cpu.CF) {
		carry = 1
	}
	switch op {
	case 0, 2:
		if op == 0 {
			carry = 0
		}
		out = x + y + carry
		e.set(cpu.CF, out&^mask != 0)
		e.set(cpu.AF, (x&15)+(y&15)+carry > 15)
		e.set(cpu.OF, (x^out)&(y^out)&sign != 0)
	case 3, 5, 7:
		if op != 3 {
			carry = 0
		}
		out = x - y - carry
		e.set(cpu.CF, out&^mask != 0)
		e.set(cpu.AF, x&15 < (y&15)+carry)
		e.set(cpu.OF, (x^y)&(x^out)&sign != 0)
	case 1:
		out = x | y
	case 4:
		out = x & y
	case 6:
		out = x ^ y
	}
	if op == 1 || op == 4 || op == 6 {
		e.set(cpu.CF|cpu.AF|cpu.OF, false)
	}
	out &= mask
	e.szp(uint16(out), size)
	return uint16(out)
}
func (e *effect) inc(a uint16, size int, down bool) uint16 {
	cf := e.flag(cpu.CF)
	op := 0
	if down {
		op = 5
	}
	v := e.alu(op, a, 1, size)
	e.set(cpu.CF, cf)
	return v
}
func (e *effect) neg(a uint16, size int) uint16 { return e.alu(5, 0, a, size) }
func (e *effect) shift(op int, v, n uint16, size int) uint16 {
	if n == 0 {
		return v
	}
	mask, sign := uint16(65535), uint16(32768)
	if size == 8 {
		mask, sign = 255, 128
	}
	x := v & mask
	for i := uint16(0); i < n; i++ {
		cf := e.flag(cpu.CF)
		last := false
		switch op {
		case 0:
			last = x&sign != 0
			x = x << 1 & mask
			if last {
				x |= 1
			}
			e.set(cpu.OF, (x&sign != 0) != last)
		case 1:
			last = x&1 != 0
			x >>= 1
			if last {
				x |= sign
			}
			e.set(cpu.OF, (x&sign != 0) != (x&(sign>>1) != 0))
		case 2:
			last = x&sign != 0
			x = x << 1 & mask
			if cf {
				x |= 1
			}
			e.set(cpu.OF, (x&sign != 0) != last)
		case 3:
			last = x&1 != 0
			x >>= 1
			if cf {
				x |= sign
			}
			e.set(cpu.OF, (x&sign != 0) != (x&(sign>>1) != 0))
		case 4:
			last = x&sign != 0
			x = x << 1 & mask
			e.set(cpu.OF, (x&sign != 0) != last)
		case 5:
			last = x&1 != 0
			e.set(cpu.OF, x&sign != 0)
			x >>= 1
		case 7:
			last = x&1 != 0
			x = x>>1 | x&sign
			e.set(cpu.OF, false)
		}
		e.set(cpu.CF, last)
	}
	if op >= 4 {
		e.szp(x, size)
		e.set(cpu.AF, false)
	}
	return x
}
func (e *effect) mul(v uint16, size int) {
	e.charge(14)
	if size == 8 {
		p := e.reg8(0) * v
		e.s.R[0] = p
		e.set(cpu.CF|cpu.OF, p > 255)
		e.szp(p, 16)
	} else {
		p := uint32(e.s.R[0]) * uint32(v)
		e.s.R[0] = uint16(p)
		e.s.R[2] = uint16(p >> 16)
		e.set(cpu.CF|cpu.OF, p>>16 != 0)
		e.szp(uint16(p>>16), 16)
		e.set(cpu.ZF, p == 0)
	}
}
func (e *effect) div(v uint16, size int) {
	e.charge(25)
	if size == 8 {
		e.alu(5, e.reg8(4), v, 8)
		if v == 0 || e.s.R[0]/v > 255 {
			e.bad = true
			return
		}
		a := e.s.R[0]
		e.s.R[0] = a/v | (a%v)<<8
	} else {
		e.alu(5, e.s.R[2], v, 16)
		a := uint32(e.s.R[2])<<16 | uint32(e.s.R[0])
		if v == 0 || a/uint32(v) > 65535 {
			e.bad = true
			return
		}
		e.s.R[0] = uint16(a / uint32(v))
		e.s.R[2] = uint16(a % uint32(v))
	}
}
func (e *effect) cond(n int) bool {
	cf, zf, sf, of, pf := e.flag(cpu.CF), e.flag(cpu.ZF), e.flag(cpu.SF), e.flag(cpu.OF), e.flag(cpu.PF)
	switch n {
	case 0:
		return of
	case 1:
		return !of
	case 2:
		return cf
	case 3:
		return !cf
	case 4:
		return zf
	case 5:
		return !zf
	case 6:
		return cf || zf
	case 7:
		return !cf && !zf
	case 8:
		return sf
	case 9:
		return !sf
	case 10:
		return pf
	case 11:
		return !pf
	case 12:
		return sf != of
	case 13:
		return sf == of
	case 14:
		return zf || sf != of
	case 15:
		return !zf && sf == of
	}
	return false
}
func (e *effect) str(op, rep, src int) {
	size := 8
	if op&1 != 0 {
		size = 16
	}
	d := uint16(size / 8)
	if e.flag(cpu.DF) {
		d = -d
	}
	count := uint16(1)
	if rep != 0 {
		count = e.s.R[1]
	}
	if count > 0 && (op == 0xa4 || op == 0xa5 || op == 0xaa || op == 0xab) && e.transfer(op, rep, src, count, size, d) {
		return
	}
	for i := uint16(0); i < count; i++ {
		e.charge(5)
		if e.dos {
			e.p.Cycles++
		}
		ss, ds := e.s.Seg[src], e.s.Seg[0]
		si, di := e.s.R[6], e.s.R[7]
		switch op {
		case 0xa4, 0xa5:
			e.write(ds, di, e.read(ss, si, size), size)
			e.s.R[6] += d
			e.s.R[7] += d
		case 0xa6, 0xa7:
			e.alu(5, e.read(ss, si, size), e.read(ds, di, size), size)
			e.s.R[6] += d
			e.s.R[7] += d
		case 0xaa, 0xab:
			e.write(ds, di, e.s.R[0], size)
			e.s.R[7] += d
		case 0xac:
			e.set8(0, e.read(ss, si, size))
			e.s.R[6] += d
		case 0xad:
			e.s.R[0] = e.read(ss, si, size)
			e.s.R[6] += d
		case 0xae, 0xaf:
			e.alu(5, e.s.R[0]&uint16((1<<uint(size))-1), e.read(ds, di, size), size)
			e.s.R[7] += d
		}
		if e.bad {
			return
		}
		if rep != 0 {
			e.s.R[1]--
			if (op == 0xa6 || op == 0xa7 || op == 0xae || op == 0xaf) && e.flag(cpu.ZF) != (rep == 0xf3) {
				break
			}
		}
	}
}

func transferBounds(s machine.HostTransfer) (uint32, uint32) {
	lo, hi := s.Address, s.Address+uint32(s.Width)*uint32(s.Count)
	if s.Backward {
		lo = s.Address - uint32(s.Count-1)*uint32(s.Width)
		hi = s.Address + uint32(s.Width)
	}
	return lo, hi
}
func overlapCode(lo, hi uint32) bool {
	for _, r := range [][2]uint32{{0x48a10 + 0x2772, 0x48a10 + 0x757f}, {0x1ed40 + 0x1a5c, 0x1ed40 + 0x1a93}, {0x1d1f0 + 0x46e, 0x1d1f0 + 0x487}} {
		if lo < r[1] && r[0] < hi {
			return true
		}
	}
	return false
}
func (e *effect) transfer(op, rep, src int, count uint16, size int, d uint16) bool {
	width := uint8(size / 8)
	back := e.flag(cpu.DF)
	off := int(e.s.R[7])
	lo, hi := off, off+int(width)*int(count)
	if back {
		lo = off - (int(count)-1)*int(width)
		hi = off + int(width)
	}
	if lo < 0 || hi > 65536 {
		return false
	}
	base := uint32(e.s.Seg[0]) * 16
	dlo, dhi := base+uint32(lo), base+uint32(hi)
	if dlo < 0x500 || dhi > 0xafa00 || dlo < 0xa0000 && dhi > 0xa0000 || overlapCode(dlo, dhi) {
		return false
	}
	s := machine.HostTransfer{Address: base + uint32(off), Width: width, Count: count, Backward: back, Value: e.s.R[0], Step: e.p.Instructions, CS: e.p.LastCS, IP: e.p.LastIP}
	if op == 0xa4 || op == 0xa5 {
		so := int(e.s.R[6])
		sl, sh := so, so+int(width)*int(count)
		if back {
			sl = so - (int(count)-1)*int(width)
			sh = so + int(width)
		}
		sb := uint32(e.s.Seg[src]) * 16
		if sl < 0 || sh > 65536 || sb+uint32(sh) > 0xafa00 || sb+uint32(sl) < dhi && dlo < sb+uint32(sh) {
			return false
		}
		s.Data = make([]byte, int(width)*int(count))
		e.copyBytes(sb+uint32(sl), s.Data)
		if back {
			for i, j := 0, int(count)-1; i < j; i, j = i+1, j-1 {
				for b := 0; b < int(width); b++ {
					s.Data[i*int(width)+b], s.Data[j*int(width)+b] = s.Data[j*int(width)+b], s.Data[i*int(width)+b]
				}
			}
		}
		e.s.R[6] += count * d
	}
	if len(e.p.Transfers) == 0 || dlo < e.spanLo {
		e.spanLo = dlo
	}
	if dhi > e.spanHi {
		e.spanHi = dhi
	}
	for a := range e.shadow {
		if dlo <= a && a < dhi {
			delete(e.shadow, a)
		}
	}
	e.p.Transfers = append(e.p.Transfers, s)
	e.payload += int(count) * int(width)
	e.s.R[7] += count * d
	if rep != 0 {
		e.s.R[1] = 0
	}
	e.charge(5 * uint64(count))
	if e.dos {
		e.p.Cycles += uint64(count)
	}
	return true
}

func copyView(v CPUView, a uint32, dst []byte) {
	if c, ok := v.(interface{ CopyRAM(uint32, []byte) bool }); ok && c.CopyRAM(a, dst) {
		return
	}
	for i := range dst {
		dst[i] = v.Read8(a + uint32(i))
	}
}
func (e *effect) copyBytes(a uint32, dst []byte) {
	copyView(e.v, a, dst)
	end := a + uint32(len(dst))
	if a < e.spanHi && e.spanLo < end {
		for _, s := range e.p.Transfers {
			lo, hi := transferBounds(s)
			if lo >= end || hi <= a {
				continue
			}
			if lo < a {
				lo = a
			}
			if hi > end {
				hi = end
			}
			for at := lo; at < hi; at++ {
				index := int(at - s.Address)
				if s.Backward {
					d := int64(at) - int64(s.Address)
					r := int(d % int64(s.Width))
					if r < 0 {
						r += int(s.Width)
					}
					index = int((-d+int64(r))/int64(s.Width))*int(s.Width) + r
				}
				value := byte(s.Value >> uint(8*(index%int(s.Width))))
				if s.Data != nil {
					value = s.Data[index]
				}
				dst[at-a] = value
			}
		}
	}
	for at, value := range e.shadow {
		if a <= at && at < end {
			dst[at-a] = value
		}
	}
}
