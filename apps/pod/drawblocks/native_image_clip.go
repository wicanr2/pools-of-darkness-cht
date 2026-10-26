package drawblocks

// Spec025/#49: Go packet clipping and row scanning; ABI effects are separate.

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var imageSkipCode = smallFingerprint(30, "a8b2dbb0dbf04e34cca2aca89962a68329389b7bdfa17bf485b5feba3fec1383")
var imageClipCode = smallFingerprint(240, "e6e2c71563b3b8b991f450d0c1184b02acfd37ec8245bff32009f7d03ac1f23d")
var flippedClipCode = smallFingerprint(240, "f4d4dd51e317b3264eb6e20050901bd97adb082b87e34a69befb8cb5db549005")

type clipProgram struct {
	delta, header, skip uint16
	code                smallCode
}

var normalClipProgram = clipProgram{header: 0x689e, skip: 0x6832, code: imageClipCode}
var flippedClipProgram = clipProgram{delta: 0x241, header: 0x6abe, skip: 0x6a3a, code: flippedClipCode}

func remapClipPlan(p *machine.HostWorkPlan, program clipProgram) {
	if p == nil || program.delta == 0 {
		return
	}
	p.LastIP += program.delta
	switch p.After.IP {
	case 0x689e:
		p.After.IP = program.header
	case 0x6832:
		p.After.IP = program.skip
	default:
		p.After.IP += program.delta
	}
	for i := range p.Writes {
		p.Writes[i].IP += program.delta
	}
	for i := range p.Transfers {
		p.Transfers[i].IP += program.delta
	}
}

func imageBasePlan(v CPUView) *machine.HostWorkPlan {
	return &machine.HostWorkPlan{Algorithmic: true, After: machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: v.IP(), Flags: v.Flags()}, LastCS: 0x48a1}
}
func imageWord(p *machine.HostWorkPlan, off, value uint16, step uint64, ip uint16) {
	p.Writes = append(p.Writes, machine.HostWrite{Address: 0x48a10 + uint32(off), Value: byte(value), Step: step, CS: 0x48a1, IP: ip}, machine.HostWrite{Address: 0x48a10 + uint32(off) + 1, Value: byte(value >> 8), Step: step, CS: 0x48a1, IP: ip})
}
func imageAdmit(v CPUView, lo int, fp smallCode) bool {
	if v.CS() != 0x48a1 || v.ES() != 0xa000 || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 {
		return false
	}
	var buf [240]byte
	copyView(v, 0x48a10+uint32(lo), buf[:fp.size])
	return sha256.Sum256(buf[:fp.size]) == fp.sum
}
func imageTiming(v CPUView, p *machine.HostWorkPlan, budget, steps, cycles, loads uint64, last, next uint16) *machine.HostWorkPlan {
	if steps > budget {
		return nil
	}
	p.Instructions, p.Cycles, p.LastIP, p.After.IP = steps, cycles, last, next
	if v.DOSBoxCost() {
		p.Cycles = steps + loads
	}
	return p
}

