package drawblocks

import (
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func paletteSeed(t *testing.T) *view {
	t.Helper()
	for _, v := range drawCalls(t) {
		if v.CS() == 0x48a1 && v.IP() == 0x4944 {
			return v
		}
	}
	t.Fatal("original palette entry absent")
	return nil
}

func TestNativePaletteRowsOriginal(t *testing.T) {
	seed := paletteSeed(t)
	plans := 0
	for _, width := range []uint16{1, 2, 3, 17, 256, 320} {
		for _, dos := range []bool{false, true} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
			v.s.IP = 0x4a67
			v.s.Seg[cpu.DS] = 0xa000
			v.s.R[cpu.SI] = 320
			v.s.R[cpu.CX] = width
			v.s.R[cpu.BX] = 0
			v.s.Flags &^= cpu.DF
			for x := uint16(0); x < width; x++ {
				v.mem[0xa0140+int(x)] = byte(x * 11)
			}
			for _, budget := range []uint64{4, 8, 64, 1024, 1280} {
				p := planPaletteScan(v, budget)
				if p == nil {
					t.Fatal("valid row rejected")
				}
				compare(t, v, p)
				plans++
			}
		}
	}
	t.Log("palette synthetic original plan comparisons", plans)
}

func TestNativePaletteNormalStates(t *testing.T) {
	m := original(paletteSeed(t))
	var states []*view
	for m.Steps < 1000000 && len(states) < 32 {
		if m.CPU.Seg[cpu.CS] == 0x48a1 && m.CPU.IP == 0x4a67 {
			states = append(states, &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}})
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if len(states) != 32 {
		t.Fatal("original palette scan states absent")
	}
	plans := 0
	for _, v := range states {
		for _, dos := range []bool{false, true} {
			v.dos = dos
			for _, budget := range []uint64{4, 64, 1280} {
				p := planPaletteScan(v, budget)
				if p == nil {
					t.Fatal("normal palette scan rejected")
				}
				compare(t, v, p)
				plans++
			}
		}
	}
	t.Log("palette normal original plan comparisons", plans)
}

func TestNativePaletteRefusals(t *testing.T) {
	for _, kind := range []string{"budget", "code", "mode", "direction", "source", "bounds", "row", "counter", "high-byte"} {
		t.Run(kind, func(t *testing.T) {
			seed := paletteSeed(t)
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			v.s.IP = 0x4a67
			v.s.Seg[cpu.DS] = 0xa000
			v.s.R[cpu.SI] = 0
			v.s.R[cpu.CX] = 320
			v.s.R[cpu.BX] = 0
			v.s.Flags &^= cpu.DF
			budget := uint64(1280)
			switch kind {
			case "budget":
				budget = 3
			case "code":
				v.mem[0x48a10+0x4a67] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "direction":
				v.s.Flags |= cpu.DF
			case "source":
				v.s.Seg[cpu.DS] = 0x5000
			case "bounds":
				v.s.R[cpu.SI] = 63999
			case "row":
				v.s.R[cpu.SI] = 1
			case "counter":
				v.s.R[cpu.CX] = 0
			case "high-byte":
				v.s.R[cpu.BX] = 0x100
			}
			if planPaletteScan(v, budget) != nil {
				t.Fatal("unsupported scan accepted")
			}
		})
	}
}
