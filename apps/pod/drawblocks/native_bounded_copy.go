// Spec025/#49: budgeted Go memory/image spans.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

// Decode one immutable far-call frame into an ordered byte-copy job. The
// checkpoint projection preserves the guest ABI at a scheduler boundary.
func planBoundedCopy(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x1ed4 || v.Read8(0x449) != 0x13 {
		return nil
	}
	valid := false
	for _, ip := range []uint16{0x1a5c, 0x1a5e, 0x1a60, 0x1a64, 0x1a68, 0x1a6c, 0x1a6d, 0x1a6f, 0x1a71, 0x1a73, 0x1a75, 0x1a76, 0x1a77, 0x1a78, 0x1a7a, 0x1a7c} {
		valid = valid || v.IP() == ip
	}
	if !valid {
		return nil
	}
	var raw [35]byte
	copyView(v, 0x1ed40+0x1a5c, raw[:])
	if sha256.Sum256(raw[:]) != smallCalls[0x1ed41a5c].sum {
		return nil
	}
	sp, stack := uint32(v.SP()), uint32(v.SS())*16+uint32(v.SP())
	if sp+14 > 65536 || stack < 0x500 || stack+14 > 0xa0000 {
		return nil
	}
	count, dst, ds, src, ss := v.Read16(stack+4), v.Read16(stack+6), v.Read16(stack+8), v.Read16(stack+10), v.Read16(stack+12)
	backward := src < dst
	source, dest := uint32(ss)*16+uint32(src), uint32(ds)*16+uint32(dst)
	var data []byte
	if count != 0 {
		if uint32(src)+uint32(count) > 65536 || uint32(dst)+uint32(count) > 65536 || source+uint32(count) > 0xafa00 || dest < 0x500 || dest+uint32(count) > 0xafa00 || dest < 0xa0000 && dest+uint32(count) > 0xa0000 || overlapCode(dest, dest+uint32(count)) || dest < stack+14 && stack < dest+uint32(count) {
			return nil
		}
		if dest < source+uint32(count) && source < dest+uint32(count) && (backward && dest < source || !backward && dest > source) {
			return nil
		}
		data = make([]byte, int(count))
		copyView(v, source, data)
		if backward {
			for a, b := 0, len(data)-1; a < b; a, b = a+1, b-1 {
				data[a], data[b] = data[b], data[a]
			}
			dest += uint32(count) - 1
		}
	}
	current := boundedMemoryRegisters(v)
	r := current
	r.IP = 0x1a5c
	savedDS := v.DS()
	if v.IP() >= 0x1a64 && v.IP() <= 0x1a7a {
		savedDS = v.DX()
	}
	r.Seg[cpu.DS] = savedDS
	steps := []boundedMemoryCheckpoint{}
	add := func(next uint16, native, dos uint64, tr *machine.HostTransfer, change func()) {
		before := r
		change()
		r.IP = next
		steps = append(steps, boundedMemoryCheckpoint{before: before, after: r, nativeCycles: native, dosCycles: dos, transfer: tr})
	}
	none := func() {}
	add(0x1a5e, 2, 1, nil, func() { r.R[cpu.BX] = uint16(sp) })
	add(0x1a60, 2, 1, nil, func() { r.R[cpu.DX] = savedDS })
	add(0x1a64, 25, 1, nil, func() { r.R[cpu.SI] = src; r.Seg[cpu.DS] = ss })
	add(0x1a68, 25, 1, nil, func() { r.R[cpu.DI] = dst; r.Seg[cpu.ES] = ds })
	add(0x1a6c, 7, 1, nil, func() { r.R[cpu.CX] = count })
	add(0x1a6d, 2, 1, nil, func() { r.Flags &^= cpu.DF })
	add(0x1a6f, 2, 1, nil, func() { r.Flags = nativeCompareFlags(r.Flags, src, dst) })
	if backward {
		add(0x1a71, 2, 1, nil, none)
		add(0x1a73, 2, 1, nil, func() { r.Flags = copyAddFlags(r.Flags, src, count); r.R[cpu.SI] = src + count })
		add(0x1a75, 2, 1, nil, func() { r.Flags = copyAddFlags(r.Flags, dst, count); r.R[cpu.DI] = dst + count })
		add(0x1a76, 2, 1, nil, func() {
			r.Flags = ellipseSubFlags(r.Flags, r.R[cpu.SI], 1, 0)&^uint16(cpu.CF) | r.Flags&cpu.CF
			r.R[cpu.SI]--
		})
		add(0x1a77, 2, 1, nil, func() {
			r.Flags = ellipseSubFlags(r.Flags, r.R[cpu.DI], 1, 0)&^uint16(cpu.CF) | r.Flags&cpu.CF
			r.R[cpu.DI]--
		})
		add(0x1a78, 2, 1, nil, func() { r.Flags |= cpu.DF })
	} else {
		add(0x1a78, 7, 1, nil, none)
	}
	var tr *machine.HostTransfer
	if count != 0 {
		tr = &machine.HostTransfer{Address: dest, Width: 1, Count: count, Backward: backward, Data: data, CS: 0x1ed4, IP: 0x1a78}
	}
	add(0x1a7a, 3+5*uint64(count), 1+uint64(count), tr, func() {
		r.R[cpu.CX] = 0
		if backward {
			r.R[cpu.SI] = src - 1
			r.R[cpu.DI] = dst - 1
		} else {
			r.R[cpu.SI] = src + count
			r.R[cpu.DI] = dst + count
		}
	})
	add(0x1a7c, 20, 1, nil, func() { r.Seg[cpu.DS] = savedDS })
	add(v.Read16(stack), 24, 1, nil, func() { r.Seg[cpu.CS] = v.Read16(stack + 2); r.R[cpu.SP] += 14 })
	start := -1
	for i, s := range steps {
		if s.before == current {
			start = i
			break
		}
	}
	if start < 0 {
		return nil
	}
	n := len(steps) - start
	if uint64(n) > budget {
		n = int(budget)
	}
	p := &machine.HostWorkPlan{Algorithmic: true, Instructions: uint64(n), After: steps[start+n-1].after, LastCS: 0x1ed4, LastIP: steps[start+n-1].before.IP}
	for i := 0; i < n; i++ {
		s := steps[start+i]
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
