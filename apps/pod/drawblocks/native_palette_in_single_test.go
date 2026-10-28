package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestPaletteSingleBoundaries(t *testing.T) {
	plans := 0
	seen := map[uint16]int{}
	for _, entry := range []uint16{0x45e0, 0x4944} {
		seed := fadeInWholeSeed(t, 3, 2)
		seed.s.IP = entry
		m := fadeInMachine(seed)
		for m.Steps < 1000000 {
			if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
				break
			}
			ip := m.CPU.IP
			if m.CPU.Seg[cpu.CS] == 0x48a1 && paletteSingleWriteIP(ip) && seen[ip] < 4 {
				for _, dos := range []bool{false, true} {
					v := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: dos}
					p := planPaletteSingle(v, 1)
					if p == nil {
						t.Fatalf("single palette refused %x regs%+v", ip, v.s)
					}
					compare(t, v, p)
					plans++
				}
				seen[ip]++
			}
			if e := m.Step(); e != nil {
				t.Fatal(e)
			}
		}
	}
	for _, ip := range []uint16{0x4b82, 0x4b86, 0x48fb, 0x48ff, 0x4a6a, 0x47bd, 0x4a8d, 0x47e0, 0x4b00, 0x4b04, 0x4b08, 0x486a, 0x486e, 0x4872, 0x4b53, 0x48cc} {
		if seen[ip] == 0 {
			t.Fatalf("no palette phase %x", ip)
		}
	}
	t.Logf("palette fade-in/out one-step original comparisons=%d distinct phases=%d", plans, len(seen))
}
