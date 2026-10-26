package machine

import (
	"bytes"
	"reflect"
	"testing"
)

func hostStatusFixture() (*Machine, *HostWorkPlan) {
	m, p := hostPaletteFixture()
	p.PaletteWrites = nil
	s, ok := m.VGAStatusSnapshot()
	if !ok {
		panic("fixture status model")
	}
	for i, step := range []uint64{2, 4, 5} {
		value, _ := s.At(step, uint64(i+1))
		p.VideoStatusReads = append(p.VideoStatusReads, HostVideoStatusRead{Value: value, Step: step, Cycles: step * 10})
	}
	return m, p
}

func TestHostStatusUsesExistingServiceAtEachMoment(t *testing.T) {
	cases := 0
	for _, phase := range []bool{false, true} {
		for _, start := range []uint64{0, 151000} {
			for _, pre := range []int{0, 15, 31} {
				a, _ := hostPaletteFixture()
				b, _ := hostPaletteFixture()
				a.Mode13Phase, b.Mode13Phase = phase, phase
				a.VGAFrameEvery, b.VGAFrameEvery = 165000, 165000
				for i := 0; i < pre; i++ {
					a.In8(0x3da)
					b.In8(0x3da)
				}
				a.Steps, b.Steps = start, start
				s, _ := a.VGAStatusSnapshot()
				for i, at := range []uint64{1, 2, 40, 500, 1000} {
					expected, ok := s.At(at, uint64(i+1))
					if !ok {
						t.Fatal("valid prediction rejected")
					}
					command := hostPaletteByte{port: 0x3da, value: expected, step: at, cycles: at * 2, read: true}
					actual := a.executeHostVideoIO(&command)
					b.Steps = start + at
					original := b.In8(0x3da)
					if actual != expected || actual != original || a.hostPalette != nil {
						t.Fatal("actual status service lost moment", phase, start, pre, at, actual, expected, original)
					}
					cases++
				}
				if !reflect.DeepEqual(a.PortsIn, b.PortsIn) || !reflect.DeepEqual(a.VGA, b.VGA) || a.portTicks != b.portTicks {
					t.Fatal("status helper bypassed device effects")
				}
			}
		}
	}
	t.Log("actual status I/O reference comparisons", cases)
}

func TestHostStatusCommitAndACPhase(t *testing.T) {
	m, p := hostStatusFixture()
	m.VGA.acFlip = true
	if err := m.CommitHostBlock(p); err != nil {
		t.Fatal(err)
	}
	if m.VGA.acFlip || m.PortsIn[0x3da] != 3 || m.portTicks != 3 || len(m.PortLog) != 0 || m.Steps != 5 || m.CPU.Cycles != 50 || m.hostPalette != nil {
		t.Fatal("status work bypassed service or counters")
	}
}

func TestHostStatusRejectsBeforeMutation(t *testing.T) {
	for _, kind := range []string{"mode", "planar", "CRT", "value", "zero-step", "same-step", "step-bounds", "zero-cycle", "same-cycle", "cycle-bounds", "RAM-conflict", "transfer-conflict", "palette-write", "palette-read", "read-order", "event-budget", "counter-overflow"} {
		t.Run(kind, func(t *testing.T) {
			m, p := hostStatusFixture()
			switch kind {
			case "mode":
				m.SetVideoMode(0xd)
			case "planar":
				m.planarOn = true
			case "CRT":
				m.CRTCTiming = true
			case "value":
				p.VideoStatusReads[0].Value ^= 8
			case "zero-step":
				p.VideoStatusReads[0].Step = 0
			case "same-step":
				p.VideoStatusReads[1].Step = 2
			case "step-bounds":
				p.VideoStatusReads[2].Step = 6
			case "zero-cycle":
				p.VideoStatusReads[0].Cycles = 0
			case "same-cycle":
				p.VideoStatusReads[1].Cycles = 20
			case "cycle-bounds":
				p.VideoStatusReads[2].Cycles = 51
			case "RAM-conflict":
				p.Writes = []HostWrite{{Address: 0x90000, Value: 3, Step: 2}}
			case "transfer-conflict":
				p.Transfers = []HostTransfer{{Address: 0x90000, Value: 3, Step: 2, Width: 1, Count: 1}}
			case "palette-write":
				p.PaletteWrites = []HostPaletteWrite{{Index: 7, RGB: [3]byte{1, 2, 3}, Steps: [4]uint64{1, 2, 3, 4}, Cycles: [4]uint64{10, 20, 30, 40}}}
			case "palette-read":
				rgb, _ := m.PaletteRGB(7)
				p.PaletteReads = []HostPaletteRead{{Index: 7, RGB: rgb, Steps: [4]uint64{1, 2, 3, 4}, Cycles: [4]uint64{10, 20, 30, 40}}}
			case "read-order":
				p.VideoStatusReads[0], p.VideoStatusReads[1] = p.VideoStatusReads[1], p.VideoStatusReads[0]
			case "event-budget":
				m.VGAFrameEvery = 4
				m.nextFrame = 4
			case "counter-overflow":
				m.portTicks = ^uint64(0)
			}
			before := paletteReadState(t, m)
			ports := append([]PortWrite(nil), m.PortLog...)
			if m.CommitHostBlock(p) == nil {
				t.Fatal("invalid status accepted")
			}
			if !bytes.Equal(before, paletteReadState(t, m)) || !reflect.DeepEqual(ports, m.PortLog) || m.HostBlockCommits != 0 || m.hostPalette != nil {
				t.Fatal("rejected status work changed machine")
			}
		})
	}
}

func TestHostStatusOwnsCommandsAndMergesWrites(t *testing.T) {
	m, p := hostStatusFixture()
	p.Instructions, p.Cycles = 8, 80
	p.VideoStatusReads = []HostVideoStatusRead{{Value: 0, Step: 2, Cycles: 20}, {Value: 0, Step: 5, Cycles: 50}}
	p.Writes = []HostWrite{{Address: vgaLo, Value: 1, Step: 1}, {Address: vgaLo + 1, Value: 2, Step: 4}, {Address: vgaLo + 2, Value: 3, Step: 7}}
	var seen []uint64
	m.ObserveVideoWrites(func(w VideoWrite) {
		seen = append(seen, m.PortsIn[0x3da])
		p.VideoStatusReads[0].Value = 255
		p.VideoStatusReads[1].Step = 1
	})
	if err := m.CommitHostBlock(p); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(seen, []uint64{0, 1, 2}) || m.PortsIn[0x3da] != 2 {
		t.Fatal("status command ownership/order differs", seen)
	}
}

func TestHostStatusYieldsAtEventBoundary(t *testing.T) {
	m, p := hostStatusFixture()
	m.VGAFrameEvery = 5
	m.nextFrame = 5
	frames := 0
	m.SetOnFrame(func() {
		frames++
		if m.PortsIn[0x3da] != 3 {
			t.Fatal("frame saw partial status work")
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
		t.Fatal("missing frame")
	}
}
