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
	if p := planSurfaceSetup(v, budget); p != nil {
		return p
	}
	if p := planColorFillSpan(v, budget); p != nil {
		return p
	}
	if p := planResidentGlyphPixel(v, budget); p != nil {
		return p
	}
	if p := planSurfaceFillSpan(v, budget); p != nil {
		return p
	}
	if p := planAxisSpan(v, budget); p != nil {
		return p
	}
	if p := planSurfaceCopySpan(v, budget); p != nil {
		return p
	}
	if p := planPaletteSingle(v, budget); p != nil {
		return p
	}
	if v.IP() == 0x4892 {
		if p := planFadeDifferenceLayout(v, budget, fadeInRGBLayout); p != nil {
			return p
		}
	}
	if v.IP() == 0x48e5 {
		if p := planFadeBlendLayout(v, budget, fadeInRGBLayout); p != nil {
			return p
		}
	}

	if p := planGlyph5aa4Rows(v, budget); p != nil {
		return p
	}
	if p := planGlyph5aa4Spans(v, budget); p != nil {
		return p
	}
	if p := dissolve4337(v, budget); p != nil {
		return p
	}
	if p := flip5886(v, budget); p != nil {
		return p
	}
	if p := scale5615(v, budget); p != nil {
		return p
	}
	if p := bitmap51(v, budget); p != nil {
		return p
	}
	if budget == 0 || v.Read8(0x449) != 0x13 {
		return nil
	}
	if p := planSmallCall(v, budget); p != nil {
		return p
	}
	if p := planBoundedZero(v, budget); p != nil {
		return p
	}
	if p := planBoundedCopy(v, budget); p != nil {
		return p
	}
	if p := planLine(v, budget); p != nil {
		return p
	}
	if p := planPrimitivePixel(v, budget); p != nil {
		return p
	}
	if p := planFilledEllipseBounded(v, budget); p != nil {
		return p
	}
	if p := planEllipseBounded(v, budget); p != nil {
		return p
	}
	if p := planSurfaceCopy(v, budget); p != nil {
		return p
	}
	if p := planRegionBounded(v, budget); p != nil {
		return p
	}
	if p := planPoint(v, budget); p != nil {
		return p
	}
	if p := planWait(v, budget); p != nil {
		return p
	}
	if p := planImageInitialize(v, budget); p != nil {
		return p
	}
	if p := planImageReturn(v, budget); p != nil {
		return p
	}
	if p := planRectangle(v, budget); p != nil {
		return p
	}
	if p := planRLERuns(v, budget); p != nil {
		return p
	}
	if p := planBoundedRLE(v, budget); p != nil {
		return p
	}
	if p := planImageSpan(v, budget); p != nil {
		return p
	}
	if p := planImageRow(v, budget); p != nil {
		return p
	}
	if p := planImageControl(v, budget); p != nil {
		return p
	}
	if p := planClipSetup(v, budget); p != nil {
		return p
	}
	if p := planClippedRun(v, budget); p != nil {
		return p
	}
	if p := planFlipInitialize(v, budget); p != nil {
		return p
	}
	if p := planFlipRow(v, budget); p != nil {
		return p
	}
	if p := planFlipHeader(v, budget); p != nil {
		return p
	}
	if p := planFlipControl(v, budget); p != nil {
		return p
	}
	if p := planMirrorSpan(v, budget); p != nil {
		return p
	}
	if p := planMirrorPacket(v, budget); p != nil {
		return p
	}
	if p := planMirrorClipSetup(v, budget); p != nil {
		return p
	}
	if p := planMirrorClippedPacket(v, budget); p != nil {
		return p
	}
	if p := planPaletteScan(v, budget); p != nil {
		return p
	}
	if p := planPaletteStage(v, budget); p != nil {
		return p
	}
	if p := planPaletteInitialize(v, budget); p != nil {
		return p
	}
	if p := planPaletteSourceControl(v, budget); p != nil {
		return p
	}
	if p := planPaletteSourceSample(v, budget); p != nil {
		return p
	}
	if p := planPaletteSourceStore(v, budget); p != nil {
		return p
	}
	if p := planPaletteAdvance(v, budget); p != nil {
		return p
	}
	if p := planPaletteFrame(v, budget); p != nil {
		return p
	}
	if p := planMirrorPairs(v, budget); p != nil {
		return p
	}
	if p := planFadeDifference(v, budget); p != nil {
		return p
	}
	if p := planFadeBlend(v, budget); p != nil {
		return p
	}
	if p := planFadeColor(v, budget); p != nil {
		return p
	}
	if p := planPaletteList(v, budget); p != nil {
		return p
	}
	if p := planPaletteTarget(v, budget); p != nil {
		return p
	}
	if p := planPaletteError(v, budget); p != nil {
		return p
	}
	if p := planPaletteOutput(v, budget); p != nil {
		return p
	}
	if p := planPaletteOutputBounded(v, budget); p != nil {
		return p
	}
	if p := planPaletteRead(v, budget); p != nil {
		return p
	}
	if p := planDrawingJobBoundary(v, budget); p != nil {
		return p
	}
	at := uint32(v.CS())*16 + uint32(v.IP())
	if b, ok := blocks[at]; !ok || b.cs != v.CS() {
		return nil
	}
	e := &effect{v: v, dos: v.DOSBoxCost(), budget: budget, shadow: make(map[uint32]byte), s: machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: v.IP(), Flags: v.Flags()}}
	for e.p.Instructions < budget {
		if e.p.Instructions > 0 && algorithmBoundary(e.s.Seg[1], e.s.IP) {
			break
		}
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
		if e.bad || colorFillFixedPixels(e.p.Writes[nw:], e.p.Transfers[nt:]) || surfaceFillFixedPixels(e.p.Writes[nw:], e.p.Transfers[nt:]) || primitiveFixedPixels(e.p.Writes[nw:], e.p.Transfers[nt:]) || axisSpanFixedPixels(e.p.Writes[nw:], e.p.Transfers[nt:]) || surfaceCopyFixedRaster(e.p.Writes[nw:], e.p.Transfers[nt:]) || paletteFixedSingleWrites(e.p.Writes[nw:], e.p.Transfers[nt:]) || glyph5aa4FixedPixels(e.p.Writes[nw:], e.p.Transfers[nt:]) || dissolve4337FixedPixels(e.p.Writes[nw:], e.p.Transfers[nt:]) || flip5886FixedRaster(e.p.Writes[nw:], e.p.Transfers[nt:]) || scale5615FixedRaster(e.p.Writes[nw:], e.p.Transfers[nt:]) || bitmap51FixedRaster(e.p.Writes[nw:], e.p.Transfers[nt:]) || ellipseFixedPixelWrites(e.p.Writes[nw:]) || filledEllipseFixedTransfers(e.p.Transfers[nt:]) {
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

// Stop a fixed-effects prefix before an algorithm's entry, so its pixel
// work is offered to the typed Go path using the committed guest state.
func algorithmBoundary(cs, ip uint16) bool {
	if cs == 0x48a1 && drawingJobBoundaryIP(ip) {
		return true
	}
	if cs == 0x48a1 && (ip == 0x2786 || ip == 0x28d5) {
		return true
	}
	if cs == 0x1ed4 && ip == 0x1a8e {
		return true
	}
	if cs == 0x48a1 && surfaceFillPixelIP(ip) {
		return true
	}
	if cs == 0x48a1 && primitivePixelIP(ip) {
		return true
	}
	if cs == 0x48a1 && axisSpanPixelIP(ip) {
		return true
	}
	if cs == 0x48a1 && (ip == 0x4169 || ip == 0x4177) {
		return true
	}
	if cs == 0x48a1 && paletteSingleWriteIP(ip) {
		return true
	}
	if cs == 0x48a1 && (ip == 0x47ba || ip == 0x47d6 || ip == 0x4867 || ip == 0x4892 || ip == 0x48b3 || ip == 0x48e5) {
		return true
	}

	if cs == 0x48a1 && (ip == 0x5c00 || ip == 0x5c02 || ip == 0x5c26 || ip == 0x5c2c || ip == 0x5c2f) {
		return true
	}
	if cs == 0x48a1 {
		switch ip {
		case 0x4337, 0x456c, 0x45c4, 0x45ce, 0x4569, 0x45bb:
			return true
		case 0x5886, 0x5947, 0x59a9, 0x59c8, 0x59cf, 0x59d3, 0x59f1, 0x59f8, 0x59fc, 0x5a0c, 0x5a10:
			return true
		case 0x5615, 0x579c, 0x581a, 0x5813, 0x5879, 0x587d, 0x5856, 0x585a:
			return true
		case 0x5373, 0x53c5, 0x53d8, 0x53f6, 0x53fa, 0x53fc:
			return true
		case 0x39b2, 0x3b3d, 0x3cc0, 0x3bd8, 0x3bdc, 0x3c2a, 0x3c2e, 0x3d47, 0x3d4b, 0x3d99, 0x3d9d:
			return true
		case 0x346a, 0x35f4, 0x37f6, 0x3660, 0x36b6, 0x370c, 0x3762, 0x384e, 0x38a4, 0x38fa, 0x3950:
			return true
		case 0x3f5e, 0x415e, 0x309c, 0x30b9, 0x3206, 0x321c, 0x2cfc, 0x3418, 0x3338, 0x2b10, 0x2c04, 0x4b0b, 0x4b8d, 0x4b91, 0x4bc6, 0x4944, 0x4971, 0x4a91, 0x4aa4, 0x4aad, 0x4abd, 0x4ae3, 0x4af6:
			return true
		case 0x2aeb, 0x4a31, 0x4a72, 0x4b55:
			return true
		case 0x2a90, 0x2a9d, 0x2a9e, 0x2aa0, 0x2aa2, 0x2aa3, 0x2aa5, 0x2aa7:
			return true
		case 0x6733, 0x682d, 0x6e5e, 0x6a0c, 0x6a3a, 0x6a48, 0x6a50, 0x6a58, 0x6abe, 0x6b33, 0x6b5d, 0x6b71, 0x6bea, 0x6bee, 0x6bfe, 0x6c4d, 0x6c50, 0x6c53, 0x6c8f, 0x6cf6, 0x6cf9, 0x6d08, 0x6dc0, 0x6dc4, 0x6dd2:
			return true
		}
	}
	if cs == 0x48a1 && (ip == 0x6832 || ip == 0x6840 || ip == 0x6848 || ip == 0x691c || ip == 0x6930 || ip == 0x69a9 || ip == 0x69ad || ip == 0x69bd) {
		return true
	}
	return cs == 0x48a1 && (ip == 0x41af || ip == 0x4268 || ip == 0x68f4 || ip == 0x689e || ip == 0x6850 || ip == 0x2aa8 || ip == 0x4a67 || ip == 0x4b19 || ip == 0x4b6c || ip == 0x4b9a || ip == 0x4bc1 || ip == 0x4a83 || ip == 0x4afd || ip == 0x4b3a || mirrorPairBase(ip) != 0)
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
