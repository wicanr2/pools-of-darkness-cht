package machine

import "fmt"

// HostPaletteRead selects one DAC color and reads its RGB triplet through
// the existing VGA service. RGB is a revalidated snapshot, not supplied data.
type HostPaletteRead struct {
	Index         byte
	RGB           [3]byte
	Steps, Cycles [4]uint64
}

func (m *Machine) PaletteRGB(index byte) ([3]byte, bool) {
	if m.planarOn || m.Unchained() || m.Peek8(0x449) != 0x13 {
		return [3]byte{}, false
	}
	at := int(index) * 3
	rgb := [3]byte{m.DAC[at], m.DAC[at+1], m.DAC[at+2]}
	if rgb[0] > 63 || rgb[1] > 63 || rgb[2] > 63 {
		return [3]byte{}, false
	}
	return rgb, true
}

func (m *Machine) validateHostPaletteReads(p *HostWorkPlan) error {
	if len(p.PaletteReads) == 0 {
		return nil
	}
	if len(p.PaletteWrites) > 0 {
		return fmt.Errorf("machine: mixed host palette read/write jobs")
	}
	moments := make(map[uint64]struct{}, 4*len(p.PaletteReads))
	var step, cycles uint64
	for _, job := range p.PaletteReads {
		rgb, ok := m.PaletteRGB(job.Index)
		if !ok || rgb != job.RGB {
			return fmt.Errorf("machine: invalid or stale host palette read snapshot")
		}
		for i := range job.Steps {
			if job.Steps[i] <= step || job.Steps[i] > p.Instructions || job.Cycles[i] <= cycles || job.Cycles[i] > p.Cycles {
				return fmt.Errorf("machine: invalid host palette read moment")
			}
			step, cycles = job.Steps[i], job.Cycles[i]
			moments[step] = struct{}{}
		}
	}
	for _, w := range p.Writes {
		if _, ok := moments[w.Step]; ok {
			return fmt.Errorf("machine: ambiguous host palette read/write order")
		}
	}
	for _, s := range p.Transfers {
		if _, ok := moments[s.Step]; ok {
			return fmt.Errorf("machine: ambiguous host palette read/transfer order")
		}
	}
	return nil
}

func hostPaletteReadBytes(jobs []HostPaletteRead) []hostPaletteByte {
	if len(jobs) == 0 {
		return nil
	}
	out := make([]hostPaletteByte, 0, 4*len(jobs))
	for _, job := range jobs {
		out = append(out, hostPaletteByte{port: 0x3c7, value: job.Index, step: job.Steps[0], cycles: job.Cycles[0]})
		for i := range job.RGB {
			out = append(out, hostPaletteByte{port: 0x3c9, value: job.RGB[i], step: job.Steps[i+1], cycles: job.Cycles[i+1], read: true})
		}
	}
	return out
}
