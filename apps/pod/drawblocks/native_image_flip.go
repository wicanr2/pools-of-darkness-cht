package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

var flipImageCode = smallFingerprint(1111, "e95a2f6e11b374a9faa3b42a3936665692a05d289ce55e5718c2e58179c43048")

func flipAdmit(v CPUView) bool {
	if v.CS() != 0x48a1 || v.ES() != 0xa000 || v.Flags()&cpu.DF != 0 || v.Read8(0x449) != 0x13 {
		return false
	}
	var code [1111]byte
	copyView(v, 0x48a10+0x6a0c, code[:])
	return sha256.Sum256(code[:]) == flipImageCode.sum
}
func flipArgs(v CPUView) (uint32, uint16, bool) {
	if v.BP() > 65522 {
		return 0, 0, false
	}
	a := uint32(v.SS())*16 + uint32(v.BP())
	if a < 0x500 || a+14 > 0xa0000 {
		return 0, 0, false
	}
	return a, v.Read16(a + 12), true
}

func imageRestoredFlags(v CPUView) (uint16, bool) {
	reader, ok := v.(interface{ CPUModel() (cpu.Model, bool) })
	if !ok {
		return 0, false
	}
	model, known := reader.CPUModel()
	if !known {
		return 0, false
	}
	switch model {
	case cpu.Model80386:
		return v.Flags()&0x7fd5 | 2, true
	case cpu.Model8086, cpu.Model80186:
		return v.Flags()&0x0fd5 | 0xf002, true
	}
	return 0, false
}
func flipInput(v CPUView, n uint32) (uint32, uint32, bool) {
	base, si := uint32(v.DS())*16, uint32(v.SI())
	ok := si+n <= 65535 && base+si >= 0x500 && base+si+n <= 0xa0000 && !(base+si < 0x48a10+0x1252 && base+si+n > 0x48a10+0x11fb)
	return base, si, ok
}

func planFlipInitialize(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() != 0x6a0c || !flipAdmit(v) {
		return nil
	}
	base, si, ok := flipInput(v, 4)
	args, flags, argOK := flipArgs(v)
	if !ok || !argOK || flags&3 == 0 || args+8 < 0x48a10+0x1252 && args+14 > 0x48a10+0x11fb {
		return nil
	}
	width, height := v.Read16(base+si), v.Read16(base+si+2)
	x, y := int32(int16(v.Read16(args+8))), int32(int16(v.Read16(args+10)))
	if width == 0 || width > 320 || height == 0 || height > 200 || x < -320 || x > 639 || y < -200 || y > 399 {
		return nil
	}
	p := imageBasePlan(v)
	values := []uint16{width, height, uint16(y), uint16(y + int32(height) - 1), uint16(x), uint16(x + int32(width) - 1)}
	offsets := []uint16{0x11fb, 0x11ff, 0x1204, 0x1206, 0x1208, 0x120a}
	steps := []uint64{2, 4, 6, 9, 11, 14}
	ips := []uint16{0x6a0d, 0x6a12, 0x6a19, 0x6a23, 0x6a2a, 0x6a34}
	for i, n := range values {
		imageWord(p, offsets[i], n, steps[i], ips[i])
	}
	p.After.R[cpu.AX] = values[5]
	p.After.R[cpu.SI] += 4
	p.After.Flags = nativeCompareFlags(v.Flags(), uint16(x+int32(width)), 1) &^ cpu.CF
	if uint32(uint16(x))+uint32(width) > 65535 {
		p.After.Flags |= cpu.CF
	}
	return imageTiming(v, p, budget, 15, 69, 2, 0x6a38, 0x6a58)
}

