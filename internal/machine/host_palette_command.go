package machine

import "fmt"

type HostPaletteCommandKind uint8

const (
	HostPaletteSelectIndex HostPaletteCommandKind = iota + 1
	HostPaletteWriteData
)

// HostPaletteCommand addresses only the existing DAC index/data service.
// It preserves the device's current phase; it cannot choose arbitrary ports.
type HostPaletteCommand struct {
	Kind         HostPaletteCommandKind
	Value        byte
	Step, Cycles uint64
}

func (m *Machine) validateHostPaletteCommands(p *HostWorkPlan) error {
	if len(p.PaletteCommands) == 0 {
		return nil
	}
	if len(p.PaletteCommands) > 4 || len(p.PaletteWrites) > 0 || len(p.PaletteReads) > 0 || len(p.VideoStatusReads) > 0 {
		return fmt.Errorf("machine: mixed or oversized host DAC commands")
	}
	if m.planarOn || m.Unchained() || m.Peek8(0x449) != 0x13 {
		return fmt.Errorf("machine: unsupported host DAC command mode")
	}
	var step, cycles uint64
	moments := map[uint64]bool{}
	for _, command := range p.PaletteCommands {
		if command.Kind != HostPaletteSelectIndex && command.Kind != HostPaletteWriteData || command.Kind == HostPaletteWriteData && command.Value > 63 {
			return fmt.Errorf("machine: invalid host DAC command value")
		}
		if command.Step <= step || command.Step > p.Instructions || command.Cycles <= cycles || command.Cycles > p.Cycles {
			return fmt.Errorf("machine: invalid host DAC command moment")
		}
		step, cycles = command.Step, command.Cycles
		moments[step] = true
	}
	for _, write := range p.Writes {
		if moments[write.Step] {
			return fmt.Errorf("machine: ambiguous host DAC/write order")
		}
	}
	for _, transfer := range p.Transfers {
		if moments[transfer.Step] {
			return fmt.Errorf("machine: ambiguous host DAC/transfer order")
		}
	}
	return nil
}

func hostPaletteCommandBytes(commands []HostPaletteCommand) []hostPaletteByte {
	out := make([]hostPaletteByte, 0, len(commands))
	for _, command := range commands {
		port := uint16(0x3c8)
		if command.Kind == HostPaletteWriteData {
			port = 0x3c9
		}
		out = append(out, hostPaletteByte{port: port, value: command.Value, step: command.Step, cycles: command.Cycles})
	}
	return out
}
