package drawblocks

import (
	"crypto/sha256"
	"github.com/wicanr2/dosgolem/internal/cpu"
	machine "github.com/wicanr2/dosgolem/internal/machine"
)

type paletteCallCheckpoint struct {
	before, after machine.HostRegisters
	cycles        uint64
	writes        []machine.HostWrite
	command       *machine.HostPaletteCommand
}

// Decode one RGB call frame. The VGA commands come from its index/channels;
// checkpoints only preserve the original ABI and scheduler-visible moments.
func planPaletteOutputBounded(v CPUView, budget uint64) *machine.HostWorkPlan {
	if budget == 0 || v.CS() != 0x48a1 || v.Read8(0x449) != 0x13 {
		return nil
	}
	known := false
	for _, ip := range []uint16{0x2aa8, 0x2aa9, 0x2aab, 0x2aac, 0x2aaf, 0x2ab2, 0x2ab3, 0x2ab4, 0x2ab7, 0x2ab8, 0x2abb, 0x2abc, 0x2abf, 0x2ac0, 0x2ac1, 0x2ac2} {
		known = known || v.IP() == ip
	}
	if !known {
		return nil
	}
	modelView, ok := v.(interface{ CPUModel() (cpu.Model, bool) })
	if !ok {
		return nil
	}
	model, ok := modelView.CPUModel()
	if !ok || model != cpu.Model80386 {
		return nil
	}
	var raw [27]byte
	copyView(v, 0x48a10+0x2aa8, raw[:])
	if sha256.Sum256(raw[:]) != paletteOutputCode.sum {
		return nil
	}
	sp := uint32(v.SP())
	if v.IP() != 0x2aa8 && v.IP() != 0x2ac2 {
		sp += 2
	}
	stack := uint32(v.SS())*16 + sp
	if sp < 2 || sp+12 > 65536 || stack-2 < 0x500 || stack+12 > 0xa0000 || stack-2 < 0x48a10+0x7580 && 0x48a10 < stack+12 {
		return nil
	}
	index := v.Read16(stack + 4)
	rgb := [3]uint16{v.Read16(stack + 6), v.Read16(stack + 8), v.Read16(stack + 10)}
	if index > 255 || rgb[0] > 63 || rgb[1] > 63 || rgb[2] > 63 {
		return nil
	}
	oldBP := v.BP()
	if v.IP() != 0x2aa8 {
		oldBP = v.Read16(stack - 2)
	}
	current := boundedMemoryRegisters(v)
	r := current
	r.IP = 0x2aa8
	r.R[cpu.SP] = uint16(sp)
	r.R[cpu.BP] = oldBP
	checkpoints := []paletteCallCheckpoint{}
	add := func(next uint16, cycles uint64, writes []machine.HostWrite, command *machine.HostPaletteCommand, change func()) {
		before := r
		change()
		r.IP = next
		checkpoints = append(checkpoints, paletteCallCheckpoint{before: before, after: r, cycles: cycles, writes: writes, command: command})
	}
	none := func() {}
	add(0x2aa9, 4, []machine.HostWrite{{Address: stack - 2, Value: byte(oldBP), CS: 0x48a1, IP: 0x2aa8}, {Address: stack - 1, Value: byte(oldBP >> 8), CS: 0x48a1, IP: 0x2aa8}}, nil, func() { r.R[cpu.SP] -= 2 })
	add(0x2aab, 2, nil, nil, func() { r.R[cpu.BP] = uint16(sp) - 2 })
	add(0x2aac, 2, nil, nil, func() { r.Flags &^= cpu.IF })
	add(0x2aaf, 6, nil, nil, func() { r.R[cpu.AX] = index })
	add(0x2ab2, 2, nil, nil, func() { r.R[cpu.DX] = 0x3c8 })
	add(0x2ab3, 16, nil, &machine.HostPaletteCommand{Kind: machine.HostPaletteSelectIndex, Value: byte(index)}, none)
	add(0x2ab4, 2, nil, nil, func() {
		r.R[cpu.DX] = 0x3c9
		r.Flags &^= uint16(cpu.PF | cpu.AF | cpu.ZF | cpu.SF | cpu.OF)
		r.Flags |= cpu.PF
	})
	add(0x2ab7, 6, nil, nil, func() { r.R[cpu.AX] = rgb[0] })
	add(0x2ab8, 16, nil, &machine.HostPaletteCommand{Kind: machine.HostPaletteWriteData, Value: byte(rgb[0])}, none)
	add(0x2abb, 6, nil, nil, func() { r.R[cpu.AX] = rgb[1] })
	add(0x2abc, 16, nil, &machine.HostPaletteCommand{Kind: machine.HostPaletteWriteData, Value: byte(rgb[1])}, none)
	add(0x2abf, 6, nil, nil, func() { r.R[cpu.AX] = rgb[2] })
	add(0x2ac0, 16, nil, &machine.HostPaletteCommand{Kind: machine.HostPaletteWriteData, Value: byte(rgb[2])}, none)
	add(0x2ac1, 2, nil, nil, func() { r.Flags |= cpu.IF })
	add(0x2ac2, 4, nil, nil, func() { r.R[cpu.BP] = oldBP; r.R[cpu.SP] += 2 })
	add(v.Read16(stack), 24, nil, nil, func() { r.Seg[cpu.CS] = v.Read16(stack + 2); r.R[cpu.SP] += 4 })
	first := -1
	for i, c := range checkpoints {
		if c.before == current {
			first = i
			break
		}
	}
	if first < 0 {
		return nil
	}
	n := len(checkpoints) - first
	if uint64(n) > budget {
		n = int(budget)
	}
	p := &machine.HostWorkPlan{Algorithmic: true, Instructions: uint64(n), After: checkpoints[first+n-1].after, LastCS: 0x48a1, LastIP: checkpoints[first+n-1].before.IP}
	for i := 0; i < n; i++ {
		c := checkpoints[first+i]
		if v.DOSBoxCost() {
			p.Cycles++
		} else {
			p.Cycles += c.cycles
		}
		for _, write := range c.writes {
			write.Step = uint64(i + 1)
			p.Writes = append(p.Writes, write)
		}
		if c.command != nil {
			command := *c.command
			command.Step = uint64(i + 1)
			command.Cycles = p.Cycles
			p.PaletteCommands = append(p.PaletteCommands, command)
		}
	}
	return p
}