func planFlipRow(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() != 0x6a58 || !flipAdmit(v) {
		return nil
	}
	base, si, ok := flipInput(v, 1)
	_, flags, argOK := flipArgs(v)
	if !ok || !argOK {
		return nil
	}
	raw := v.Read8(base + si)
	p := imageBasePlan(v)
	imageWord(p, 0x1250, 0, 1, 0x6a58)
	p.After.R[cpu.SI]++
	p.After.R[cpu.AX] = uint16(raw)
	if raw == 255 {
		p.After.Flags = nativeCompareFlags(v.Flags(), 255, 255)
		return imageTiming(v, p, budget, 6, 27, 1, 0x6a66, 0x6e5e)
	}
	y := int32(int16(v.Read16(0x48a10+0x1204))) + int32(raw)
	addN, addCost := uint64(0), uint64(0)
	if flags&2 != 0 {
		y = int32(int16(v.Read16(0x48a10+0x1206))) - int32(raw)
		addN, addCost = 3, 11
	}
	top, bottom := int32(v.Read16(0x48a10+0xec3)), int32(v.Read16(0x48a10+0xec7))
	if y < -32768 || y > 32767 || top > bottom || bottom >= 200 {
		return nil
	}
	p.After.R[cpu.AX] = uint16(y)
	if y > bottom {
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(y), uint16(bottom))
		return imageTiming(v, p, budget, 10+addN, 59+addCost, 1, 0x6a86, 0x6a3a)
	}
	if y < top {
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(y), uint16(top))
		return imageTiming(v, p, budget, 12+addN, 68+addCost, 1, 0x6a8d, 0x6a3a)
	}
	stride := uint32(v.Read16(0x48a10 + 0xedd))
	x := int32(v.Read16(0x48a10+0xec1)) - int32(v.Read16(0x48a10+0xed3))
	dy := y - int32(v.Read16(0x48a10+0xed5))
	if stride == 0 || stride > 320 || x < 0 || uint32(x) >= stride || dy < 0 || dy > 199 {
		return nil
	}
	row := uint32(x) + uint32(dy)*stride
	if row >= 64000 {
		return nil
	}
	n, cost := uint64(23)+addN, uint64(125)+addCost
	p.After.R[cpu.AX] = uint16(uint32(dy) * stride)
	p.After.R[cpu.DX] = 0
	p.After.Flags = imageRowAddFlags(v.Flags(), uint16(x), uint16(uint32(dy)*stride))
	if stride == 320 {
		n, cost = 25+addN, 115+addCost
		p.After.R[cpu.AX] = 320
		p.After.R[cpu.DX] = uint16(dy * 64)
		p.After.Flags = imageRowAddFlags(v.Flags(), uint16(x+dy*256), uint16(dy*64))
	}
	p.After.R[cpu.DI] = uint16(row)
	imageWord(p, 0x120c, uint16(row), n, 0x6ab9)
	return imageTiming(v, p, budget, n, cost, 1, 0x6ab9, 0x6abe)
}

func planFlipHeader(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() != 0x6abe || !flipAdmit(v) {
		return nil
	}
	p := imageBasePlan(v)
	marker := v.Read16(0x48a10 + 0x1250)
	if marker >= 0x8000 {
		p.After.Flags = nativeCompareFlags(v.Flags(), marker, 0x8000)
		return imageTiming(v, p, budget, 2, 14, 0, 0x6ac5, 0x6a58)
	}
	base, si, ok := flipInput(v, 4)
	_, flags, argOK := flipArgs(v)
	if !ok || !argOK {
		return nil
	}
	raw := v.Read16(base + si)
	x := int32(raw&0x7fff) + int32(int16(v.Read16(0x48a10+0x1208)))
	if flags&1 != 0 {
		x = int32(int16(v.Read16(0x48a10+0x120a))) - int32(raw&0x7fff)
	}
	if x < -320 || x > 639 {
		return nil
	}
	imageWord(p, 0x1250, raw, 4, 0x6ac8)
	p.After.R[cpu.AX] = uint16(x)
	p.After.R[cpu.SI] += 2
	if flags&1 != 0 {
		p.After.Flags = nativeCompareFlags(v.Flags(), 1, 0) &^ cpu.PF
		return imageTiming(v, p, budget, 14, 67, 1, 0x6aee, 0x6c53)
	}
	width, n := int32(v.Read8(base+si+2)), uint32(v.Read8(base+si+3))
	left, right, row := int32(v.Read16(0x48a10+0xec1)), int32(v.Read16(0x48a10+0xec5)), int32(v.Read16(0x48a10+0x120c))
	if width == 0 || n == 0 || left > right || right >= 320 || si+4+n > 65535 || base+si+4+n > 0xa0000 {
		return nil
	}
	endX := x + width - 1
	p.After.R[cpu.DI] = uint16(row + x - left)
	p.After.R[cpu.DX] = uint16(endX)
	if x > right {
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(x), uint16(right))
		return imageTiming(v, p, budget, 20, 98, 1, 0x6b0c, 0x6a48)
	}
	if endX < left {
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(endX), uint16(left))
		return imageTiming(v, p, budget, 22, 112, 1, 0x6b16, 0x6a50)
	}
	p.After.R[cpu.SI] += 2
	p.After.R[cpu.BX] = uint16(si + 4 + n)
	p.After.R[cpu.CX] &= 255
	if x < left {
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(x), uint16(left))
		return imageTiming(v, p, budget, 28, 138, 1, 0x6b2a, 0x6b5d)
	}
	p.After.Flags = nativeCompareFlags(v.Flags(), uint16(endX), uint16(right))
	if endX > right {
		return imageTiming(v, p, budget, 30, 147, 1, 0x6b31, 0x6b5d)
	}
	return imageTiming(v, p, budget, 30, 142, 1, 0x6b31, 0x6b33)
}

