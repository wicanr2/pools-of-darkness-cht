package machine

import "fmt"

// HostPaletteWrite is one RGB job for the existing VGA DAC service. It cannot
// select arbitrary ports. Moments preserve the pinned guest timing model.
type HostPaletteWrite struct {
	Index         byte
	RGB           [3]byte
	Steps, Cycles [4]uint64
}

type hostPaletteByte struct {
	port         uint16
	value        byte
	step, cycles uint64
	read         bool
}

// Shared internal clock context for the bounded VGA commands. The public
// commands still select only their documented palette/status interfaces.
func (m *Machine) executeHostVideoIO(command *hostPaletteByte) byte {
	m.hostPalette = command
	defer func() { m.hostPalette = nil }()
	if command.read {
		return m.In8(command.port)
	}
	m.Out8(command.port, command.value)
	return 0
}

func (m *Machine) validateHostPaletteWrites(p *HostWorkPlan) error {
	if len(p.PaletteWrites) == 0 {
		return nil
	}
	if m.planarOn || m.Unchained() || m.Peek8(0x449) != 0x13 {
		return fmt.Errorf("machine: unsupported host palette mode")
	}
	moments := make(map[uint64]struct{}, 4*len(p.PaletteWrites))
	var step, cycles uint64
	for _, job := range p.PaletteWrites {
		if job.RGB[0] > 63 || job.RGB[1] > 63 || job.RGB[2] > 63 {
			return fmt.Errorf("machine: invalid host palette RGB")
		}
		for i := range job.Steps {
			if job.Steps[i] <= step || job.Steps[i] > p.Instructions || job.Cycles[i] <= cycles || job.Cycles[i] > p.Cycles {
				return fmt.Errorf("machine: invalid host palette moment")
			}
			step, cycles = job.Steps[i], job.Cycles[i]
			moments[step] = struct{}{}
		}
	}
	for _, w := range p.Writes {
		if _, ok := moments[w.Step]; ok {
			return fmt.Errorf("machine: ambiguous host palette/write order")
		}
	}
	for _, w := range p.Transfers {
		if _, ok := moments[w.Step]; ok {
			return fmt.Errorf("machine: ambiguous host palette/transfer order")
		}
	}
	return nil
}

func hostPaletteBytes(jobs []HostPaletteWrite) []hostPaletteByte {
	if len(jobs) == 0 {
		return nil
	}
	out := make([]hostPaletteByte, 0, 4*len(jobs))
	for _, job := range jobs {
		for i, value := range [4]byte{job.Index, job.RGB[0], job.RGB[1], job.RGB[2]} {
			port := uint16(0x3c9)
			if i == 0 {
				port = 0x3c8
			}
			out = append(out, hostPaletteByte{port: port, value: value, step: job.Steps[i], cycles: job.Cycles[i]})
		}
	}
	return out
}
