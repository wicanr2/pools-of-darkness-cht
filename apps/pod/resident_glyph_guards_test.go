package pod

import (
	"bytes"
	"encoding/binary"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"testing"
)

type residentUnknown struct{ residentView }

func (v residentUnknown) CPUModel() (cpu.Model, bool) { return cpu.Model80386, false }

type resident8086 struct{ residentView }

func (v resident8086) CPUModel() (cpu.Model, bool) { return cpu.Model8086, true }
func TestResidentPixelRefusals(t *testing.T) {
	seeds := map[uint16]*glyphView{}
	for _, seed := range glyphViews(t) {
		m := glyphMachine(seed, cpu.Model80386)
		for n := 0; n < 700 && len(seeds) < 2; n++ {
			if m.CPU.Seg[cpu.CS] == 0x1b2e && (m.CPU.IP == 0x11ea || m.CPU.IP == 0x11f1) {
				w := residentViewAt(m, false)
				if residentGlyphPlan(w, 1) == nil {
					t.Fatal("bad positive")
				}
				seeds[m.CPU.IP] = &glyphView{memory: append([]byte(nil), m.Mem...), before: w.before, step: m.Steps}
			}
			if e := m.Step(); e != nil {
				t.Fatal(e)
			}
		}
		if len(seeds) == 2 {
			break
		}
	}
	if len(seeds) != 2 {
		t.Fatal("need both glyph colors")
	}
	cases := 0
	for ip, seed := range seeds {
		for _, kind := range []string{"budget", "mode", "code", "CS", "IP", "ES", "DS", "SS", "SP", "BP", "AX", "BX", "CX", "DI", "SI", "DF", "CF", "PF", "SF", "OF", "row", "x", "y", "char", "color", "font-seg", "font-base", "saved-DS", "font-bit", "alias", "unknown-model", "8086"} {
			v := &glyphView{memory: append([]byte(nil), seed.memory...), before: seed.before, step: seed.step}
			put := func(at uint32, x uint16) { binary.LittleEndian.PutUint16(v.memory[at:], x) }
			stack := uint32(v.SS())*16 + uint32(v.SP())
			table := uint32(v.Read16(stack+2))*16 + 0x9d86
			budget := uint64(1)
			switch kind {
			case "budget":
				budget = 0
			case "mode":
				v.memory[0x449] = 3
			case "code":
				v.memory[0x1c486] ^= 1
			case "CS":
				v.before.Seg[cpu.CS]++
			case "IP":
				v.before.IP++
			case "ES":
				v.before.Seg[cpu.ES]++
			case "DS":
				v.before.Seg[cpu.DS]++
			case "SS":
				v.before.Seg[cpu.SS]++
			case "SP":
				v.before.R[cpu.SP]++
			case "BP":
				v.before.R[cpu.BP]++
			case "AX":
				v.before.R[cpu.AX]++
			case "BX":
				v.before.R[cpu.BX] ^= 1
			case "CX":
				v.before.R[cpu.CX] = 0
			case "DI":
				v.before.R[cpu.DI]++
			case "SI":
				v.before.R[cpu.SI]++
			case "DF":
				v.before.Flags |= cpu.DF
			case "CF":
				v.before.Flags ^= cpu.CF
			case "PF":
				v.before.Flags ^= cpu.PF
			case "SF":
				v.before.Flags ^= cpu.SF
			case "OF":
				v.before.Flags ^= cpu.OF
			case "row":
				put(stack, 0)
			case "x":
				put(stack+18, 40)
			case "y":
				put(stack+16, 25)
			case "char":
				v.memory[stack+10]++
			case "color":
				at := stack + 12
				if ip == 0x11f1 {
					at = stack + 14
				}
				v.memory[at]++
			case "font-seg":
				put(table+2, v.Read16(table+2)+1)
			case "font-base":
				put(table, v.Read16(table)+1)
			case "saved-DS":
				put(stack+2, v.Read16(stack+2)+1)
			case "font-bit":
				v.memory[uint32(v.DS())*16+uint32(v.SI())] ^= byte(v.BX())
			case "alias":
				put(table+2, v.SS())
				put(table, v.SP()-uint16(v.Read8(stack+10))*8)
				v.before.Seg[cpu.DS] = v.SS()
			}
			before := append([]byte(nil), v.memory...)
			regs := v.before
			w := residentView{v}
			var input HostCallCPU = w
			if kind == "unknown-model" {
				input = residentUnknown{w}
			}
			if kind == "8086" {
				input = resident8086{w}
			}
			if residentGlyphPlan(input, budget) != nil {
				t.Fatalf("unsafe resident accepted %s IP%x", kind, ip)
			}
			if !bytes.Equal(before, v.memory) || regs != v.before {
				t.Fatal("refusal changed input")
			}
			cases++
		}
	}
	t.Logf("resident glyph positive phases=%d source/frame/cursor/model refusals=%d unchanged", len(seeds), cases)
}