// Parse whole excluded spans by their asset lengths, without interpreting code.
func planImageControl(v CPUView, budget uint64) *machine.HostWorkPlan {
	ip := v.IP()
	if ip != 0x6832 && ip != 0x6840 && ip != 0x6848 && ip != 0x689e {
		return nil
	}
	if !imageAdmit(v, 0x6832, imageSkipCode) {
		return nil
	}
	p := imageBasePlan(v)
	marker := v.Read16(0x48a10 + 0x1250)
	if (ip == 0x6832 || ip == 0x689e) && marker >= 0x8000 {
		if ip == 0x689e && !imageAdmit(v, 0x689e, imageSpanCode) {
			return nil
		}
		p.After.Flags = nativeCompareFlags(v.Flags(), marker, 0x8000)
		last := uint16(0x6839)
		if ip == 0x689e {
			last = 0x68a5
		}
		return imageTiming(v, p, budget, 2, 14, 0, last, 0x6850)
	}
	src, si := uint32(v.DS())*16, uint32(v.SI())
	if ip == 0x6840 || ip == 0x6848 {
		if si+2 > 65536 || src+si < 0x500 || src+si+2 > 0xa0000 {
			return nil
		}
		n := uint32(v.Read8(src + si + 1))
		if si+2+n > 65535 || src+si+2+n > 0xa0000 {
			return nil
		}
		p.After.R[cpu.AX] = uint16(n)
		p.After.R[cpu.SI] = uint16(si + 2 + n)
		p.After.Flags = imageRowAddFlags(v.Flags(), uint16(si+2), uint16(n))
		next := uint16(0x6832)
		if ip == 0x6848 {
			next = 0x689e
		}
		return imageTiming(v, p, budget, 5, 20, 1, ip+6, next)
	}
	if ip == 0x6832 {
		if si+4 > 65536 || src+si < 0x500 || src+si+4 > 0xa0000 {
			return nil
		}
		raw := v.Read16(src + si)
		n := uint32(v.Read8(src + si + 3))
		if si+4+n > 65535 || src+si+4+n > 0xa0000 || src+si < 0x48a10+0x1252 && src+si+4+n > 0x48a10+0x1250 {
			return nil
		}
		imageWord(p, 0x1250, raw, 4, 0x683c)
		p.After.R[cpu.AX] = uint16(n)
		p.After.R[cpu.SI] = uint16(si + 4 + n)
		p.After.Flags = imageRowAddFlags(v.Flags(), uint16(si+4), uint16(n))
		return imageTiming(v, p, budget, 9, 39, 2, 0x6846, 0x6832)
	}
	if !imageAdmit(v, 0x689e, imageSpanCode) || v.BP() > 65525 || si+4 > 65536 {
		return nil
	}
	arg := uint32(v.SS())*16 + uint32(v.BP()) + 8
	if src+si < 0x500 || src+si+4 > 0xa0000 || arg < 0x500 || arg+2 > 0xa0000 {
		return nil
	}
	raw := v.Read16(src + si)
	x := int32(raw&0x7fff) + int32(int16(v.Read16(arg)))
	width, n := int32(v.Read8(src+si+2)), uint32(v.Read8(src+si+3))
	left, right := int32(v.Read16(0x48a10+0xec1)), int32(v.Read16(0x48a10+0xec5))
	row := int32(v.Read16(0x48a10 + 0x120c))
	if x < -255 || x > 574 || width == 0 || n == 0 || left > right || right >= 320 || si+4+n > 65535 || src+si+4+n > 0xa0000 || src+si < 0x48a10+0x1252 && src+si+4+n > 0x48a10+0x1250 || arg < 0x48a10+0x1252 && arg+2 > 0x48a10+0x1250 {
		return nil
	}
	endX := x + width - 1
	di := row + x - left
	visibleX, visibleEnd := x, endX
	if visibleX < left {
		visibleX = left
	}
	if visibleEnd > right {
		visibleEnd = right
	}
	if visibleX <= visibleEnd {
		at := row + visibleX - left
		count := visibleEnd - visibleX + 1
		if at < 0 || at+count > 64000 || at%320+count > 320 {
			return nil
		}
	}
	imageWord(p, 0x1250, raw, 4, 0x68a8)
	p.After.R[cpu.AX] = uint16(x)
	p.After.R[cpu.DX] = uint16(endX)
	p.After.R[cpu.DI] = uint16(di)
	p.After.R[cpu.SI] += 2
	if x > right {
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(x), uint16(right))
		return imageTiming(v, p, budget, 16, 71, 1, 0x68cd, 0x6840)
	}
	if endX < left {
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(endX), uint16(left))
		return imageTiming(v, p, budget, 18, 85, 1, 0x68d7, 0x6848)
	}
	p.After.R[cpu.SI] += 2
	p.After.R[cpu.BX] = uint16(si + 4 + n)
	p.After.R[cpu.CX] &= 255
	if x < left {
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(x), uint16(left))
		return imageTiming(v, p, budget, 24, 111, 1, 0x68eb, 0x691c)
	}
	if endX > right {
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(endX), uint16(right))
		return imageTiming(v, p, budget, 26, 120, 1, 0x68f2, 0x691c)
	}
	return nil
}

func imageClipSource(v CPUView) (uint32, uint32, uint32, bool) {
	src, si, end := uint32(v.DS())*16, uint32(v.SI()), uint32(v.BX())
	ok := si < end && src+si >= 0x500 && src+end <= 0xa0000 && !(src+si < 0x48a10+0x1234 && src+end > 0x48a10+0x1230)
	return src, si, end, ok
}

