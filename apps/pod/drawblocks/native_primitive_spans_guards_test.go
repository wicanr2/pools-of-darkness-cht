package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

type primitiveUnknownModel struct{ *view }

func (v primitiveUnknownModel) CPUModel() (cpu.Model, bool) { return 0, false }

type primitive8086Model struct{ *view }

func (v primitive8086Model) CPUModel() (cpu.Model, bool) { return cpu.Model8086, true }

func TestPrimitiveRefusals(t *testing.T) {
	seeds := map[uint16]*view{}
	for _, entry := range []struct {
		kind string
		xy   [4]uint16
	}{{"point", [4]uint16{17, 99, 0, 0}}, {"region", [4]uint16{17, 19, 23, 20}}, {"line", [4]uint16{17, 19, 35, 25}}, {"line", [4]uint16{35, 19, 17, 25}}, {"line", [4]uint16{17, 19, 23, 39}}, {"line", [4]uint16{23, 19, 17, 39}}} {
		v := primitiveSeed(t, entry.kind, entry.xy, 320)
		m := callMachine(v)
		for m.Steps < 100000 {
			if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
				break
			}
			if m.CPU.Seg[cpu.CS] == 0x48a1 && primitivePixelIP(m.CPU.IP) {
				if _, ok := seeds[m.CPU.IP]; !ok {
					w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}}
					if planPrimitivePixel(w, 1) == nil {
						t.Fatal("bad positive primitive seed")
					}
					seeds[m.CPU.IP] = w
				}
			}
			if e := m.Step(); e != nil {
				t.Fatal(e)
			}
		}
	}
	if len(seeds) != 7 {
		t.Fatal("missing positive phases", len(seeds))
	}
	cases := 0
	for ip, seed := range seeds {
		for _, kind := range []string{"budget", "mode", "CS", "IP", "DS", "ES", "SP", "BP", "slot", "source", "AX", "CX", "SI", "DI", "DF", "CF", "color", "stride", "root", "cycle", "unknown-model", "8086"} {
			point := ip == 0x2bfc
			region := ip == 0x345c || ip == 0x3460
			if point && (kind == "CX" || kind == "SI" || kind == "DF") {
				continue
			}
			if kind == "CF" && !region {
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
			case "SP":
				w.s.R[cpu.SP]++
			case "BP":
				w.s.R[cpu.BP]++
			case "slot":
				putRowWord(w, stack+12, 256)
			case "source":
				start := uint32(0x2cfc)
				if point {
					start = 0x2b10
				} else if region {
					start = 0x3338
				}
				w.mem[0x48a10+start] ^= 1
			case "AX":
				w.s.R[cpu.AX]++
			case "CX":
				w.s.R[cpu.CX]++
			case "SI":
				w.s.R[cpu.SI]++
			case "DI":
				w.s.R[cpu.DI]++
			case "DF":
				w.s.Flags ^= cpu.DF
			case "CF":
				w.s.Flags ^= cpu.CF
			case "color":
				off := uint32(22)
				if point {
					off = 18
				}
				putRowWord(w, stack+off, w.Read16(stack+off)+1)
			case "stride":
				putRowWord(w, 0x48a10+0xedd, 319)
			case "root":
				putRowWord(w, 0x48a10+0xbd, 0xb000)
			case "cycle":
				putRowWord(w, 0x48a10+0xbd, 0)
				putRowWord(w, 0x48a10+0xcbd, 0xc0)
			}
			before := append([]byte(nil), w.mem...)
			regs := w.s
			var input CPUView = w
			if kind == "unknown-model" {
				input = primitiveUnknownModel{w}
			}
			if kind == "8086" {
				input = primitive8086Model{w}
			}
			if planPrimitivePixel(input, budget) != nil {
				t.Fatalf("unproved primitive accepted %s IP%x", kind, ip)
			}
			if !bytes.Equal(before, w.mem) || regs != w.s {
				t.Fatal("refusal mutated input")
			}
			cases++
		}
	}
	t.Logf("primitive positive phases=%d source/frame/cursor/model refusals=%d unchanged", len(seeds), cases)
}

func TestPrimitiveParents(t *testing.T) {
	cases := 0
	inputs := []struct {
		kind string
		xy   [4]uint16
	}{{"point", [4]uint16{30, 20, 0, 0}}, {"region", [4]uint16{20, 10, 80, 45}}, {"line", [4]uint16{25, 15, 35, 40}}, {"line", [4]uint16{70, 15, 60, 40}}, {"line", [4]uint16{25, 15, 70, 35}}, {"line", [4]uint16{70, 15, 25, 35}}}
	for _, links := range []uint16{1, 3, 8} {
		for _, entry := range inputs {
			for _, dos := range []bool{false, true} {
				for _, budget := range []uint64{1, 1024} {
					v := primitiveSeed(t, entry.kind, entry.xy, 320)
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
					primitiveWhole(t, v, budget)
					cases++
				}
			}
		}
	}
	t.Logf("primitive clipped subroot/eight-parent whole SaveState+PortLog cases=%d CPU/fixed primitive pixels=0", cases)
}

func TestPrimitiveFullCanvas(t *testing.T) {
	cases := 0
	for _, dos := range []bool{false, true} {
		for _, budget := range []uint64{1, 1024} {
			v := regionSeed(t, [4]uint16{0, 0, 319, 199}, 320, false)
			v.dos = dos
			primitiveWhole(t, v, budget)
			cases++
		}
	}
	t.Logf("primitive full-canvas region SaveState+PortLog cases=%d CPU/fixed primitive pixels=0", cases)
}
