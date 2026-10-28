package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func surfaceSpanSeed(t *testing.T, width, height uint16, bx, by, dos bool) *view {
	v := copySurfaceSeed(t, width, height, bx, by)
	v.dos = dos
	v.s.Seg[cpu.SS] = 0x8000
	v.s.R[cpu.SP] = 0xe000
	for off, value := range map[uint32]uint16{0: 0x1234, 2: 0x4567, 4: 0, 6: 1} {
		putRowWord(v, 0x8e000+off, value)
	}
	return v
}

func TestSurfaceSpanOriginalBoundaries(t *testing.T) {
	comparisons, zero := 0, 0
	for _, width := range []uint16{1, 2, 3, 8, 19} {
		for _, height := range []uint16{1, 3} {
			for _, bx := range []bool{false, true} {
				for _, by := range []bool{false, true} {
					for _, dos := range []bool{false, true} {
						v := surfaceSpanSeed(t, width, height, bx, by, dos)
						m := callMachine(v)
						for m.Steps < 100000 {
							if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
								break
							}
							if m.CPU.Seg[cpu.CS] == 0x48a1 && (m.CPU.IP == 0x4169 || m.CPU.IP == 0x4177) {
								w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: dos}
								p := planSurfaceCopySpan(w, 1)
								if p == nil {
									t.Fatalf("span rejected width%d height%d bx%v by%v IP%x flags%x", width, height, bx, by, w.IP(), w.Flags())
								}
								compare(t, w, p)
								comparisons++
								if w.CX() == 0 {
									zero++
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
	}
	t.Logf("surface span independently reached ordered original comparisons=%d zero-count=%d", comparisons, zero)
}

type surfaceUnknownModel struct{ *view }

func (v surfaceUnknownModel) CPUModel() (cpu.Model, bool) { return 0, false }

type surface8086Model struct{ *view }

func (v surface8086Model) CPUModel() (cpu.Model, bool) { return cpu.Model8086, true }

func TestSurfaceSpanRefusals(t *testing.T) {
	cases := 0
	for _, bx := range []bool{false, true} {
		for _, by := range []bool{false, true} {
			seed := surfaceSpanSeed(t, 3, 2, bx, by, false)
			m := callMachine(seed)
			seen := map[uint16]bool{}
			for m.Steps < 100000 && len(seen) < 2 {
				if m.CPU.Seg[cpu.CS] == 0x48a1 && (m.CPU.IP == 0x4169 || m.CPU.IP == 0x4177) && !seen[m.CPU.IP] {
					state := machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}
					seen[m.CPU.IP] = true
					for _, kind := range []string{"budget", "mode", "CS", "IP", "DS", "ES", "DF", "source", "AX", "BX", "CX", "SI", "DI", "SP", "BP", "width", "remaining", "direction", "root-stride", "surface", "flags", "unknown-model", "8086"} {
						w := &view{mem: append([]byte(nil), m.Mem...), s: state}
						budget := uint64(1)
						switch kind {
						case "budget":
							budget = 0
						case "mode":
							w.mem[0x449] = 3
						case "CS":
							w.s.Seg[cpu.CS]--
						case "IP":
							w.s.IP++
						case "DS":
							w.s.Seg[cpu.DS]--
						case "ES":
							w.s.Seg[cpu.ES]--
						case "DF":
							w.s.Flags ^= cpu.DF
						case "source":
							w.mem[0x48a10+0x3f5e] ^= 1
						case "AX":
							w.s.R[cpu.AX]++
						case "BX":
							w.s.R[cpu.BX]++
						case "CX":
							w.s.R[cpu.CX]++
						case "SI":
							w.s.R[cpu.SI]++
						case "DI":
							w.s.R[cpu.DI]++
						case "SP":
							w.s.R[cpu.SP]++
						case "BP":
							w.s.R[cpu.BP]++
						case "width":
							putRowWord(w, 0x48a10+0x1135, 0)
						case "remaining":
							putRowWord(w, 0x48a10+0x1133, 0)
						case "direction":
							putRowWord(w, 0x48a10+0x1147, 0)
						case "root-stride":
							putRowWord(w, 0x48a10+0xedd, 319)
						case "surface":
							putRowWord(w, 0x48a10+0xec1, 0)
						case "flags":
							if w.IP() == 0x4169 {
								w.s.Flags ^= cpu.CF
							} else {
								w.s.Flags ^= cpu.PF
							}
						}
						before := append([]byte(nil), w.mem...)
						regs := w.s
						var input CPUView = w
						if kind == "unknown-model" {
							input = surfaceUnknownModel{w}
						}
						if kind == "8086" {
							input = surface8086Model{w}
						}
						if p := planSurfaceCopySpan(input, budget); p != nil {
							t.Fatalf("unproved surface span accepted %s bx%v by%v IP%x", kind, bx, by, w.IP())
						}
						if !bytes.Equal(before, w.mem) || regs != w.s {
							t.Fatalf("refusal mutated input %s", kind)
						}
						cases++
					}
				}
				if e := m.Step(); e != nil {
					t.Fatal(e)
				}
			}
			if len(seen) != 2 {
				t.Fatal("missing original write boundaries")
			}
		}
	}
	t.Logf("surface span source/frame/flags/model refusals=%d unchanged", cases)
}
