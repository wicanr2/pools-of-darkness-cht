package machine

import "fmt"

type HostVideoStatusRead struct {
	Value        byte
	Step, Cycles uint64
}

func (m *Machine) validateHostVideoStatusReads(p *HostWorkPlan) error {
	if len(p.VideoStatusReads) == 0 {
		return nil
	}
	if len(p.PaletteReads) > 0 || len(p.PaletteWrites) > 0 {
		return fmt.Errorf("machine: mixed host status/palette work")
	}
	snapshot, ok := m.VGAStatusSnapshot()
	if !ok {
		return fmt.Errorf("machine: unsupported host VGA status model")
	}
	moments := make(map[uint64]struct{}, len(p.VideoStatusReads))
	var step, cycles uint64
	for i, read := range p.VideoStatusReads {
		if read.Step <= step || read.Step > p.Instructions || read.Cycles <= cycles || read.Cycles > p.Cycles {
			return fmt.Errorf("machine: invalid host status moment")
		}
		value, known := snapshot.At(read.Step, uint64(i+1))
		if !known || value != read.Value {
			return fmt.Errorf("machine: invalid or stale host status prediction")
		}
		step, cycles = read.Step, read.Cycles
		moments[step] = struct{}{}
	}
	for _, w := range p.Writes {
		if _, ok := moments[w.Step]; ok {
			return fmt.Errorf("machine: ambiguous host status/write order")
		}
	}
	for _, s := range p.Transfers {
		if _, ok := moments[s.Step]; ok {
			return fmt.Errorf("machine: ambiguous host status/transfer order")
		}
	}
	return nil
}

func hostVideoStatusBytes(reads []HostVideoStatusRead) []hostPaletteByte {
	if len(reads) == 0 {
		return nil
	}
	out := make([]hostPaletteByte, len(reads))
	for i, r := range reads {
		out[i] = hostPaletteByte{port: 0x3da, value: r.Value, step: r.Step, cycles: r.Cycles, read: true}
	}
	return out
}
