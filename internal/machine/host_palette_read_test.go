package machine

import (
	"bytes"
	"reflect"
	"testing"
)

func hostPaletteReadFixture() (*Machine, *HostWorkPlan) {
	m, p := hostPaletteFixture()
	p.PaletteWrites = nil
	m.DAC[21], m.DAC[22], m.DAC[23] = 1, 2, 3
	p.PaletteReads = []HostPaletteRead{{Index: 7, RGB: [3]byte{1, 2, 3}, Steps: [4]uint64{2, 3, 4, 5}, Cycles: [4]uint64{20, 30, 40, 50}}}
	return m, p
}

func paletteReadState(t *testing.T, m *Machine) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := m.SaveState(&b); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestHostPaletteReadMomentsAndDeviceState(t *testing.T) {
	m, p := hostPaletteReadFixture()
	m.Steps, m.CPU.Cycles = 100, 300
	dac := m.DAC
	if err := m.CommitHostBlock(p); err != nil {
		t.Fatal(err)
	}
	want := []PortWrite{{Port: 0x3c7, Val: 7, Step: 102, Cycles: 320}}
	if !reflect.DeepEqual(m.PortLog, want) || m.PortsIn[0x3c9] != 3 || m.Ports[0x3c7] != 7 || m.portTicks != 3 || m.DAC != dac || m.dacReadIndex != 8 || m.dacReadPhase != 0 || m.Steps != 105 || m.CPU.Cycles != 350 || m.hostPalette != nil {
		t.Fatal("RGB read bypassed service or lost timing", m.PortLog)
	}
}

func TestHostPaletteRGBSnapshotIsReadOnly(t *testing.T) {
	m, _ := hostPaletteReadFixture()
	m.Out8(0x3c7, 19)
	m.In8(0x3c9)
	before := paletteReadState(t, m)
	rgb, ok := m.PaletteRGB(7)
	if !ok || rgb != ([3]byte{1, 2, 3}) || !bytes.Equal(before, paletteReadState(t, m)) {
		t.Fatal("RGB query changed machine")
	}
	m.SetVideoMode(0xd)
	if _, ok = m.PaletteRGB(7); ok {
		t.Fatal("unsupported mode queried")
	}
}

func TestHostPaletteReadRejectsBeforeMutation(t *testing.T) {
	for _, kind := range []string{"mode", "planar", "stale", "invalid-DAC", "red", "green", "blue", "zero-step", "same-step", "step-bounds", "zero-cycle", "same-cycle", "cycle-bounds", "RAM-conflict", "transfer-conflict", "job-order", "mixed-jobs", "event-budget"} {
		t.Run(kind, func(t *testing.T) {
			m, p := hostPaletteReadFixture()
			switch kind {
			case "mode":
				m.SetVideoMode(0xd)
			case "planar":
				m.planarOn = true
			case "stale":
				m.DAC[21] = 2
			case "invalid-DAC":
				m.DAC[21] = 64
				p.PaletteReads[0].RGB[0] = 64
			case "red":
				p.PaletteReads[0].RGB[0] = 64
			case "green":
				p.PaletteReads[0].RGB[1] = 64
			case "blue":
				p.PaletteReads[0].RGB[2] = 64
			case "zero-step":
				p.PaletteReads[0].Steps[0] = 0
			case "same-step":
				p.PaletteReads[0].Steps[1] = 2
			case "step-bounds":
				p.PaletteReads[0].Steps[3] = 6
			case "zero-cycle":
				p.PaletteReads[0].Cycles[0] = 0
			case "same-cycle":
				p.PaletteReads[0].Cycles[1] = 20
			case "cycle-bounds":
				p.PaletteReads[0].Cycles[3] = 51
			case "RAM-conflict":
				p.Writes = []HostWrite{{Address: 0x90000, Value: 5, Step: 2}}
			case "transfer-conflict":
				p.Transfers = []HostTransfer{{Address: 0x90000, Value: 5, Step: 2, Count: 1, Width: 1}}
			case "job-order":
				p.PaletteReads = append(p.PaletteReads, p.PaletteReads[0])
			case "mixed-jobs":
				p.PaletteWrites = []HostPaletteWrite{{Index: 8, RGB: [3]byte{1, 2, 3}, Steps: [4]uint64{1, 2, 3, 4}, Cycles: [4]uint64{10, 20, 30, 40}}}
			case "event-budget":
				m.VGAFrameEvery = 4
				m.nextFrame = 4
			}
			before := paletteReadState(t, m)
			ports := append([]PortWrite(nil), m.PortLog...)
			if m.CommitHostBlock(p) == nil {
				t.Fatal("invalid read admitted")
			}
			if !bytes.Equal(before, paletteReadState(t, m)) || !reflect.DeepEqual(ports, m.PortLog) || m.HostBlockCommits != 0 || m.hostPalette != nil {
				t.Fatal("decline changed machine")
			}
		})
	}
}

func TestHostPaletteReadOwnsJobsAndMergesWrites(t *testing.T) {
	m, p := hostPaletteReadFixture()
	p.Instructions, p.Cycles = 8, 80
	p.PaletteReads[0].Steps = [4]uint64{2, 3, 5, 6}
	p.PaletteReads[0].Cycles = [4]uint64{20, 30, 50, 60}
	p.Writes = []HostWrite{{Address: vgaLo, Value: 11, Step: 1}, {Address: vgaLo + 1, Value: 22, Step: 4}, {Address: vgaLo + 2, Value: 33, Step: 7}}
	type moment struct {
		step         uint64
		index, phase byte
	}
	var seen []moment
	m.ObserveVideoWrites(func(w VideoWrite) {
		seen = append(seen, moment{w.Step, m.dacReadIndex, m.dacReadPhase})
		p.PaletteReads[0].Index = 255
		p.PaletteReads[0].RGB = [3]byte{63, 63, 63}
		p.PaletteReads[0].Steps = [4]uint64{1, 1, 1, 1}
	})
	if err := m.CommitHostBlock(p); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(seen, []moment{{1, 0, 0}, {4, 7, 1}, {7, 8, 0}}) || m.PortLog[0].Val != 7 || m.PortsIn[0x3c9] != 3 {
		t.Fatal("RGB read ownership/merge differs", seen)
	}
}

func TestHostPaletteReadYieldsAtFrameBoundary(t *testing.T) {
	m, p := hostPaletteReadFixture()
	m.VGAFrameEvery = 5
	m.nextFrame = 5
	frames := 0
	m.SetOnFrame(func() {
		frames++
		if m.dacReadIndex != 8 || m.dacReadPhase != 0 {
			t.Fatal("frame saw partial RGB read")
		}
	})
	if err := m.CommitHostBlock(p); err != nil {
		t.Fatal(err)
	}
	if frames != 0 {
		t.Fatal("early frame")
	}
	if err := m.Step(); err != nil {
		t.Fatal(err)
	}
	if frames != 1 {
		t.Fatal("lost boundary frame")
	}
}
