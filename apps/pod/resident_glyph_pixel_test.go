package pod

import (
	"bytes"
	"github.com/wicanr2/dosgolem/apps/pod/drawblocks"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"reflect"
	"testing"
)

type residentView struct{ *glyphView }

func (v residentView) CPUModel() (cpu.Model, bool) { return cpu.Model80386, true }
func residentViewAt(m *machine.Machine, dos bool) residentView {
	return residentView{&glyphView{memory: m.Mem, before: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, step: m.Steps, dos: dos}}
}
func residentEqual(t *testing.T, a, b *machine.Machine) {
	t.Helper()
	var sa, sb bytes.Buffer
	if e := a.SaveState(&sa); e != nil {
		t.Fatal(e)
	}
	if e := b.SaveState(&sb); e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(sa.Bytes(), sb.Bytes()) || !reflect.DeepEqual(a.PortLog, b.PortLog) {
		t.Fatalf("resident state differs IP%x/%x steps%d/%d cycles%d/%d", a.CPU.IP, b.CPU.IP, a.Steps, b.Steps, a.CPU.Cycles, b.CPU.Cycles)
	}
}
func TestResidentPixelBoundaries(t *testing.T) {
	cases := 0
	phases := map[uint16]int{}
	for _, seed := range glyphViews(t) {
		for _, dos := range []bool{false, true} {
			seed.dos = dos
			m := glyphMachine(seed, cpu.Model80386)
			ip := seed.Read16(uint32(seed.SS())*16 + uint32(seed.SP()))
			cs := seed.Read16(uint32(seed.SS())*16 + uint32(seed.SP()+2))
			for n := 0; n < 1000; n++ {
				if n > 0 && m.CPU.IP == ip && m.CPU.Seg[cpu.CS] == cs {
					break
				}
				if m.CPU.Seg[cpu.CS] == 0x1b2e && (m.CPU.IP == 0x11ea || m.CPU.IP == 0x11f1) {
					live := residentViewAt(m, dos)
					v := &glyphView{memory: append([]byte(nil), m.Mem...), before: live.before, step: m.Steps, dos: dos}
					p := residentGlyphPlan(residentView{v}, 1)
					if p == nil {
						t.Fatalf("resident pixel rejected IP%x regs%v flags%x", m.CPU.IP, m.CPU.R, m.CPU.Flags)
					}
					a, b := glyphMachine(v, cpu.Model80386), glyphMachine(v, cpu.Model80386)
					var writes []machine.HostWrite
					a.WatchWrite(0, 0xfffff, func(w *machine.Machine, addr uint32, old, value byte) {
						cs, ip := w.CPU.OpAddr()
						writes = append(writes, machine.HostWrite{Address: addr, Value: value, Step: w.Steps - v.step, CS: cs, IP: ip})
					})
					if e := a.Step(); e != nil {
						t.Fatal(e)
					}
					if e := b.CommitHostWork(p); e != nil {
						t.Fatal(e)
					}
					if a.CPU.R != b.CPU.R || a.CPU.Seg != b.CPU.Seg || a.CPU.IP != b.CPU.IP || a.CPU.Flags != b.CPU.Flags || a.CPU.Hi != b.CPU.Hi || a.CPU.Cycles != b.CPU.Cycles || a.Steps != b.Steps || !bytes.Equal(a.Mem, b.Mem) || !reflect.DeepEqual(writes, p.Writes) {
						t.Fatal("resident ordered pixel differs")
					}
					cases++
					phases[m.CPU.IP]++
				}
				if e := m.Step(); e != nil {
					t.Fatal(e)
				}
			}
		}
	}
	if cases != 4096 || len(phases) != 2 {
		t.Fatal("missing ordered pixels", cases, phases)
	}
	t.Logf("resident glyph ordered original pixel comparisons=%d phases=%d", cases, len(phases))
}
func TestResidentPixelWhole(t *testing.T) {
	cases, goPixels, cpuControl, whole := 0, 0, 0, 0
	for _, seed := range glyphViews(t) {
		for _, dos := range []bool{false, true} {
			for _, budget := range []uint64{1, 7, 1024} {
				seed.dos = dos
				a, b := glyphMachine(seed, cpu.Model80386), glyphMachine(seed, cpu.Model80386)
				ip := seed.Read16(uint32(seed.SS())*16 + uint32(seed.SP()))
				cs := seed.Read16(uint32(seed.SS())*16 + uint32(seed.SP()+2))
				for n := 0; n < 1000; n++ {
					if n > 0 && a.CPU.IP == ip && a.CPU.Seg[cpu.CS] == cs {
						break
					}
					if e := a.Step(); e != nil {
						t.Fatal(e)
					}
				}
				returned := false
				for n := 0; n < 1000; n++ {
					if n > 0 && b.CPU.IP == ip && b.CPU.Seg[cpu.CS] == cs {
						returned = true
						break
					}
					v := residentViewAt(b, dos)
					p := residentGlyphPlan(v, budget)
					if p == nil {
						p = PlanGlyphCall(v, budget)
					}
					if p != nil {
						if p.Instructions == 1 {
							goPixels++
						} else {
							whole++
						}
						if e := b.CommitHostWork(p); e != nil {
							t.Fatal(e)
						}
					} else {
						if v.CS() == 0x1b2e && (v.IP() == 0x11ea || v.IP() == 0x11f1) {
							t.Fatal("CPU resident pixel")
						}
						if e := b.Step(); e != nil {
							t.Fatal(e)
						}
						cpuControl++
					}
				}
				if !returned {
					t.Fatal("no resident return")
				}
				residentEqual(t, a, b)
				cases++
			}
		}
	}
	t.Logf("resident glyph whole SaveState+PortLog cases=%d Go pixels=%d full glyph plans=%d CPU control=%d CPU resident pixels=0", cases, goPixels, whole, cpuControl)
}

func residentGlyphPlan(v HostCallCPU, budget uint64) *machine.HostWorkPlan {
	return drawblocks.Plan(v, budget)
}
