package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"reflect"
	"testing"
)

func fadeInWholeSeed(t *testing.T, width, height uint16) *view {
	v := copySurfaceSeed(t, width, height, false, false)
	v.s.IP = 0x45e0
	v.s.Seg[cpu.SS] = 0x8000
	v.s.R[cpu.SP] = 0xe000
	for off, value := range map[uint32]uint16{0: 0x1234, 2: 0x4567, 4: 0, 6: 1} {
		putRowWord(v, 0x8e000+off, value)
	}
	return v
}
func fadeInMachine(v *view) *machine.Machine {
	m := callMachine(v)
	m.Out8(0x3c8, 0)
	for i := 0; i < 256; i++ {
		m.Out8(0x3c9, byte(i*3%64))
		m.Out8(0x3c9, byte(i*5%64))
		m.Out8(0x3c9, byte(i*7%64))
	}
	m.PortLog = nil
	return m
}
func TestFadeInWhole(t *testing.T) {
	cases, goPlans, fixed, cpuSteps := 0, 0, 0, 0
	families := map[uint16]int{}
	for _, dims := range [][2]uint16{{1, 1}, {3, 2}, {7, 5}} {
		for _, dos := range []bool{false, true} {
			v := fadeInWholeSeed(t, dims[0], dims[1])
			v.dos = dos
			a, b := fadeInMachine(v), fadeInMachine(v)
			runCall(t, a, v, false)
			if a.Mem[0x48a10+0x1173] == 0 {
				t.Fatal("colorful fade oracle not exercised")
			}
			for b.Steps < 1000000 {
				if b.Steps > 0 && b.CPU.Seg[cpu.CS] == 0x4567 && b.CPU.IP == 0x1234 {
					break
				}
				w := &view{mem: b.Mem, s: machine.HostRegisters{R: b.CPU.R, Seg: b.CPU.Seg, IP: b.CPU.IP, Flags: b.CPU.Flags}, dos: dos}
				p := Plan(w, 1024)
				if p != nil {
					if p.Algorithmic {
						goPlans++
						families[w.IP()]++
					} else {
						fixed++
					}
					if e := b.CommitHostBlock(p); e != nil {
						t.Fatal(e)
					}
				} else {
					if e := b.Step(); e != nil {
						t.Fatal(e)
					}
					cpuSteps++
				}
			}
			var sa, sb bytes.Buffer
			if e := a.SaveState(&sa); e != nil {
				t.Fatal(e)
			}
			if e := b.SaveState(&sb); e != nil {
				t.Fatal(e)
			}
			if !bytes.Equal(sa.Bytes(), sb.Bytes()) || !reflect.DeepEqual(a.PortLog, b.PortLog) {
				t.Fatalf("fade-in whole differs dims%v dos%v steps%d/%d IP%x ports%d/%d", dims, dos, a.Steps, b.Steps, b.CPU.IP, len(a.PortLog), len(b.PortLog))
			}
			cases++
		}
	}
	for _, entry := range []uint16{0x47ba, 0x47d6, 0x4867, 0x4892, 0x48b3, 0x48e5} {
		if families[entry] == 0 {
			t.Fatalf("whole did not use Go core %x", entry)
		}
	}
	t.Logf("fade-in whole colored SaveState+PortLog cases=%d Go plans=%d fixed control=%d CPU control=%d", cases, goPlans, fixed, cpuSteps)
}
