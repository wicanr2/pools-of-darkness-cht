package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func paletteStageStates(t *testing.T) []*view {
	seed := paletteInitStates(t)[0]
	m := original(seed)
	for i := range m.DAC {
		m.DAC[i] = byte((i*11 + 7) % 64)
	}
	out := []*view{}
	seen := map[uint16]bool{}
	for m.Steps < 1000000 {
		ip := m.CPU.IP
		if m.CPU.Seg[cpu.CS] == 0x48a1 && (ip == 0x4b0b || ip == 0x4b8d || ip == 0x4b91 || ip == 0x4bc6) && !seen[ip] {
			seen[ip] = true
			out = append(out, &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}})
			if ip == 0x4bc6 {
				break
			}
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if len(out) != 4 {
		t.Fatal("stages", len(out), seen, m.CPU.IP)
	}
	return out
}
func TestNativePaletteStageCosts(t *testing.T) {
	for _, v := range paletteStageStates(t) {
		m := original(v)
		n := map[uint16]int{0x4b0b: 5, 0x4b8d: 2, 0x4b91: 3, 0x4bc6: 5}[v.IP()]
		for i := 0; i < n; i++ {
			if err := m.Step(); err != nil {
				t.Fatal(err)
			}
		}
		t.Logf("stage%x N%d C%d IP%x", v.IP(), m.Steps, m.CPU.Cycles, m.CPU.IP)
	}
}

func TestNativePaletteStageOriginal(t *testing.T) {
	cases := 0
	for _, seed := range paletteStageStates(t) {
		for _, count := range []uint16{1, 17, 256} {
			for _, flags := range []uint16{2, 0x2d7, 0xad7} {
				for _, dos := range []bool{false, true} {
					v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
					v.s.Flags = flags
					putRowWord(v, 0x48a10+0x116f, count)
					p := planPaletteStage(v, 100)
					if p == nil {
						t.Fatal("stage rejected", v.IP())
					}
					compare(t, v, p)
					if planPaletteStage(v, p.Instructions-1) != nil {
						t.Fatal("short budget")
					}
					compare(t, v, planPaletteStage(v, p.Instructions))
					cases += 2
				}
			}
		}
	}
	t.Log("palette stage comparisons", cases)
}

func TestNativePaletteStageRefusals(t *testing.T) {
	cases := 0
	for _, seed := range paletteStageStates(t) {
		for _, kind := range []string{"budget", "mode", "code", "CS", "stack", "data"} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			b := uint64(100)
			switch kind {
			case "budget":
				b = 0
			case "mode":
				v.mem[0x449] = 0xd
			case "code":
				ip := v.IP()
				if ip == 0x4b91 {
					ip = 0x4b8d
				}
				v.mem[0x48a10+uint32(ip)] ^= 1
			case "CS":
				v.s.Seg[cpu.CS] = 0x4800
			case "stack":
				if v.IP() == 0x4b0b || v.IP() == 0x4b91 {
					v.s.Seg[cpu.DS] = 0x4800
				} else {
					v.s.Seg[cpu.SS] = 0xa000
				}
			case "data":
				if v.IP() == 0x4bc6 {
					v.s.R[cpu.SP] = 65535
				} else {
					putRowWord(v, 0x48a10+0x116f, 257)
				}
			}
			before := append([]byte(nil), v.mem...)
			regs := v.s
			if planPaletteStage(v, b) != nil {
				t.Fatal("invalid stage admitted", seed.IP(), kind)
			}
			if string(before) != string(v.mem) || regs != v.s {
				t.Fatal("mutated")
			}
			cases++
		}
	}
	t.Log("stage refusals", cases)
}