// Reversed literal spans preserve word byte order, with ABI metadata for each
// original unroll stripe. No pixel is generated by an instruction evaluator.
func mirrorLiteral(p *machine.HostWorkPlan, v CPUView, source uint32, di uint16, count uint32, wordBase uint16, prefix, noPairsPrefix uint64) bool {
	pairs := count / 2
	src := uint32(v.DS()) * 16
	if pairs > 0 {
		if v.SP() < 2 {
			return false
		}
		stack := uint32(v.SS())*16 + uint32(v.SP())
		if stack-2 < 0x500 || stack > 0xa0000 || stack-2 < 0x48a10+0x7580 && stack > 0x48a10 || stack-2 < src+uint32(v.BX()) && stack > src+uint32(v.SI()) {
			return false
		}
		saved := nativeCompareFlags(v.Flags(), uint16(pairs), 0)
		if count%2 != 0 {
			saved |= cpu.CF
		}
		step := prefix - 1
		p.Writes = append(p.Writes, machine.HostWrite{Address: stack - 2, Value: byte(saved), Step: step, CS: 0x48a1, IP: wordBase - 2}, machine.HostWrite{Address: stack - 1, Value: byte(saved >> 8), Step: step, CS: 0x48a1, IP: wordBase - 2})
		for i := uint32(0); i < pairs; i++ {
			a, b := v.Read8(src+source+2*i), v.Read8(src+source+2*i+1)
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + uint32(di) - 1 - 2*i, Data: []byte{b, a}, Count: 1, Width: 2, Step: prefix + uint64(i)*6 + uint64(i/4) + 3, CS: 0x48a1, IP: wordBase + uint16(i%4)*12 + 3})
			p.After.R[cpu.AX] = uint16(a)<<8 | uint16(b)
		}
	}
	if count%2 != 0 {
		step := noPairsPrefix + 2
		if pairs > 0 {
			step = prefix + uint64(pairs)*6 + uint64((pairs-1)/4) + 5
		}
		value := v.Read8(src + source + count - 1)
		p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + uint32(di) - pairs*2, Value: uint16(value), Count: 1, Width: 1, Step: step, CS: 0x48a1, IP: wordBase + 55})
		p.After.R[cpu.AX] = p.After.R[cpu.AX]&0xff00 | uint16(value)
	}
	return true
}

