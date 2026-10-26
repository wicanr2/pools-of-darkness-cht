package drawblocks

import (
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func paletteSetupSeed(t *testing.T, ip uint16) *view {
	seed := paletteSeed(t)
	v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
	v.s.IP = ip
	v.s.Seg[cpu.DS], v.s.Seg[cpu.ES] = 0x48a1, 0x48a1
	v.s.Flags &^= cpu.DF
	return v
}

func TestNativePaletteListOriginal(t *testing.T) {
	plans := 0
	for _, pattern := range []int{0, 1, 2, 3} {
		for _, dos := range []bool{false, true} {
			v := paletteSetupSeed(t, 0x4a83)
			v.dos = dos
			v.s.R[cpu.BX], v.s.R[cpu.DI], v.s.R[cpu.AX] = 255, 0x1572, 0x3507
			for i := 0; i < 256; i++ {
				value := byte(0)
				if pattern == 1 || (pattern == 2 && i%2 == 0) || (pattern == 3 && (i == 0 || i == 255)) {
					value = 7
				}
				v.mem[0x48a10+0x1772+i] = value
			}
			for _, budget := range []uint64{5, 7, 32, 512, 4096} {
				if p := planPaletteList(v, budget); p != nil {
					compare(t, v, p)
					plans++
				}
			}
		}
	}
	t.Log("palette list synthetic original comparisons", plans)
}

func TestNativePaletteTargetOriginal(t *testing.T) {
	plans := 0
	for _, count := range []uint16{1, 2, 109, 256} {
		for _, rgb := range [][3]byte{{0, 0, 0}, {63, 7, 31}, {255, 128, 64}} {
			for _, dos := range []bool{false, true} {
				v := paletteSetupSeed(t, 0x4afd)
				v.dos = dos
				v.s.R[cpu.CX], v.s.R[cpu.DI], v.s.R[cpu.AX] = count, 0x1b72, 0x3500
				v.mem[0x48a10+0x1179], v.mem[0x48a10+0x117b], v.mem[0x48a10+0x117d] = rgb[0], rgb[1], rgb[2]
				for _, budget := range []uint64{7, 14, 2048} {
					p := planPaletteTarget(v, budget)
					if p == nil {
						t.Fatal("valid RGB target rejected")
					}
					compare(t, v, p)
					plans++
				}
			}
		}
	}
	t.Log("palette target synthetic original comparisons", plans)
}

func TestNativePaletteErrorOriginal(t *testing.T) {
	plans := 0
	for _, count := range []uint16{1, 2, 109, 256} {
		for _, maximum := range []byte{0, 1, 31, 63} {
			for _, dos := range []bool{false, true} {
				v := paletteSetupSeed(t, 0x4b3a)
				v.dos = dos
				v.s.R[cpu.BX] = 0x3500 | uint16(maximum)
				v.mem[0x48a10+0x116f], v.mem[0x48a10+0x1170] = byte(count), byte(count>>8)
				p := planPaletteError(v, 9)
				if p == nil {
					t.Fatal("valid interpolation initializer rejected")
				}
				compare(t, v, p)
				plans++
			}
		}
	}
	t.Log("palette error synthetic original comparisons", plans)
}

func TestNativePaletteSetupStates(t *testing.T) {
	m := original(paletteSeed(t))
	states := map[uint16][]*view{}
	for m.Steps < 1000000 && (len(states[0x4a83]) < 32 || len(states[0x4afd]) < 32 || len(states[0x4b3a]) < 1) {
		ip := m.CPU.IP
		limit := 32
		if ip == 0x4b3a {
			limit = 1
		}
		if m.CPU.Seg[cpu.CS] == 0x48a1 && (ip == 0x4a83 || ip == 0x4afd || ip == 0x4b3a) && len(states[ip]) < limit {
			states[ip] = append(states[ip], &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}})
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	plans := 0
	for _, ip := range []uint16{0x4a83, 0x4afd, 0x4b3a} {
		want := 32
		if ip == 0x4b3a {
			want = 1
		}
		if len(states[ip]) != want {
			t.Fatalf("palette setup states absent %04x %d", ip, len(states[ip]))
		}
		for _, v := range states[ip] {
			for _, dos := range []bool{false, true} {
				v.dos = dos
				for _, budget := range []uint64{9, 64, 4096} {
					var p *machine.HostWorkPlan
					switch ip {
					case 0x4a83:
						p = planPaletteList(v, budget)
					case 0x4afd:
						p = planPaletteTarget(v, budget)
					case 0x4b3a:
						p = planPaletteError(v, budget)
					}
					if p == nil {
						t.Fatalf("palette setup original state rejected %04x", ip)
					}
					compare(t, v, p)
					plans++
				}
			}
		}
	}
	t.Log("palette setup original entry-state comparisons", plans)
}

func TestNativePaletteSetupRefusals(t *testing.T) {
	for _, ip := range []uint16{0x4a83, 0x4afd, 0x4b3a} {
		for _, kind := range []string{"budget", "code", "mode", "direction", "segment", "count", "destination"} {
			t.Run(kind, func(t *testing.T) {
				v := paletteSetupSeed(t, ip)
				v.s.R[cpu.BX], v.s.R[cpu.DI], v.s.R[cpu.CX] = 255, 0x1572, 1
				if ip == 0x4afd {
					v.s.R[cpu.DI] = 0x1b72
				}
				if ip == 0x4b3a {
					v.s.R[cpu.BX] = 31
					v.mem[0x48a10+0x116f], v.mem[0x48a10+0x1170] = 1, 0
				}
				budget := uint64(4096)
				switch kind {
				case "budget":
					budget = 1
				case "code":
					v.mem[0x48a10+uint32(ip)] ^= 1
				case "mode":
					v.mem[0x449] = 0xd
				case "direction":
					v.s.Flags |= cpu.DF
				case "segment":
					v.s.Seg[cpu.ES] = 0xa000
				case "count":
					switch ip {
					case 0x4a83:
						v.s.R[cpu.BX] = 256
					case 0x4afd:
						v.s.R[cpu.CX] = 0
					case 0x4b3a:
						v.mem[0x48a10+0x116f] = 0
					}
				case "destination":
					if ip == 0x4b3a {
						v.s.R[cpu.BX] = 64
					} else {
						v.s.R[cpu.DI]++
					}
				}
				var p *machine.HostWorkPlan
				switch ip {
				case 0x4a83:
					p = planPaletteList(v, budget)
				case 0x4afd:
					p = planPaletteTarget(v, budget)
				case 0x4b3a:
					p = planPaletteError(v, budget)
				}
				if p != nil {
					t.Fatalf("unproved palette setup accepted %04x", ip)
				}
			})
		}
	}
}
