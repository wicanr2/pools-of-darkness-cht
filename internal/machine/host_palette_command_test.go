package machine

import (
	"bytes"
	"reflect"
	"testing"
)

func commandFixtureResearch() (*Machine, *HostWorkPlan) {
	m, p := hostPaletteFixture()
	p.PaletteWrites = nil
	p.PaletteCommands = []HostPaletteCommand{{Kind: HostPaletteSelectIndex, Value: 7, Step: 2, Cycles: 20}, {Kind: HostPaletteWriteData, Value: 1, Step: 3, Cycles: 30}, {Kind: HostPaletteWriteData, Value: 2, Step: 4, Cycles: 40}, {Kind: HostPaletteWriteData, Value: 3, Step: 5, Cycles: 50}}
	return m, p
}
func TestHostPaletteCommandsExistingDevice(t *testing.T) {
	cases := 0
	for phase := 0; phase < 3; phase++ {
		for _, kind := range []string{"index", "data", "partial", "full"} {
			a, _ := commandFixtureResearch()
			b, p := commandFixtureResearch()
			for _, m := range []*Machine{a, b} {
				m.Out8(0x3c8, 17)
				for i := 0; i < phase; i++ {
					m.Out8(0x3c9, byte(i+11))
				}
				m.Out8(0x3c7, 8)
				m.In8(0x3c9)
				m.PortLog = nil
			}
			switch kind {
			case "index":
				p.PaletteCommands = p.PaletteCommands[:1]
			case "data":
				p.PaletteCommands = []HostPaletteCommand{{Kind: HostPaletteWriteData, Value: 63, Step: 2, Cycles: 20}}
			case "partial":
				p.PaletteCommands = p.PaletteCommands[:3]
			}
			for _, command := range p.PaletteCommands {
				a.Steps = command.Step
				a.CPU.Cycles = command.Cycles
				port := uint16(0x3c8)
				if command.Kind == HostPaletteWriteData {
					port = 0x3c9
				}
				a.Out8(port, command.Value)
			}
			if err := b.CommitHostBlock(p); err != nil {
				t.Fatal(err)
			}
			if a.DAC != b.DAC || !reflect.DeepEqual(a.VGA, b.VGA) || !reflect.DeepEqual(a.Ports, b.Ports) || !reflect.DeepEqual(a.PortLog, b.PortLog) {
				t.Fatalf("DAC command differs phase%d kind%s", phase, kind)
			}
			cases++
		}
	}
	t.Logf("DAC commands existing device/phase comparisons=%d", cases)
}
func TestHostPaletteCommandsRejectAtomically(t *testing.T) {
	cases := 0
	for _, kind := range []string{"unknown", "data-range", "zero-step", "step-order", "step-outside", "zero-cycles", "cycle-order", "cycle-outside", "oversized", "mode", "planar", "RGB-mix", "read-mix", "status-mix", "RAM-collision", "transfer-collision"} {
		m, p := commandFixtureResearch()
		switch kind {
		case "unknown":
			p.PaletteCommands[0].Kind = 0
		case "data-range":
			p.PaletteCommands[1].Value = 64
		case "zero-step":
			p.PaletteCommands[0].Step = 0
		case "step-order":
			p.PaletteCommands[1].Step = 2
		case "step-outside":
			p.PaletteCommands[3].Step = 6
		case "zero-cycles":
			p.PaletteCommands[0].Cycles = 0
		case "cycle-order":
			p.PaletteCommands[1].Cycles = 20
		case "cycle-outside":
			p.PaletteCommands[3].Cycles = 51
		case "oversized":
			p.PaletteCommands = append(p.PaletteCommands, p.PaletteCommands[3])
		case "mode":
			m.Mem[0x449] = 3
		case "planar":
			m.planarOn = true
		case "RGB-mix":
			p.PaletteWrites = []HostPaletteWrite{{Index: 8, RGB: [3]byte{1, 2, 3}, Steps: [4]uint64{2, 3, 4, 5}, Cycles: [4]uint64{20, 30, 40, 50}}}
		case "read-mix":
			p.PaletteReads = []HostPaletteRead{{}}
		case "status-mix":
			p.VideoStatusReads = []HostVideoStatusRead{{}}
		case "RAM-collision":
			p.Writes = []HostWrite{{Address: 0x60000, Value: 42, Step: 2}}
		case "transfer-collision":
			p.Transfers = []HostTransfer{{Address: 0x60000, Value: 42, Width: 1, Count: 1, Step: 2}}
		}
		var before, after bytes.Buffer
		if e := m.SaveState(&before); e != nil {
			t.Fatal(e)
		}
		if e := m.CommitHostBlock(p); e == nil {
			t.Fatal("invalid DAC commands accepted", kind)
		}
		if e := m.SaveState(&after); e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(before.Bytes(), after.Bytes()) || len(m.PortLog) != 0 {
			t.Fatal("rejected DAC commands changed state", kind)
		}
		cases++
	}
	t.Logf("DAC command atomic refusal controls=%d", cases)
}
func TestHostPaletteCommandsOwnAndOrder(t *testing.T) {
	m, p := commandFixtureResearch()
	p.Writes = []HostWrite{{Address: vgaLo, Value: 12, Step: 1}}
	mutated := false
	m.ObserveVideoWrites(func(w VideoWrite) {
		mutated = true
		p.PaletteCommands[0].Value = 255
		p.PaletteCommands[1].Value = 63
		p.PaletteCommands[1].Step = 1
	})
	if e := m.CommitHostBlock(p); e != nil {
		t.Fatal(e)
	}
	if !mutated || m.Peek8(vgaLo) != 12 || m.DAC[21] != 1 || m.DAC[22] != 2 || m.DAC[23] != 3 || len(m.PortLog) != 4 || m.PortLog[0].Step != 2 {
		t.Fatal("DAC command ownership/order differs")
	}
}
func TestHostPaletteCommandsYieldBetweenComponents(t *testing.T) {
	m, p := commandFixtureResearch()
	p.Instructions = 4
	p.Cycles = 40
	p.PaletteCommands = p.PaletteCommands[:3]
	m.VGAFrameEvery = 4
	m.nextFrame = 4
	frames := 0
	m.SetOnFrame(func() {
		frames++
		if m.DAC[21] != 1 || m.DAC[22] != 2 || m.DAC[23] != 0 {
			t.Fatal("partial component frame differs")
		}
	})
	if e := m.CommitHostBlock(p); e != nil {
		t.Fatal(e)
	}
	if frames != 0 {
		t.Fatal("frame before budget boundary")
	}
	if e := m.Step(); e != nil {
		t.Fatal(e)
	}
	if frames != 1 {
		t.Fatal("missing partial palette frame")
	}
}