func planMirrorPacket(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() != 0x6c8f || !flipAdmit(v) || v.CX()&0xff00 != 0 {
		return nil
	}
	base, si, end, ok := imageClipSource(v)
	if !ok {
		return nil
	}
	ctrl := v.Read8(base + si)
	count := uint32(ctrl>>1) + 1
	repeat := ctrl&1 != 0
	used := count + 1
	if repeat {
		used = 2
	}
	if si+used > end || uint32(v.DI()) >= 64000 || count > uint32(v.DI())%320+1 {
		return nil
	}
	p := imageBasePlan(v)
	p.After.R[cpu.CX] = 0
	p.After.R[cpu.SI] = uint16(si + used)
	p.After.R[cpu.DI] -= uint16(count)
	last := si+used == end
	steps, cycles, loads := uint64(0), uint64(0), uint64(1)
	pairs, odd := count/2, count%2
	if repeat {
		color := v.Read8(base + si + 1)
		value := uint16(color) * 257
		p.After.R[cpu.AX] = value
		steps, cycles, loads = 17, 56, 2
		if pairs > 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + uint32(v.DI()) - 1, Value: value, Count: uint16(pairs), Width: 2, Backward: true, Step: 11, CS: 0x48a1, IP: 0x6ce7})
		}
		if odd > 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + uint32(v.DI()) - pairs*2, Value: value, Count: 1, Width: 1, Step: 14, CS: 0x48a1, IP: 0x6cec})
		}
		cycles += 5 * uint64(pairs+odd)
	} else {
		p.After.R[cpu.AX] = v.AX()&0xff00 | uint16(ctrl>>1)
		if !mirrorLiteral(p, v, si+1, v.DI(), count, 0x6c9e, 9, 7) {
			return nil
		}
		if pairs == 0 {
			steps, cycles = 12, 49
		} else {
			wrap := (pairs - 1) / 4
			steps = 14 + uint64(pairs)*6 + uint64(wrap) + uint64(odd)*3
			cycles = 52 + uint64(pairs)*22 + uint64(wrap)*7
			if odd > 0 {
				cycles += 11
			}
		}
		loads = 1 + uint64(pairs+odd)
	}
	finalFlags := v.Flags()
	if !repeat && pairs > 0 {
		var ok bool
		finalFlags, ok = imageRestoredFlags(v)
		if !ok {
			return nil
		}
	}
	p.After.Flags = nativeCompareFlags(finalFlags, p.After.R[cpu.SI], v.BX())
	next, lastIP := uint16(0x6c8f), uint16(0x6cdb)
	if repeat {
		lastIP = 0x6cf1
	}
	if last {
		steps++
		cycles += 2
		next, lastIP = 0x6abe, 0x6cdd
		if repeat {
			lastIP = 0x6cf3
		}
	}
	if v.DOSBoxCost() {
		cycles = steps + loads
		if repeat {
			cycles += uint64(pairs + odd)
		}
	}
	if steps > budget {
		return nil
	}
	p.Instructions, p.Cycles, p.LastIP, p.After.IP = steps, cycles, lastIP, next
	return p
}

func planFlipControl(v CPUView, budget uint64) *machine.HostWorkPlan {
	ip := v.IP()
	if ip != 0x6a3a && ip != 0x6a48 && ip != 0x6a50 && ip != 0x6c4d && ip != 0x6c50 {
		return nil
	}
	if !flipAdmit(v) {
		return nil
	}
	p := imageBasePlan(v)
	if ip == 0x6c4d {
		return imageTiming(v, p, budget, 1, 7, 0, ip, 0x6a48)
	}
	if ip == 0x6c50 {
		return imageTiming(v, p, budget, 1, 7, 0, ip, 0x6a50)
	}
	marker := v.Read16(0x48a10 + 0x1250)
	if ip == 0x6a3a && marker >= 0x8000 {
		p.After.Flags = nativeCompareFlags(v.Flags(), marker, 0x8000)
		return imageTiming(v, p, budget, 2, 14, 0, 0x6a41, 0x6a58)
	}
	base, si, ok := flipInput(v, 2)
	if !ok {
		return nil
	}
	if ip != 0x6a3a {
		n := uint32(v.Read8(base + si + 1))
		if si+2+n > 65535 || base+si+2+n > 0xa0000 {
			return nil
		}
		p.After.R[cpu.AX] = uint16(n)
		p.After.R[cpu.SI] = uint16(si + 2 + n)
		p.After.Flags = imageRowAddFlags(v.Flags(), uint16(si+2), uint16(n))
		next := uint16(0x6a3a)
		if ip == 0x6a50 {
			next = 0x6abe
		}
		return imageTiming(v, p, budget, 5, 20, 1, ip+6, next)
	}
	base, si, ok = flipInput(v, 4)
	if !ok {
		return nil
	}
	raw := v.Read16(base + si)
	n := uint32(v.Read8(base + si + 3))
	if si+4+n > 65535 || base+si+4+n > 0xa0000 {
		return nil
	}
	imageWord(p, 0x1250, raw, 4, 0x6a44)
	p.After.R[cpu.AX] = uint16(n)
	p.After.R[cpu.SI] = uint16(si + 4 + n)
	p.After.Flags = imageRowAddFlags(v.Flags(), uint16(si+4), uint16(n))
	return imageTiming(v, p, budget, 9, 39, 2, 0x6a4e, 0x6a3a)
}

