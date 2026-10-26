package machine

import (
	"bytes"
	"testing"
)

func TestVGAStatusSnapshotMatchesExistingService(t *testing.T) {
	cases := 0
	for _, phase := range []bool{false, true} {
		for _, period := range []uint64{0, 165000} {
			for _, reads := range []int{0, 15, 16, 31} {
				for _, start := range []uint64{0, 149000, 151000} {
					m, _ := hostPaletteFixture()
					m.Mode13Phase = phase
					m.VGAFrameEvery = period
					for i := 0; i < reads; i++ {
						m.In8(0x3da)
					}
					m.Steps = start
					before := paletteReadState(t, m)
					snapshot, ok := m.VGAStatusSnapshot()
					if !ok {
						t.Fatal("supported status model rejected")
					}
					moments := []uint64{1, 2, 40, 1000, 20000}
					predicted := make([]byte, len(moments))
					for i, at := range moments {
						value, known := snapshot.At(at, uint64(i+1))
						if !known {
							t.Fatal("valid status query rejected")
						}
						predicted[i] = value
					}
					if !bytes.Equal(before, paletteReadState(t, m)) {
						t.Fatal("status snapshot query changed machine")
					}
					for i, at := range moments {
						m.Steps = start + at
						if got := m.In8(0x3da); got != predicted[i] {
							t.Fatal("status model prediction differs", phase, period, reads, start, at, got, predicted[i])
						}
						cases++
					}
				}
			}
		}
	}
	t.Log("status snapshot independent service comparisons", cases)
}

func TestVGAStatusSnapshotRefusesUnknownAndOverflow(t *testing.T) {
	for _, kind := range []string{"mode", "planar", "CRT"} {
		m, _ := hostPaletteFixture()
		switch kind {
		case "mode":
			m.SetVideoMode(0xd)
		case "planar":
			m.planarOn = true
		case "CRT":
			m.CRTCTiming = true
		}
		before := paletteReadState(t, m)
		if _, ok := m.VGAStatusSnapshot(); ok {
			t.Fatal("unknown status model accepted", kind)
		}
		if !bytes.Equal(before, paletteReadState(t, m)) {
			t.Fatal("refusal changed state")
		}
	}
	m, _ := hostPaletteFixture()
	m.CRTCTiming = true
	m.Mode13Phase = true
	s, ok := m.VGAStatusSnapshot()
	if !ok {
		t.Fatal("existing phase priority lost")
	}
	if _, ok = s.At(0, 1); ok {
		t.Fatal("zero moment accepted")
	}
	if _, ok = s.At(1, 0); ok {
		t.Fatal("zero read number accepted")
	}
	s.Steps = ^uint64(0)
	if _, ok = s.At(1, 1); ok {
		t.Fatal("step overflow accepted")
	}
	s.Steps = 0
	s.readTicks = ^uint64(0)
	if _, ok = s.At(1, 1); ok {
		t.Fatal("read overflow accepted")
	}
	if _, ok = (VGAStatusSnapshot{}).At(1, 1); ok {
		t.Fatal("unknown snapshot accepted")
	}
}
