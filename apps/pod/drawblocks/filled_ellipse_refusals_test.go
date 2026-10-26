package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestNativeFilledEllipseRefusals(t *testing.T) {
	v := ellipseSeed(t, 160, 100, 10, 20, [4]uint16{0, 0, 319, 199})
	v.s.IP = 0x39b2
	m := original(v)
	states := map[uint16]*view{}
	for m.Steps < 100000 {
		if m.Steps > 0 && m.CPU.Seg[cpu.CS] == 0x4567 && m.CPU.IP == 0x1234 {
			break
		}
		ip := m.CPU.IP
		if m.CPU.Seg[cpu.CS] == 0x48a1 && (ip == 0x3b3d || ip == 0x3cc0 || ip == 0x3bd8) {
			if states[ip] == nil {
				states[ip] = &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}}
			}
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if len(states) != 3 {
		t.Fatal("ellipse refusal states absent")
	}
	cases := 0
	for ip, seed := range states {
		for _, kind := range []string{"DF", "budget", "mode", "CS", "IP", "DS", "ES", "code", "SP", "BP", "stack-video", "slot", "center", "color", "radius", "surface-table", "root-cycle", "scratch-color", "gradient", "error"} {
			w := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			budget := uint64(1000)
			stack := uint32(w.SS())*16 + uint32(w.SP())
			shift := uint32(0)
			if ip == 0x3bd8 {
				shift = 0
			}
			switch kind {
			case "DF":
				w.s.Flags ^= cpu.DF
			case "budget":
				budget = 0
			case "mode":
				w.mem[0x449] = 0xd
			case "CS":
				w.s.Seg[cpu.CS]++
			case "IP":
				w.s.IP++
			case "DS":
				w.s.Seg[cpu.DS]--
			case "ES":
				w.s.Seg[cpu.ES]--
			case "code":
				w.mem[0x48a10+0x39b2] ^= 1
			case "SP":
				w.s.R[cpu.SP] = 65535
			case "BP":
				w.s.R[cpu.BP]++
			case "stack-video":
				w.s.Seg[cpu.SS] = 0xa000
			case "slot":
				putRowWord(w, stack+12+shift, 256)
			case "center":
				putRowWord(w, stack+14+shift, w.Read16(stack+14+shift)+1)
			case "color":
				putRowWord(w, stack+22+shift, w.Read16(stack+22+shift)^1)
			case "radius":
				putRowWord(w, stack+18+shift, 0)
			case "surface-table":
				putRowWord(w, 0x48a10+0x8bd, 0)
			case "root-cycle":
				putRowWord(w, 0x48a10+0xbd, 0)
			case "scratch-color":
				putRowWord(w, 0x48a10+0x1129, 0)
			case "gradient":
				putRowWord(w, 0x48a10+0x114d, 65535)
			case "error":
				putRowWord(w, 0x48a10+0x1149, w.Read16(0x48a10+0x1149)^1)
			}
			before := append([]byte(nil), w.mem...)
			regs := w.s
			if planFilledEllipseBounded(w, budget) != nil {
				t.Fatal("ellipse invalid accepted", ip, kind)
			}
			if !bytes.Equal(before, w.mem) || regs != w.s {
				t.Fatal("ellipse refusal changed state")
			}
			cases++
		}
	}
	t.Log("filled ellipse cursor and span refusals", cases)
}