func planMirrorSpan(v CPUView, budget uint64) *machine.HostWorkPlan {
	if v.IP() != 0x6c53 || !flipAdmit(v) {
		return nil
	}
	base, si, ok := flipInput(v, 2)
	if !ok {
		return nil
	}
	width, n := int32(v.Read8(base+si)), uint32(v.Read8(base+si+1))
	x := int32(int16(v.AX()))
	left, right, row := int32(v.Read16(0x48a10+0xec1)), int32(v.Read16(0x48a10+0xec5)), int32(v.Read16(0x48a10+0x120c))
	if x < -320 || x > 639 || width == 0 || n == 0 || left > right || right >= 320 || si+2+n > 65535 || base+si+2+n > 0xa0000 {
		return nil
	}
	p := imageBasePlan(v)
	endX := x - width + 1
	p.After.R[cpu.DI] = uint16(row + x - left)
	p.After.R[cpu.DX] = uint16(endX)
	if x < left {
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(x), uint16(left))
		return imageTiming(v, p, budget, 9, 42, 0, 0x6c6c, 0x6c4d)
	}
	if endX > right {
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(endX), uint16(right))
		return imageTiming(v, p, budget, 11, 51, 0, 0x6c73, 0x6c50)
	}
	visibleX, visibleEnd := x, endX
	if visibleX > right {
		visibleX = right
	}
	if visibleEnd < left {
		visibleEnd = left
	}
	at, count := row+visibleX-left, visibleX-visibleEnd+1
	if at < 0 || at >= 64000 || count < 1 || at%320+1 < count {
		return nil
	}
	p.After.R[cpu.SI] += 2
	p.After.R[cpu.BX] = uint16(si + 2 + n)
	p.After.R[cpu.CX] &= 255
	if x > right {
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(x), uint16(right))
		return imageTiming(v, p, budget, 18, 74, 0, 0x6c86, 0x6cf9)
	}
	p.After.Flags = nativeCompareFlags(v.Flags(), uint16(endX), uint16(left))
	if endX < left {
		return imageTiming(v, p, budget, 20, 83, 0, 0x6c8d, 0x6cf9)
	}
	return imageTiming(v, p, budget, 20, 78, 0, 0x6c8d, 0x6c8f)
}

func planMirrorClipSetup(v CPUView, budget uint64) *machine.HostWorkPlan {
	ip := v.IP()
	if ip != 0x6cf9 && ip != 0x6cf6 && ip != 0x6dc0 && ip != 0x6dc4 && ip != 0x6dd2 {
		return nil
	}
	if !flipAdmit(v) {
		return nil
	}
	_, _, _, ok := imageClipSource(v)
	if !ok {
		return nil
	}
	p := imageBasePlan(v)
	left, right := int32(v.Read16(0x48a10+0xec1)), int32(v.Read16(0x48a10+0xec5))
	if left > right || right >= 320 {
		return nil
	}
	if ip == 0x6dd2 {
		if v.DX() == 0 {
			return nil
		}
		p.After.R[cpu.SI] = v.BX()
		p.After.Flags = nativeCompareFlags(v.Flags(), v.DX(), 0)
		return imageTiming(v, p, budget, 4, 18, 0, 0x6e37, 0x6a3a)
	}
	x := int32(int16(v.AX()))
	prefix, extra := uint64(0), uint64(0)
	if ip == 0x6cf9 {
		jl := (v.Flags()&cpu.SF != 0) != (v.Flags()&cpu.OF != 0)
		if !jl {
			skip := x - right
			if skip < 1 || skip > 320 {
				return nil
			}
			imageWord(p, 0x1232, uint16(x), 2, 0x6cfb)
			imageWord(p, 0x1230, uint16(skip), 4, 0x6d04)
			p.After.R[cpu.AX] = uint16(skip)
			p.After.Flags = nativeCompareFlags(v.Flags(), uint16(x), uint16(right))
			return imageTiming(v, p, budget, 4, 15, 0, 0x6d04, 0x6d08)
		}
		prefix, extra = 2, 14
	} else if ip == 0x6dc0 {
		x = int32(int16(v.Read16(0x48a10 + 0x1232)))
		prefix, extra = 1, 3
	} else if ip == 0x6cf6 {
		prefix, extra = 1, 7
	}
	n := x - left
	if x < -320 || x > 639 {
		return nil
	}
	if n < 0 {
		p.After.R[cpu.AX] = uint16(n)
		p.After.R[cpu.SI] = v.BX()
		p.After.Flags = nativeCompareFlags(v.Flags(), uint16(x), uint16(left))
		return imageTiming(v, p, budget, prefix+4, extra+23, 0, 0x6e37, 0x6a3a)
	}
	if n >= 320 {
		return nil
	}
	n++
	p.After.R[cpu.AX] = uint16(n)
	p.After.R[cpu.DX] = 0
	p.After.Flags = nativeCompareFlags(v.Flags(), 0, 0)
	imageWord(p, 0x1230, uint16(n), prefix+4, 0x6dcc)
	return imageTiming(v, p, budget, prefix+5, extra+16, 0, 0x6dd0, 0x6dd2)
}