// Clipping setup derives a visible count and left skip from image geometry.
func planClipSetup(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() >= 0x6b5d {
		return planClipSetupProgram(v, budget, flippedClipProgram)
	}
	return planClipSetupProgram(v, budget, normalClipProgram)
}
func planClipSetupProgram(v CPUView, budget uint64, program clipProgram) (result *machine.HostWorkPlan) {
	defer func() { remapClipPlan(result, program) }()
	ip := v.IP() - program.delta
	if ip != 0x691c && ip != 0x69a9 && ip != 0x69ad && ip != 0x69bd {
		return nil
	}
	if !imageAdmit(v, 0x691c+int(program.delta), program.code) {
		return nil
	}
	_, _, _, ok := imageClipSource(v)
	if !ok {
		return nil
	}
	p := imageBasePlan(v)
	left, right := int32(v.Read16(0x48a10+0xec1)), int32(v.Read16(0x48a10+0xec5))
	x := int32(int16(v.AX()))
	prefix, extra := uint64(0), uint64(0)
	if left > right || right >= 320 {
		return nil
	}
	if ip == 0x69bd {
		if v.DX() == 0 {
			return nil
		}
		p.After.R[cpu.SI] = v.BX()
		p.After.Flags = nativeCompareFlags(v.Flags(), v.DX(), 0)
		return imageTiming(v, p, budget, 4, 18, 0, 0x69e9, 0x6832)
	}
	if ip == 0x691c {
		jle := v.Flags()&cpu.ZF != 0 || (v.Flags()&cpu.SF != 0) != (v.Flags()&cpu.OF != 0)
		if jle {
			skip := left - x
			if skip < 1 || skip > 255 {
				return nil
			}
			imageWord(p, 0x1232, uint16(x), 2, 0x6921)
			imageWord(p, 0x1230, uint16(skip), 5, 0x692c)
			p.After.R[cpu.AX] = uint16(skip)
			p.After.Flags = nativeCompareFlags(v.Flags(), 0, uint16(x-left))
			return imageTiming(v, p, budget, 5, 22, 0, 0x692c, 0x6930)
		}
		prefix, extra = 2, 9
	} else if ip == 0x69a9 {
		x = int32(int16(v.Read16(0x48a10 + 0x1232)))
		prefix, extra = 1, 3
	}
	n := right - x + 1
	if x < -255 || x > 574 {
		return nil
	}
	p.After.R[cpu.AX] = uint16(n)
	p.After.Flags = imageRowAddFlags(v.Flags(), uint16(n-1), 1) &^ cpu.CF
	if x != right {
		p.After.Flags |= cpu.CF
	}
	if n <= 0 {
		p.After.R[cpu.SI] = v.BX()
		return imageTiming(v, p, budget, prefix+6, extra+27, 0, 0x69e9, 0x6832)
	}
	if n > 320 {
		return nil
	}
	imageWord(p, 0x1230, uint16(n), prefix+5, 0x69b7)
	p.After.R[cpu.DX] = 0
	p.After.Flags = nativeCompareFlags(v.Flags(), 0, 0)
	return imageTiming(v, p, budget, prefix+6, extra+18, 0, 0x69bb, 0x69bd)
}

func imageRunTransfer(p *machine.HostWorkPlan, di uint16, count uint32, color byte, step uint64, wordIP, byteIP uint16) {
	value := uint16(color) * 257
	if count/2 > 0 {
		p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + uint32(di), Value: value, Count: uint16(count / 2), Width: 2, Step: step, CS: 0x48a1, IP: wordIP})
	}
	if count%2 > 0 {
		p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + uint32(di) + count/2*2, Value: value, Count: 1, Width: 1, Step: step + 2, CS: 0x48a1, IP: byteIP})
	}
}

