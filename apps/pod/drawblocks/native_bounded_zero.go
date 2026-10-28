// Spec025/#49: budgeted Go memory/image spans.
package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"math/bits"
)

// A zero region is decoded from the stable far-call frame. Pixel generation
// uses byte/word spans; checkpoints only project the original ABI and timing.
type boundedZeroRegion struct {
	before                                     machine.HostRegisters
	offset, segment, count, returnIP, returnCS uint16
}
type boundedMemoryCheckpoint struct {
	before, after           machine.HostRegisters
	nativeCycles, dosCycles uint64
	transfer                *machine.HostTransfer
}

func boundedMemoryRegisters(v CPUView) machine.HostRegisters {
	return machine.HostRegisters{R: [8]uint16{v.AX(), v.CX(), v.DX(), v.BX(), v.SP(), v.BP(), v.SI(), v.DI()}, Seg: [4]uint16{v.ES(), v.CS(), v.SS(), v.DS()}, IP: v.IP(), Flags: v.Flags()}
}

func planBoundedZero(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x1d1f || v.Read8(0x449) != 0x13 {
		return nil
	}
	valid := false
	for _, ip := range []uint16{0x46e, 0x470, 0x474, 0x478, 0x47a, 0x47b, 0x47d, 0x47f, 0x480, 0x482, 0x484} {
		valid = valid || v.IP() == ip
	}
	if !valid {
		return nil
	}
	var raw [25]byte
	copyView(v, 0x1d1f0+0x46e, raw[:])
	if sha256.Sum256(raw[:]) != smallCalls[0x1d1f046e].sum {
		return nil
	}
	sp, stack := uint32(v.SP()), uint32(v.SS())*16+uint32(v.SP())
	if sp+10 > 65536 || stack < 0x500 || stack+10 > 0xa0000 {
		return nil
	}
	j := boundedZeroRegion{before: boundedMemoryRegisters(v), count: v.Read16(stack + 4), offset: v.Read16(stack + 6), segment: v.Read16(stack + 8), returnIP: v.Read16(stack), returnCS: v.Read16(stack + 2)}
	address := uint32(j.segment)*16 + uint32(j.offset)
	end := address + uint32(j.count)
	if j.count != 0 && (uint32(j.offset)+uint32(j.count) > 65536 || address < 0x500 || end > 0xafa00 || address < 0xa0000 && end > 0xa0000 || overlapCode(address, end) || address < stack+10 && stack < end) {
		return nil
	}
	r := j.before
	r.IP = 0x46e
	steps := []boundedMemoryCheckpoint{}
	add := func(next uint16, native, dos uint64, transfer *machine.HostTransfer, change func()) {
		before := r
		change()
		r.IP = next
		steps = append(steps, boundedMemoryCheckpoint{before: before, after: r, nativeCycles: native, dosCycles: dos, transfer: transfer})
	}
	none := func() {}
	add(0x470, 2, 1, nil, func() { r.R[cpu.BX] = uint16(sp) })
	add(0x474, 25, 1, nil, func() { r.Seg[cpu.ES] = j.segment; r.R[cpu.DI] = j.offset })
	add(0x478, 7, 1, nil, func() { r.R[cpu.CX] = j.count })
	add(0x47a, 2, 1, nil, func() {
		r.R[cpu.AX] = 0
		r.Flags = r.Flags&^uint16(cpu.CF|cpu.PF|cpu.AF|cpu.ZF|cpu.SF|cpu.OF) | cpu.PF | cpu.ZF
	})
	add(0x47b, 2, 1, nil, func() { r.Flags &^= cpu.DF })
	words := j.count / 2
	add(0x47d, 2, 1, nil, func() {
		r.R[cpu.CX] = words
		r.Flags &^= uint16(cpu.CF | cpu.PF | cpu.ZF | cpu.SF | cpu.OF)
		if j.count&1 != 0 {
			r.Flags |= cpu.CF
		}
		if j.count&0x8000 != 0 {
			r.Flags |= cpu.OF
		}
		if words == 0 {
			r.Flags |= cpu.ZF
		}
		if bits.OnesCount8(byte(words))%2 == 0 {
			r.Flags |= cpu.PF
		}
	})
	if j.count&1 == 0 {
		add(0x482, 7, 1, nil, none)
	} else {
		add(0x47f, 2, 1, nil, none)
		tail := &machine.HostTransfer{Address: address, Value: 0, Count: 1, Width: 1, CS: 0x1d1f, IP: 0x47f}
		add(0x480, 7, 2, tail, func() { r.R[cpu.DI]++ })
		next := uint16(0x482)
		if words == 0 {
			next = 0x484
		}
		add(next, 2, 1, nil, none)
	}
	if words != 0 || j.count&1 == 0 {
		var fill *machine.HostTransfer
		if words != 0 {
			fill = &machine.HostTransfer{Address: address + uint32(j.count&1), Value: 0, Count: words, Width: 2, CS: 0x1d1f, IP: 0x482}
		}
		add(0x484, 3+5*uint64(words), 1+uint64(words), fill, func() { r.R[cpu.DI] = j.offset + j.count; r.R[cpu.CX] = 0 })
	}
	add(j.returnIP, 24, 1, nil, func() { r.R[cpu.SP] += 10; r.Seg[cpu.CS] = j.returnCS })
	start := -1
	for i, s := range steps {
		if s.before == j.before {
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
	p := &machine.HostWorkPlan{Algorithmic: true, Instructions: uint64(n), After: steps[start+n-1].after, LastCS: 0x1d1f, LastIP: steps[start+n-1].before.IP}
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
