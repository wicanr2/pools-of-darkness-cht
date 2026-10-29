package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestSurfaceFillSpanRefusals(t *testing.T) {
	seeds := map[uint16]*view{}
	v := surfaceFillSpanSeed(t, 3, 2)
	m := callMachine(v)
	for m.Steps < 100000 && len(seeds) < 2 {
		if m.CPU.Seg[cpu.CS] == 0x48a1 && surfaceFillPixelIP(m.CPU.IP) {
			if _, ok := seeds[m.CPU.IP]; !ok {
				w := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}}
				if planSurfaceFillSpan(w, 1) == nil {
					t.Fatal("bad positive seed")
				}
				seeds[m.CPU.IP] = w
			}
		}
		if e := m.Step(); e != nil {
			t.Fatal(e)
		}
	}
	if len(seeds) != 2 {
		t.Fatal("missing two positive phases")
	}
	cases := 0
	for ip, seed := range seeds {
		for _, kind := range []string{"budget", "mode", "source", "CS", "IP", "DS", "ES", "SP", "BP", "slot", "AX", "BX", "CX", "DI", "CF", "DF", "color", "scratch-left", "stride", "root", "cycle", "unknown-model", "8086"} {
			w := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			budget := uint64(1)
			stack := uint32(w.SS())*16 + uint32(w.SP())
			switch kind {
			case "budget":
				budget = 0
			case "mode":
				w.mem[0x449] = 3
			case "source":
				w.mem[0x48a10+0x41af] ^= 1
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
			case "AX":
				w.s.R[cpu.AX]++
			case "BX":
				w.s.R[cpu.BX]++
			case "CX":
				w.s.R[cpu.CX]++
			case "DI":
				w.s.R[cpu.DI]++
			case "CF":
				w.s.Flags ^= cpu.CF
			case "DF":
				w.s.Flags |= cpu.DF
			case "color":
				putRowWord(w, stack+14, w.Read16(stack+14)+1)
			case "scratch-left":
				w.mem[0x48a10+0x113b] ^= 1
			case "stride":
				putRowWord(w, 0x48a10+0xedd, 319)
			case "root":
				putRowWord(w, 0x48a10+0xbd+2, 0xb000)
			case "cycle":
				putRowWord(w, 0x48a10+0xbd+2, 2)
				putRowWord(w, 0x48a10+0xcbd+2, 0xc0)
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
			if planSurfaceFillSpan(input, budget) != nil {
				t.Fatalf("unproved surface fill accepted %s IP%x", kind, ip)
			}
			if !bytes.Equal(before, w.mem) || regs != w.s {
				t.Fatal("refusal changed input")
			}
			cases++
		}
	}
	t.Logf("surface-fill span positive phases=%d source/frame/cursor/model refusals=%d unchanged", len(seeds), cases)
}
func TestSurfaceFillSpanParents(t *testing.T) {
	cases := 0
	for _, links := range []uint16{1, 3, 8} {
		for _, dos := range []bool{false, true} {
			for _, budget := range []uint64{1, 7, 1024} {
				v := surfaceFillSpanSeed(t, 3, 2)
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
				a, b := callMachine(v), callMachine(v)
				runCall(t, a, v, false)
				for b.Steps < 1000000 {
					if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
						break
					}
					w := paletteMachineView{b}
					p := Plan(w, budget)
					if p != nil {
						if !p.Algorithmic && surfaceFillFixedPixels(p.Writes, p.Transfers) {
							t.Fatal("fixed surface-fill parent")
						}
						if e := b.CommitHostBlock(p); e != nil {
							t.Fatal(e)
						}
					} else {
						if w.CS() == 0x48a1 && surfaceFillPixelIP(w.IP()) {
							t.Fatalf("CPU surface-fill parent IP%x", w.IP())
						}
						if e := b.Step(); e != nil {
							t.Fatal(e)
						}
					}
				}
				paletteMachinesEqual(t, a, b)
				cases++
			}
		}
	}
	t.Logf("surface-fill nonzero root/eight-parent SaveState+PortLog cases=%d CPU/fixed surface-fill pixels=0", cases)
}