// Decode one asset packet and clip its visible interval using Go arithmetic.
func planClippedRun(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() >= 0x6b5d {
		return planClippedRunProgram(v, budget, flippedClipProgram)
	}
	return planClippedRunProgram(v, budget, normalClipProgram)
}
func planClippedRunProgram(v CPUView, budget uint64, program clipProgram) (result *machine.HostWorkPlan) {
	defer func() { remapClipPlan(result, program) }()
	ip := v.IP() - program.delta
	left := ip == 0x6930
	if !left && ip != 0x69bd {
		return nil
	}
	if !imageAdmit(v, 0x691c+int(program.delta), program.code) || v.CX()&0xff00 != 0 {
		return nil
	}
	src, si, end, ok := imageClipSource(v)
	if !ok || !left && v.DX() != 0 {
		return nil
	}
	ctrl := v.Read8(src + si)
	run := uint32(ctrl>>1) + 1
	repeat := ctrl&1 != 0
	used := run + 1
	if repeat {
		used = 2
	}
	if si+used > end {
		return nil
	}
	remaining := int32(int16(v.Read16(0x48a10 + 0x1230)))
	if remaining < 0 || remaining > 320 {
		return nil
	}
	p := imageBasePlan(v)
	p.After.R[cpu.AX] = uint16(remaining)
	loads := uint64(1)
	if repeat {
		loads++
	}
	cycles, steps, wordStep := uint64(0), uint64(0), uint64(0)
	wordIP, byteIP := uint16(0), uint16(0)
	skip, visible := uint32(0), run
	over := int32(0)
	last, nextIP, lastIP := false, uint16(0), uint16(0)
	source := si + 1
	if left {
		pos := int32(int16(v.Read16(0x48a10 + 0x1232)))
		right := int32(v.Read16(0x48a10 + 0xec5))
		if right >= 320 || pos < -255 || pos > 574 {
			return nil
		}
		pos += int32(run)
		imageWord(p, 0x1232, uint16(pos), 6, 0x6939)
		if repeat {
			p.Writes = nil
			imageWord(p, 0x1232, uint16(pos), 6, 0x6972)
		}
		imageWord(p, 0x1230, uint16(remaining-int32(run)), 8, 0x6942)
		if repeat {
			p.Writes = p.Writes[:2]
			imageWord(p, 0x1230, uint16(remaining-int32(run)), 8, 0x697b)
		}
		if remaining >= int32(run) {
			p.After.R[cpu.CX] = uint16(run)
			p.After.R[cpu.SI] = uint16(si + used)
			p.After.R[cpu.DI] += uint16(run)
			p.After.Flags = imageRowAddFlags(v.Flags(), v.DI(), uint16(run))
			cycles = 45
			if repeat {
				cycles += 5
			}
			return imageTiming(v, p, budget, 12, cycles, 1, 0x694d+boolOffset(repeat, 0x38), 0x6930)
		}
		skip = uint32(remaining)
		over = pos - 1 - right
		if over < 0 {
			over = 0
		}
		if over+remaining >= int32(run) {
			return nil
		}
		visible = run - skip - uint32(over)
		source += skip
		p.After.R[cpu.DX] = uint16(pos - 1 - right)
		p.After.R[cpu.DI] += uint16(skip)
		p.After.R[cpu.SI] = uint16(source + visible)
		steps = 23
		wordStep = 18
		wordIP, byteIP = 0x6966, 0x696a
		cycles = 89
		if repeat {
			p.After.R[cpu.SI] = uint16(si + 2)
			steps = 23
			wordStep = 19
			wordIP, byteIP = 0x699f, 0x69a3
			cycles = 94
		}
		if over > 0 {
			steps++
			wordStep++
			cycles -= 3
		}
		last = uint32(p.After.R[cpu.SI]) == end
		nextIP, lastIP = 0x69a9, 0x6970
		if repeat {
			lastIP = 0x69a7
			if last {
				steps++
				cycles += 12
			}
		}
		if last {
			nextIP, lastIP = 0x689e, 0x69e4
			if !repeat {
				cycles += 5
			}
		}
	} else {
		visible = run
		over = 0
		if uint32(remaining) < run {
			visible = uint32(remaining)
			over = 1
			p.After.R[cpu.DX] = 1
		}
		imageWord(p, 0x1230, uint16(remaining-int32(run)), 9, 0x69ce)
		if repeat {
			p.Writes = nil
			imageWord(p, 0x1230, uint16(remaining-int32(run)), 9, 0x69f0)
		}
		p.After.R[cpu.SI] = uint16(source + visible)
		steps = 16
		wordStep = 12
		wordIP, byteIP = 0x69da, 0x69de
		cycles = 55
		if repeat {
			p.After.R[cpu.SI] = uint16(si + 2)
			steps = 18
			wordStep = 14
			wordIP, byteIP = 0x69ff, 0x6a03
			cycles = 69
		}
		if over > 0 {
			steps += 2
			wordStep += 2
			cycles--
		}
		last = uint32(p.After.R[cpu.SI]) == end
		nextIP, lastIP = 0x69bd, 0x69e2
		if repeat {
			lastIP = 0x6a07
		}
		if last {
			steps++
			cycles += 2
			nextIP, lastIP = 0x689e, 0x69e4
			if repeat {
				lastIP = 0x6a09
			}
		}
	}
	at := uint32(p.After.R[cpu.DI])
	if at+visible > 64000 || at%320+visible > 320 {
		return nil
	}
	iterations := uint64(visible/2 + visible%2)
	p.After.R[cpu.CX] = 0
	p.After.R[cpu.DI] += uint16(visible)
	p.After.Flags = nativeCompareFlags(v.Flags(), p.After.R[cpu.SI], v.BX())
	color := byte(0)
	if repeat {
		color = v.Read8(src + si + 1)
		p.After.R[cpu.AX] = uint16(color) * 257
	}
	imageRunTransfer(p, uint16(at), visible, color, wordStep, wordIP, byteIP)
	if !repeat {
		for i := range p.Transfers {
			s := &p.Transfers[i]
			offset := source + uint32(s.Address-(0xa0000+at))
			s.Data = make([]byte, int(s.Count)*int(s.Width))
			copyView(v, src+offset, s.Data)
		}
	}
	cycles += 5 * iterations
	if v.DOSBoxCost() {
		cycles = steps + loads + iterations
	}
	if steps > budget {
		return nil
	}
	p.Instructions, p.Cycles, p.LastIP, p.After.IP = steps, cycles, lastIP, nextIP
	return p
}
func boolOffset(b bool, n uint16) uint16 {
	if b {
		return n
	}
	return 0
}
