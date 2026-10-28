package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func paletteJob(v *view, budget uint64) *machine.HostWorkPlan {
	switch v.IP() {
	case 0x47ba:
		return planPaletteScan(v, budget)
	case 0x47d6:
		return planPaletteList(v, budget)
	case 0x4867:
		return planPaletteTarget(v, budget)
	case 0x4892:
		return planFadeDifferenceLayout(v, budget, fadeInRGBLayout)
	case 0x48b3:
		return planPaletteError(v, budget)
	case 0x48e5:
		return planFadeBlendLayout(v, budget, fadeInRGBLayout)
	}
	return planPaletteSingle(v, budget)
}

func TestPaletteRefusals(t *testing.T) {
	seeds := map[uint16]*view{}
	for _, entry := range []uint16{0x45e0, 0x4944} {
		v := fadeInWholeSeed(t, 3, 2)
		v.s.IP = entry
		m := fadeInMachine(v)
		for m.Steps < 1000000 {
			if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
				break
			}
			if m.CPU.Seg[cpu.CS] == 0x48a1 {
				ip := m.CPU.IP
				_, exists := seeds[ip]
				if !exists && (paletteSingleWriteIP(ip) || ip == 0x47ba || ip == 0x47d6 || ip == 0x4867 || ip == 0x4892 || ip == 0x48b3 || ip == 0x48e5) {
					w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}}
					if paletteJob(w, 4096) == nil {
						t.Fatalf("invalid positive seed %x", ip)
					}
					seeds[ip] = w
				}
			}
			if err := m.Step(); err != nil {
				t.Fatal(err)
			}
		}
	}
	if len(seeds) != 22 {
		t.Fatalf("expected 22 independently reached palette boundaries, got %d", len(seeds))
	}
	refusals := 0
	for ip, seed := range seeds {
		for _, kind := range []string{"mode", "CS", "DF", "budget", "source"} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			budget := uint64(4096)
			switch kind {
			case "mode":
				v.mem[0x449] = 3
			case "CS":
				v.s.Seg[cpu.CS]--
			case "DF":
				v.s.Flags |= cpu.DF
			case "budget":
				budget = 0
			case "source":
				start := ip
				switch ip {
				case 0x4b82, 0x4b86:
					start = 0x4b6c
				case 0x48fb, 0x48ff:
					start = 0x48e5
				case 0x4a6a:
					start = 0x4a67
				case 0x47bd:
					start = 0x47ba
				case 0x4a8d:
					start = 0x4a83
				case 0x47e0:
					start = 0x47d6
				case 0x4b00, 0x4b04, 0x4b08:
					start = 0x4afd
				case 0x486a, 0x486e, 0x4872:
					start = 0x4867
				case 0x4b53:
					start = 0x4b3a
				case 0x48cc:
					start = 0x48b3
				}
				v.mem[0x48a10+uint32(start)] ^= 1
			}
			before := append([]byte(nil), v.mem...)
			regs := v.s
			if p := paletteJob(v, budget); p != nil {
				t.Fatalf("unproved palette accepted %x %s", ip, kind)
			}
			if !bytes.Equal(before, v.mem) || regs != v.s {
				t.Fatalf("palette refusal mutated input %x %s", ip, kind)
			}
			refusals++
		}
	}
	t.Logf("palette independently reached boundaries=%d source/mode/DF/budget refusals=%d unchanged", len(seeds), refusals)
}

func TestPaletteZeroDistance(t *testing.T) {
	for _, entry := range []uint16{0x45e0, 0x4944} {
		for _, dos := range []bool{false, true} {
			v := fadeInWholeSeed(t, 3, 2)
			v.s.IP = entry
			v.dos = dos
			a, b := callMachine(v), callMachine(v)
			runCall(t, a, v, false)
			for b.Steps < 1000000 {
				if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
					break
				}
				w := paletteMachineView{b}
				p := Plan(w, 1)
				if p != nil {
					if err := b.CommitHostBlock(p); err != nil {
						t.Fatal(err)
					}
				} else {
					if w.CS() == 0x48a1 && (paletteSingleWriteIP(w.IP()) || w.IP() == 0x2ab2 || w.IP() == 0x2ab7 || w.IP() == 0x2abb || w.IP() == 0x2abf) {
						t.Fatalf("zero palette CPU write %x", w.IP())
					}
					if err := b.Step(); err != nil {
						t.Fatal(err)
					}
				}
			}
			paletteMachinesEqual(t, a, b)
		}
	}
	t.Log("palette zero-distance whole SaveState+PortLog cases=4 minimum budget=1")
}