func planMirrorClippedPacket(v CPUView, budget uint64) *machine.HostWorkPlan {
	leading := v.IP() == 0x6d08
	if !leading && v.IP() != 0x6dd2 {
		return nil
	}
	if !flipAdmit(v) || v.CX()&0xff00 != 0 || !leading && v.DX() != 0 {
		return nil
	}
	base, si, end, ok := imageClipSource(v)
	if !ok {
		return nil
	}
	ctrl := v.Read8(base + si)
	run := uint32(ctrl>>1) + 1
	repeat := ctrl&1 != 0
	used := run + 1
	if repeat {
		used = 2
	}
	if si+used > end {
		return nil
	}
	remain := int32(int16(v.Read16(0x48a10 + 0x1230)))
	if remain < 0 || remain > 320 {
		return nil
	}
	p := imageBasePlan(v)
	p.After.R[cpu.AX] = uint16(remain)
	skip, visible := uint32(0), run
	over := int32(0)
	source := si + 1
	wordBase, prefix, noPairPrefix := uint16(0), uint64(0), uint64(0)
	wordStep, byteStep := uint64(0), uint64(0)
	wordIP, byteIP := uint16(0), uint16(0)
	steps, cycles, loads := uint64(0), uint64(0), uint64(1)
	lastIP, next := uint16(0), uint16(0)
	if leading {
		pos := int32(int16(v.Read16(0x48a10 + 0x1232)))
		left := int32(v.Read16(0x48a10 + 0xec1))
		if pos < -320 || pos > 639 || left >= 320 {
			return nil
		}
		pos -= int32(run)
		posIP, remIP := uint16(0x6d11), uint16(0x6d1a)
		if repeat {
			posIP, remIP = 0x6d84, 0x6d8d
		}
		imageWord(p, 0x1232, uint16(pos), 6, posIP)
		imageWord(p, 0x1230, uint16(remain-int32(run)), 8, remIP)
		if remain >= int32(run) {
			p.After.R[cpu.CX] = uint16(run)
			p.After.R[cpu.SI] = uint16(si + used)
			p.After.R[cpu.DI] -= uint16(run)
			p.After.Flags = nativeCompareFlags(v.Flags(), v.DI(), uint16(run))
			cycles = 45
			lastIP = 0x6d25
			if repeat {
				cycles = 50
				lastIP = 0x6d97
			}
			return imageTiming(v, p, budget, 12, cycles, 1, lastIP, 0x6d08)
		}
		skip = uint32(remain)
		over = left - pos - 1
		if over < 0 {
			over = 0
		}
		if remain+over >= int32(run) {
			return nil
		}
		visible = run - skip - uint32(over)
		source += skip
		p.After.R[cpu.DX] = uint16(left - pos - 1)
		p.After.R[cpu.DI] -= uint16(skip)
		wordBase, prefix, noPairPrefix = 0x6d42, 20, 18
		cycles = 0
		p.After.R[cpu.SI] = uint16(source + visible)
		lastIP, next = 0x6d7f, 0x6dc0
		if repeat {
			wordStep, byteStep, wordIP, byteIP = 21, 24, 0x6db4, 0x6db9
			steps, cycles, loads = 27, 102, 2
			p.After.R[cpu.SI] = uint16(si + 2)
			lastIP = 0x6dbe
		}
		if over > 0 {
			prefix++
			noPairPrefix++
			if repeat {
				steps++
				wordStep++
				byteStep++
				cycles -= 3
			}
		}
	} else {
		cut := uint32(remain) <= run
		if cut {
			visible = uint32(remain)
			over = 1
			p.After.R[cpu.DX] = 1
		}
		// Original 6DEF jumps directly to the literal byte path when SHR
		// yields zero, including a zero remaining count.
		if !repeat && visible == 0 {
			visible = 1
		}
		remIP := uint16(0x6de3)
		if repeat {
			remIP = 0x6e3e
		}
		imageWord(p, 0x1230, uint16(remain-int32(run)), 9, remIP)
		wordBase, prefix, noPairPrefix = 0x6df3, 14, 12
		p.After.R[cpu.SI] = uint16(source + visible)
		lastIP, next = 0x6e30, 0x6dd2
		if cut {
			prefix += 2
			noPairPrefix += 2
		}
		if repeat {
			steps, cycles, loads = 23, 79, 2
			wordStep, byteStep, wordIP, byteIP = 16, 19, 0x6e4f, 0x6e54
			p.After.R[cpu.SI] = uint16(si + 2)
			lastIP = 0x6e5b
			if cut {
				steps += 2
				wordStep += 2
				byteStep += 2
				cycles--
			}
		}
	}
	at := uint32(p.After.R[cpu.DI])
	if at >= 64000 || visible > at%320+1 {
		return nil
	}
	pairs, odd := visible/2, visible%2
	last := uint32(p.After.R[cpu.SI]) == end
	if repeat {
		value := uint16(v.Read8(base+si+1)) * 257
		p.After.R[cpu.AX] = value
		if pairs > 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at - 1, Value: value, Count: uint16(pairs), Width: 2, Backward: true, Step: wordStep, CS: 0x48a1, IP: wordIP})
		}
		if odd > 0 {
			p.Transfers = append(p.Transfers, machine.HostTransfer{Address: 0xa0000 + at - pairs*2, Value: value, Count: 1, Width: 1, Step: byteStep, CS: 0x48a1, IP: byteIP})
		}
		cycles += 5 * uint64(pairs+odd)
		if leading && last {
			steps++
			cycles += 12
			lastIP, next = 0x6e32, 0x6abe
		} else if !leading && last {
			cycles += 5
			lastIP, next = 0x6e32, 0x6abe
		}
	} else {
		if !mirrorLiteral(p, v, source, uint16(at), visible, wordBase, prefix, noPairPrefix) {
			return nil
		}
		loads += uint64(pairs + odd)
		if leading {
			if pairs == 0 {
				steps, cycles = 23, 102
			} else {
				wrap := (pairs - 1) / 4
				steps = 25 + uint64(pairs)*6 + uint64(wrap) + uint64(odd)*3
				cycles = 105 + 22*uint64(pairs) + 7*uint64(wrap)
				if odd > 0 {
					cycles += 11
				}
			}
			if over > 0 {
				steps++
				cycles -= 3
			}
		} else {
			if pairs == 0 {
				steps, cycles = 17, 70
			} else {
				wrap := (pairs - 1) / 4
				steps = 19 + uint64(pairs)*6 + uint64(wrap) + uint64(odd)*3
				cycles = 73 + 22*uint64(pairs) + 7*uint64(wrap)
				if odd > 0 {
					cycles += 11
				}
			}
			if over > 0 {
				steps += 2
				cycles--
			}
		}
		if last {
			steps++
			cycles += 2
			lastIP, next = 0x6e32, 0x6abe
			if leading {
				lastIP = 0x6d81
			}
		}
	}
	p.After.R[cpu.CX] = 0
	p.After.R[cpu.DI] -= uint16(visible)
	flags := v.Flags()
	if !repeat && pairs > 0 {
		var ok bool
		flags, ok = imageRestoredFlags(v)
		if !ok {
			return nil
		}
	}
	p.After.Flags = nativeCompareFlags(flags, p.After.R[cpu.SI], v.BX())
	if v.DOSBoxCost() {
		cycles = steps + loads
		if repeat {
			cycles += uint64(pairs + odd)
		}
	}
	if steps > budget {
		return nil
	}
	p.Instructions, p.Cycles, p.LastIP, p.After.IP = steps, cycles, lastIP, next
	return p
}
