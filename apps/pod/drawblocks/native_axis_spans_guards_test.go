package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

type axisUnknownModel struct{ *view }

func (v axisUnknownModel) CPUModel() (cpu.Model, bool) { return 0, false }

type axis8086Model struct{ *view }

func (v axis8086Model) CPUModel() (cpu.Model, bool) { return cpu.Model8086, true }

func TestAxisSpanRefusals(t *testing.T) {
	seeds := map[uint16]*view{}
	for _, xy := range [][4]uint16{{2, 3, 8, 3}, {17, 0, 17, 20}} {
		v := axisSeed(t, xy, 320)
		m := callMachine(v)
		for m.Steps < 100000 {
			if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
				break
			}
			if m.CPU.Seg[cpu.CS] == 0x48a1 && axisSpanPixelIP(m.CPU.IP) {
				if _, ok := seeds[m.CPU.IP]; !ok {
					w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}}
					if planAxisSpan(w, 1) == nil {
						t.Fatal("bad positive seed")
					}
					seeds[m.CPU.IP] = w
				}
			}
			if e := m.Step(); e != nil {
				t.Fatal(e)
			}
		}
	}
	if len(seeds) != 10 {
		t.Fatal("missing positive phases")
	}
	cases := 0
	for ip, seed := range seeds {
		for _, kind := range []string{"budget", "mode", "CS", "IP", "DS", "ES", "DF", "source", "SP", "BP", "slot", "AX", "CX", "DI", "color", "scratch-x", "scratch-y", "root-stride", "unknown-model", "8086", "CF"} {
			if kind == "CF" && ip != 0x2e6d && ip != 0x2e71 {
				continue
			}
			w := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			budget := uint64(1)
			stack := uint32(w.SS())*16 + uint32(w.SP())
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
				w.s.Flags |= cpu.DF
			case "source":
				w.mem[0x48a10+0x2cfc] ^= 1
			case "SP":
				w.s.R[cpu.SP]++
			case "BP":
				w.s.R[cpu.BP]++
			case "slot":
				putRowWord(w, stack+12, 256)
			case "AX":
				w.s.R[cpu.AX]++
			case "CX":
				w.s.R[cpu.CX]++
			case "DI":
				w.s.R[cpu.DI]++
			case "color":
				putRowWord(w, stack+22, w.Read16(stack+22)+256)
			case "scratch-x":
				w.mem[0x48a10+0x112b] ^= 1
			case "scratch-y":
				w.mem[0x48a10+0x112f] ^= 1
			case "root-stride":
				putRowWord(w, 0x48a10+0xedd, 319)
			case "CF":
				w.s.Flags ^= cpu.CF
			}
			before := append([]byte(nil), w.mem...)
			regs := w.s
			var input CPUView = w
			if kind == "unknown-model" {
				input = axisUnknownModel{w}
			}
			if kind == "8086" {
				input = axis8086Model{w}
			}
			if planAxisSpan(input, budget) != nil {
				t.Fatalf("unproved axis accepted %s IP%x", kind, ip)
			}
			if !bytes.Equal(before, w.mem) || regs != w.s {
				t.Fatal("refusal mutated input")
			}
			cases++
		}
	}
	t.Logf("axis span positive phases=%d source/frame/cursor/model refusals=%d unchanged", len(seeds), cases)
}

func TestAxisSpanParents(t *testing.T) {
	cases := 0
	for _, links := range []uint16{1, 3, 8} {
		for _, xy := range [][4]uint16{{0, 20, 200, 20}, {70, 25, 25, 25}, {30, 0, 30, 60}, {30, 40, 30, 15}, {30, 15, 30, 22}} {
			for _, dos := range []bool{false, true} {
				for _, budget := range []uint64{1, 1024} {
					v := axisSeed(t, xy, 320)
					v.dos = dos
					for i := uint16(0); i <= links; i++ {
						seg, flags := uint16(0xa000), uint16(0x80)
						left, top, right, bottom := uint16(20), uint16(10), uint16(100), uint16(60)
						if i < links {
							seg, flags = (i+1)*2, 0xc0
							left, top, right, bottom = 25, 15, 70, 40
						}
						for off, value := range map[uint16]uint16{0xbd: seg, 0x2bd: 0, 0x4bd: left, 0x6bd: top, 0x8bd: right, 0xabd: bottom, 0xcbd: flags} {
							putRowWord(v, 0x48a10+uint32(off+i*2), value)
						}
					}
					axisWhole(t, v, budget)
					cases++
				}
			}
		}
	}
	t.Logf("axis span clipped subroot/eight-parent whole SaveState+PortLog cases=%d CPU/fixed axis pixels=0", cases)
}
