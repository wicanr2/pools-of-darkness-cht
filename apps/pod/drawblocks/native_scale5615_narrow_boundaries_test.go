package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNarrowScaleBoundaries(t *testing.T) {
	prefixes, prepare, output, spans, zero := 0, 0, 0, 0, 0
	seen := map[uint16]bool{}
	for _, height := range []uint16{1, 3, 9} {
		for _, factor := range []uint16{1, 2, 3, 8} {
			for _, center := range [][2]int16{{25, 15}, {-5, -3}, {70, 80}} {
				for _, dos := range []bool{false, true} {
					v := scale5615Seed(t, 1, height, factor, center[0], center[1])
					v.dos = dos
					p := planScale5615Prefix(v)
					if p == nil {
						t.Fatal("narrow prefix refused")
					}
					compare(t, v, p)
					prefixes++
					m := callMachine(v)
					for m.Steps < 100000 {
						if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
							break
						}
						if m.CPU.Seg[cpu.CS] == 0x48a1 {
							ip := m.CPU.IP
							w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}, dos: dos}
							switch ip {
							case 0x579c:
								p := planScale5615Prepare(w, 10000)
								if p == nil {
									t.Fatal("narrow prepare refused")
								}
								compare(t, w, p)
								prepare++
							case 0x581a:
								p := planScale5615Output(w, 10000)
								if p == nil {
									t.Fatal("narrow output refused")
								}
								compare(t, w, p)
								output++
							case 0x5813, 0x5879, 0x587d, 0x5856, 0x585a:
								p := planScale5615Span(w, 1)
								if p == nil {
									t.Fatal("narrow span refused", ip)
								}
								compare(t, w, p)
								spans++
								seen[ip] = true
								if w.CX() == 0 {
									zero++
								}
							}
						}
						if e := m.Step(); e != nil {
							t.Fatal(e)
						}
					}
				}
			}
		}
	}
	if len(seen) != 5 {
		t.Fatal("missing narrow phases")
	}
	t.Logf("width1 scale original prefix=%d prepare=%d output=%d single-span=%d zero-count=%d distinct phases=%d", prefixes, prepare, output, spans, zero, len(seen))
}

func TestNarrowScaleRefusals(t *testing.T) {
	seeds := map[uint16]*view{}
	for _, center := range [][2]int16{{25, 15}, {70, 80}} {
		v := scale5615Seed(t, 1, 3, 2, center[0], center[1])
		m := callMachine(v)
		for m.Steps < 100000 {
			if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
				break
			}
			if m.CPU.Seg[cpu.CS] == 0x48a1 {
				switch m.CPU.IP {
				case 0x579c, 0x581a, 0x5813, 0x5879, 0x587d, 0x5856, 0x585a:
					if _, ok := seeds[m.CPU.IP]; !ok {
						w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}}
						if scale5615(w, 10000) == nil {
							t.Fatal("bad narrow seed")
						}
						seeds[m.CPU.IP] = w
					}
				}
			}
			if e := m.Step(); e != nil {
				t.Fatal(e)
			}
		}
	}
	if len(seeds) != 7 {
		t.Fatal("missing narrow positive phases", len(seeds))
	}
	cases := 0
	for ip, seed := range seeds {
		for _, kind := range []string{"budget", "mode", "DS", "DF", "source", "SP", "BP", "samples", "x", "y", "root"} {
			w := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			budget := uint64(10000)
			switch kind {
			case "budget":
				budget = 0
			case "mode":
				w.mem[0x449] = 3
			case "DS":
				w.s.Seg[cpu.DS]++
			case "DF":
				w.s.Flags |= cpu.DF
			case "source":
				w.mem[0x48a10+0x5615] ^= 1
			case "SP":
				w.s.R[cpu.SP]++
			case "BP":
				w.s.R[cpu.BP]++
			case "samples":
				putRowWord(w, 0x48a10+0x1212, 1)
			case "x":
				putRowWord(w, 0x48a10+0x1141, 0)
			case "y":
				putRowWord(w, 0x48a10+0x1171, 300)
			case "root":
				putRowWord(w, 0x48a10+0xbd+4, 0xb000)
			}
			before := append([]byte(nil), w.mem...)
			regs := w.s
			if scale5615(w, budget) != nil {
				t.Fatalf("unproved narrow accepted %s IP%x", kind, ip)
			}
			for i, b := range before {
				if w.mem[i] != b {
					t.Fatal("refusal changed RAM")
				}
			}
			if regs != w.s {
				t.Fatal("refusal changed registers")
			}
			cases++
		}
	}
	t.Logf("width1 scale positive phases=%d source/frame/grid refusals=%d unchanged", len(seeds), cases)
}
