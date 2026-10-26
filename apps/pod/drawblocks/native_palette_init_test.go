package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func paletteInitStates(t *testing.T) []*view {
	var seed *view
	for _, v := range drawCalls(t) {
		if v.CS() == 0x48a1 && v.IP() == 0x4944 {
			seed = v
			break
		}
	}
	if seed == nil {
		t.Fatal("entry absent")
	}
	m := original(seed)
	for m.Steps < 1000 {
		if m.CPU.Seg[cpu.CS] == 0x48a1 && m.CPU.IP == 0x4971 {
			return []*view{seed, {mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}}}
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("surface absent")
	return nil
}
func TestNativePaletteInitCosts(t *testing.T) {
	for _, v := range paletteInitStates(t) {
		for _, dos := range []bool{false, true} {
			v.dos = dos
			m := original(v)
			for m.Steps < 1000 {
				if m.Steps > 0 && (m.CPU.IP == 0x2aeb || m.CPU.IP == 0x4a31) {
					break
				}
				if err := m.Step(); err != nil {
					t.Fatal(err)
				}
			}
			t.Logf("init %x dos%v N%d C%d IP%x", v.IP(), dos, m.Steps, m.CPU.Cycles, m.CPU.IP)
		}
	}
}

func TestNativePaletteInitOriginal(t *testing.T) {
	cases := 0
	for _, seed := range paletteInitStates(t) {
		for _, flags := range []uint16{2, 0x2d7, 0xed7} {
			for _, dos := range []bool{false, true} {
				v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
				v.s.Flags = flags
				p := planPaletteInitialize(v, 100)
				if p == nil {
					t.Fatal("init rejected", v.IP())
				}
				compare(t, v, p)
				if planPaletteInitialize(v, p.Instructions-1) != nil {
					t.Fatal("short budget")
				}
				compare(t, v, planPaletteInitialize(v, p.Instructions))
				cases += 2
			}
		}
	}
	t.Log("init original comparisons", cases)
}

func TestNativePaletteInitSurfaces(t *testing.T) {
	seed := paletteInitStates(t)[1]
	cases := 0
	for _, links := range []uint16{0, 1, 3} {
		for _, bounds := range [][4]uint16{{0, 0, 319, 199}, {10, 12, 10, 12}, {50, 80, 90, 120}} {
			for _, dos := range []bool{false, true} {
				v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
				stack := uint32(v.SS())*16 + uint32(v.SP())
				putRowWord(v, stack+20, 0)
				for i := uint16(0); i <= links; i++ {
					seg, flags := uint16(0xa000), uint16(0x80)
					if i < links {
						seg, flags = (i+1)*2, 0xc0
					}
					for _, w := range [][2]uint16{{0xbd, seg}, {0x2bd, 0}, {0x4bd, bounds[0]}, {0x6bd, bounds[1]}, {0x8bd, bounds[2]}, {0xabd, bounds[3]}, {0xcbd, flags}} {
						putRowWord(v, 0x48a10+uint32(w[0]+2*i), w[1])
					}
				}
				for i := uint32(0); i < 768; i++ {
					v.mem[0x48a10+0x1572+i] = byte(i*11 + 7)
				}
				p := planPaletteInitialize(v, 100)
				if p == nil {
					t.Fatal("surface rejected", links, bounds, dos)
				}
				compare(t, v, p)
				cases++
			}
		}
	}
	t.Log("init surface comparisons", cases)
}

func TestNativePaletteInitRefusals(t *testing.T) {
	cases := 0
	for _, seed := range paletteInitStates(t) {
		for _, kind := range []string{"budget", "code", "mode", "CS", "SP-low", "SP-high", "stack-code", "stack-video", "argument"} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			b := uint64(100)
			stack := uint32(v.SS())*16 + uint32(v.SP())
			switch kind {
			case "budget":
				b = 1
			case "code":
				v.mem[0x48a10+uint32(v.IP())] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "CS":
				v.s.Seg[cpu.CS] = 0x4800
			case "SP-low":
				if v.IP() == 0x4971 {
					v.s.R[cpu.BP]++
				} else {
					v.s.R[cpu.SP] = 1
				}
			case "SP-high":
				v.s.R[cpu.SP] = 65535
			case "stack-code":
				v.s.Seg[cpu.SS] = 0x48a1
				v.s.R[cpu.SP] = 0x7000
			case "stack-video":
				v.s.Seg[cpu.SS] = 0xa000
			case "argument":
				off := uint32(6)
				if v.IP() == 0x4971 {
					off = 20
				}
				putRowWord(v, stack+off, 256)
			}
			before := append([]byte(nil), v.mem...)
			regs := v.s
			if planPaletteInitialize(v, b) != nil {
				t.Fatal("invalid init accepted", seed.IP(), kind)
			}
			if string(before) != string(v.mem) || regs != v.s {
				t.Fatal("refusal mutated")
			}
			cases++
		}
	}
	t.Log("init refusals", cases)
}

func TestNativePaletteInitSurfaceRefusals(t *testing.T) {
	seed := paletteInitStates(t)[1]
	for _, kind := range []string{"DS", "root-seg", "left", "right", "top", "bottom", "chain-odd", "chain-large", "chain-cycle"} {
		v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
		switch kind {
		case "DS":
			v.s.Seg[cpu.DS] = 0x4800
		case "root-seg":
			putRowWord(v, 0x48a10+0xbd, 0x9000)
		case "left":
			putRowWord(v, 0x48a10+0x4bd, 320)
		case "right":
			putRowWord(v, 0x48a10+0x8bd, 320)
		case "top":
			putRowWord(v, 0x48a10+0x6bd, 200)
		case "bottom":
			putRowWord(v, 0x48a10+0xabd, 200)
		default:
			putRowWord(v, 0x48a10+0xcbd, 0xc0)
			n := uint16(0)
			if kind == "chain-odd" {
				n = 1
			} else if kind == "chain-large" {
				n = 512
			}
			putRowWord(v, 0x48a10+0xbd, n)
		}
		if planPaletteInitialize(v, 100) != nil {
			t.Fatal("invalid surface accepted", kind)
		}
	}
}
