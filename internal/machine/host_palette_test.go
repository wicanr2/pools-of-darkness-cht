package machine

import (
	"bytes"
	"maps"
	"reflect"
	"testing"
)

func hostPaletteFixture() (*Machine, *HostWorkPlan) {
	m := New()
	m.SetVideoMode(0x13)
	m.IRQ0Every = 0
	m.VGAFrameEvery = 0
	m.CPU.Seg[1] = 0x1000
	m.CPU.IP = 0x100
	m.Mem[0x10100] = 0x90
	p := &HostWorkPlan{Instructions: 5, Cycles: 50, After: HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, PaletteWrites: []HostPaletteWrite{{Index: 7, RGB: [3]byte{1, 2, 3}, Steps: [4]uint64{2, 3, 4, 5}, Cycles: [4]uint64{20, 30, 40, 50}}}}
	return m, p
}

func TestHostPaletteMomentsAndDeviceState(t *testing.T) {
	m, p := hostPaletteFixture()
	m.Steps = 100
	m.CPU.Cycles = 300
	if err := m.CommitHostBlock(p); err != nil {
		t.Fatal(err)
	}
	want := []PortWrite{{Port: 0x3c8, Val: 7, Step: 102, Cycles: 320}, {Port: 0x3c9, Val: 1, Step: 103, Cycles: 330}, {Port: 0x3c9, Val: 2, Step: 104, Cycles: 340}, {Port: 0x3c9, Val: 3, Step: 105, Cycles: 350}}
	if !reflect.DeepEqual(m.PortLog, want) || m.Steps != 105 || m.CPU.Cycles != 350 || m.hostPalette != nil {
		t.Fatal("palette timing/context differs", m.PortLog)
	}
	if m.DAC[7*3] != 1 || m.DAC[7*3+1] != 2 || m.DAC[7*3+2] != 3 || m.dacIndex != 8 || m.dacPhase != 0 || m.Ports[0x3c8] != 7 || m.Ports[0x3c9] != 3 {
		t.Fatal("DAC/ports were bypassed")
	}
}

func TestHostPaletteRejectsBeforeMutation(t *testing.T) {
	for _, kind := range []string{"mode", "planar", "red", "green", "blue", "zero-step", "same-step", "step-bounds", "zero-cycle", "same-cycle", "cycle-bounds", "ram-conflict", "transfer-conflict", "job-order", "event-budget"} {
		t.Run(kind, func(t *testing.T) {
			m, p := hostPaletteFixture()
			switch kind {
			case "mode":
				m.SetVideoMode(0xd)
			case "planar":
				m.planarOn = true
			case "red":
				p.PaletteWrites[0].RGB[0] = 64
			case "green":
				p.PaletteWrites[0].RGB[1] = 64
			case "blue":
				p.PaletteWrites[0].RGB[2] = 64
			case "zero-step":
				p.PaletteWrites[0].Steps[0] = 0
			case "same-step":
				p.PaletteWrites[0].Steps[1] = 2
			case "step-bounds":
				p.PaletteWrites[0].Steps[3] = 6
			case "zero-cycle":
				p.PaletteWrites[0].Cycles[0] = 0
			case "same-cycle":
				p.PaletteWrites[0].Cycles[1] = 20
			case "cycle-bounds":
				p.PaletteWrites[0].Cycles[3] = 51
			case "ram-conflict":
				p.Writes = []HostWrite{{Address: 0x90000, Step: 2, Value: 0xff}}
			case "transfer-conflict":
				p.Transfers = []HostTransfer{{Address: 0x90000, Step: 2, Count: 1, Width: 1, Value: 0xff}}
			case "job-order":
				p.PaletteWrites = append(p.PaletteWrites, p.PaletteWrites[0])
			case "event-budget":
				m.VGAFrameEvery = 4
				m.nextFrame = 4
			}
			mem := append([]byte(nil), m.Mem...)
			dac := m.DAC
			vga := *m.VGA
			ports := maps.Clone(m.Ports)
			state := HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}
			if err := m.CommitHostBlock(p); err == nil {
				t.Fatal("invalid palette accepted")
			}
			if !bytes.Equal(mem, m.Mem) || dac != m.DAC || !reflect.DeepEqual(vga, *m.VGA) || !reflect.DeepEqual(ports, m.Ports) || len(m.PortLog) != 0 || m.Steps != 0 || m.CPU.Cycles != 0 || state != (HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}) {
				t.Fatal("rejected job mutated machine")
			}
		})
	}
}

func TestHostPaletteOwnsJobsAndMergesVideoWrites(t *testing.T) {
	m, p := hostPaletteFixture()
	p.Writes = []HostWrite{{Address: vgaLo, Step: 1, Value: 12}}
	mutated := false
	m.ObserveVideoWrites(func(w VideoWrite) {
		mutated = true
		p.PaletteWrites[0].RGB = [3]byte{63, 63, 63}
		p.PaletteWrites[0].Steps = [4]uint64{1, 1, 1, 1}
	})
	if err := m.CommitHostBlock(p); err != nil {
		t.Fatal(err)
	}
	if !mutated || m.Peek8(vgaLo) != 12 || m.DAC[21] != 1 || m.DAC[22] != 2 || m.DAC[23] != 3 || len(m.PortLog) != 4 || m.PortLog[0].Step != 2 {
		t.Fatal("job ownership/order differs")
	}
}

func TestHostPaletteYieldsAtFrameBoundary(t *testing.T) {
	m, p := hostPaletteFixture()
	m.VGAFrameEvery = 5
	m.nextFrame = 5
	frames := 0
	m.SetOnFrame(func() {
		frames++
		if m.DAC[21] != 1 || m.DAC[22] != 2 || m.DAC[23] != 3 {
			t.Fatal("frame saw partial RGB")
		}
	})
	if err := m.CommitHostBlock(p); err != nil {
		t.Fatal(err)
	}
	if frames != 0 {
		t.Fatal("frame arrived before boundary")
	}
	if err := m.Step(); err != nil {
		t.Fatal(err)
	}
	if frames != 1 {
		t.Fatal("boundary frame lost")
	}
}
