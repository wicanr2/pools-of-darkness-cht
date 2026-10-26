package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeFilledEllipseSpanRefusals(t *testing.T) {
	v := ellipseSeed(t, 160, 100, 10, 20, [4]uint16{0, 0, 319, 199})
	v.s.IP = 0x39b2
	m := original(v)
	states := map[uint16]*view{}
	for m.Steps < 100000 {
		ip := m.CPU.IP
		if m.CPU.Seg[cpu.CS] == 0x48a1 && (ip == 0x3bd8 || ip == 0x3bdc) && states[ip] == nil {
			states[ip] = &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}}
		}
		if len(states) == 2 {
			break
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	cases := 0
	for ip, seed := range states {
		for _, kind := range []string{"CX", "DI", "AX"} {
			w := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			switch kind {
			case "CX":
				w.s.R[cpu.CX]++
			case "DI":
				w.s.R[cpu.DI]++
			case "AX":
				w.s.R[cpu.AX]++
			}
			before := append([]byte(nil), w.mem...)
			regs := w.s
			if planFilledEllipseSpan(w, 1) != nil {
				t.Fatal("invalid span accepted", ip, kind)
			}
			if !bytes.Equal(before, w.mem) || regs != w.s {
				t.Fatal("span refusal changed state")
			}
			cases++
		}
	}
	if cases != 6 {
		t.Fatal("span refusal states absent")
	}
	t.Log("filled ellipse count/address/color refusals", cases)
}
func TestNativeFilledEllipseParents(t *testing.T) {
	cases := 0
	for _, links := range []uint16{1, 3} {
		for _, dos := range []bool{false, true} {
			v := ellipseSeed(t, 50, 40, 40, 20, [4]uint16{25, 15, 70, 150})
			v.s.IP = 0x39b2
			v.dos = dos
			for i := uint16(1); i <= links; i++ {
				seg, flags := uint16(0xa000), uint16(0x80)
				if i < links {
					seg, flags = (i+1)*2, 0xc0
				}
				for _, w := range [][2]uint16{{0xbd, seg}, {0x2bd, 0}, {0x4bd, 0}, {0x6bd, 0}, {0x8bd, 319}, {0xabd, 199}, {0xcbd, flags}} {
					putRowWord(v, 0x48a10+uint32(w[0]+2*i), w[1])
				}
			}
			p := planFilledEllipsePrefix(v)
			if p == nil {
				t.Fatal("filled ellipse parent prefix refused")
			}
			compare(t, v, p)
			cases++
			a, b := callMachine(v), callMachine(v)
			runCall(t, a, v, false)
			runCall(t, b, v, true)
			var sa, sb bytes.Buffer
			if err := a.SaveState(&sa); err != nil {
				t.Fatal(err)
			}
			if err := b.SaveState(&sb); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(sa.Bytes(), sb.Bytes()) {
				t.Fatal("filled ellipse parent machine differs", links, dos)
			}
			cases++
		}
	}
	t.Log("filled ellipse parent prefix and machine comparisons", cases)
}
