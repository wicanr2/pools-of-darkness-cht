// Spec025/#49: full original-machine and negative controls.
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func boundedMemoryPlanTest(v CPUView, budget uint64) *machine.HostWorkPlan {
	if p := planBoundedZero(v, budget); p != nil {
		return p
	}
	if p := planBoundedCopy(v, budget); p != nil {
		return p
	}
	return planBoundedRLE(v, budget)
}
func wholeBoundedMemoryTest(t *testing.T, v *view, budget uint64, done func(*machine.Machine) bool) int {
	t.Helper()
	a, b := callMachine(v), callMachine(v)
	plans := 0
	for n := 0; n < 300; n++ {
		if done(a) {
			break
		}
		if err := a.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if !done(a) {
		t.Fatal("oracle did not finish")
	}
	for n := 0; n < 300; n++ {
		if done(b) {
			break
		}
		w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: v.dos}
		p := Plan(w, budget)
		if p == nil || !p.Algorithmic {
			t.Fatalf("CPU/fixed fallback %04x:%04x", w.CS(), w.IP())
		}
		if err := b.CommitHostBlock(p); err != nil {
			t.Fatal(err)
		}
		plans++
	}
	if !done(b) {
		t.Fatal("candidate did not finish")
	}
	var sa, sb bytes.Buffer
	if err := a.SaveState(&sa); err != nil {
		t.Fatal(err)
	}
	if err := b.SaveState(&sb); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
		t.Fatalf("whole SaveState differs IP%04x budget%d dos%v steps%d/%d", v.IP(), budget, v.dos, a.Steps, b.Steps)
	}
	return plans
}
func TestNativeBoundedMemoryWholeMachine(t *testing.T) {
	mem := input(t)
	cases, plans := 0, 0
	for _, entry := range []uint16{0x46e, 0x1a5c} {
		for _, count := range []uint16{0, 1, 2, 3, 31, 255} {
			for _, src := range []uint16{8, 128} {
				for _, dos := range []bool{false, true} {
					for _, budget := range []uint64{1, 2, 3, 7, 17} {
						cs := uint16(0x1d1f)
						if entry == 0x1a5c {
							cs = 0x1ed4
						}
						v := &view{mem: append([]byte(nil), mem...), dos: dos, s: machine.HostRegisters{R: [8]uint16{0xab77, 23, 45, 12, 0xe000, 0xd000, 0x440, 0x880}, Seg: [4]uint16{0x6000, cs, 0x8000, 0x7000}, IP: entry, Flags: 0xfed7}}
						for off, val := range map[uint32]uint16{0: 0x1234, 2: 0x1111, 4: count, 6: 32, 8: 0xa000, 10: src, 12: 0xa000} {
							a := uint32(v.SS())*16 + uint32(v.SP()) + off
							v.mem[a] = byte(val)
							v.mem[a+1] = byte(val >> 8)
						}
						plans += wholeBoundedMemoryTest(t, v, budget, func(m *machine.Machine) bool { return m.CPU.Seg[cpu.CS] == 0x1111 && m.CPU.IP == 0x1234 })
						cases++
					}
				}
			}
		}
	}
	t.Logf("memory whole SaveState cases=%d Go plans=%d CPU/fixed pixel fallback=0", cases, plans)
}
func TestNativeBoundedRLEWholeMachine(t *testing.T) {
	mem := input(t)
	cases, plans := 0, 0
	for _, l := range boundedRLELayouts() {
		for _, n := range []int{1, 2, 3, 127, 128} {
			for _, repeat := range []bool{false, true} {
				for _, dos := range []bool{false, true} {
					for _, budget := range []uint64{1, 2, 3, 7, 17} {
						stream := []byte{byte((n - 1) * 2)}
						if repeat {
							stream[0]++
							stream = append(stream, 0xa5)
						} else {
							for i := 0; i < n; i++ {
								stream = append(stream, byte(i*11))
							}
						}
						// A second packet also verifies nonterminal branches and reconstruction.
						stream = append(stream, 1, 0xf3)
						v := &view{mem: append([]byte(nil), mem...), dos: dos, s: machine.HostRegisters{R: [8]uint16{0x3700, 0x7f, 45, uint16(0x100 + len(stream)), 0xe000, 0xd000, 0x100, 0x1234}, Seg: [4]uint16{0xa000, 0x48a1, 0x8000, 0x5000}, IP: l.entry, Flags: 0xfad7}}
						copy(v.mem[0x50100:], stream)
						plans += wholeBoundedMemoryTest(t, v, budget, func(m *machine.Machine) bool { return m.CPU.IP == l.header })
						cases++
					}
				}
			}
		}
	}
	t.Logf("RLE whole SaveState cases=%d Go plans=%d CPU/fixed pixel fallback=0", cases, plans)
}
