package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func paletteSourceStates(t *testing.T) []*view {
	var seed *view
	for _, v := range drawCalls(t) {
		if v.CS() == 0x48a1 && v.IP() == 0x4944 {
			seed = v
			break
		}
	}
	if seed == nil {
		t.Fatal("source entry absent")
	}
	m := original(seed)
	targets := map[uint16]bool{0x4a91: true, 0x4aa4: true, 0x4aad: true, 0x4abd: true, 0x4ae3: true, 0x4af6: true}
	seen := map[uint32]bool{}
	out := []*view{}
	for m.Steps < 1000000 {
		ip := m.CPU.IP
		key := uint32(ip)
		if ip == 0x4aad && m.CPU.R[cpu.SI] == 0 {
			key |= 1 << 16
		}
		if m.CPU.Seg[cpu.CS] == 0x48a1 && targets[ip] && !seen[key] {
			seen[key] = true
			out = append(out, &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}})
			if ip == 0x4af6 {
				break
			}
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if len(out) != 7 {
		t.Fatal("source protocol states", len(out), seen)
	}
	return out
}

func TestNativePaletteSourceCost(t *testing.T) {
	ends := map[uint16]bool{0x2a90: true, 0x2aeb: true, 0x4aad: true, 0x4abd: true, 0x4af6: true, 0x4afd: true, 0x4bc6: true}
	for _, v := range paletteSourceStates(t) {
		for _, dos := range []bool{false, true} {
			v.dos = dos
			m := original(v)
			for m.Steps < 100 {
				if m.Steps > 0 && ends[m.CPU.IP] {
					break
				}
				if err := m.Step(); err != nil {
					t.Fatal(err)
				}
			}
			t.Logf("source %x dos%v N%d C%d R%v Seg%v IP%x F%x", v.IP(), dos, m.Steps, m.CPU.Cycles, m.CPU.R, m.CPU.Seg, m.CPU.IP, m.CPU.Flags)
		}
	}
}

func TestNativePaletteSourceOriginal(t *testing.T) {
	cases := 0
	for _, v := range paletteSourceStates(t) {
		for _, dos := range []bool{false, true} {
			v.dos = dos
			p := planPaletteSourceControl(v, 100)
			if p == nil {
				p = planPaletteSourceSample(v, 100)
			}
			if p == nil {
				p = planPaletteSourceStore(v, 100)
			}
			if p == nil {
				t.Fatal("source state rejected", v.IP())
			}
			compare(t, v, p)
			cases++
		}
	}
	t.Log("source palette original entry-state comparisons", cases)
}

func sourcePlan(v CPUView, b uint64) *machine.HostWorkPlan {
	if p := planPaletteSourceControl(v, b); p != nil {
		return p
	}
	if p := planPaletteSourceSample(v, b); p != nil {
		return p
	}
	return planPaletteSourceStore(v, b)
}

func TestNativePaletteSourceBoundaries(t *testing.T) {
	seeds := paletteSourceStates(t)
	cases := 0
	for _, seed := range seeds {
		for _, count := range []uint16{1, 17, 109, 256} {
			for _, flags := range []uint16{2, 0x2d7, 0xad7} {
				for _, dos := range []bool{false, true} {
					v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
					v.s.Flags = flags
					putRowWord(v, 0x48a10+0x116f, count)
					switch v.IP() {
					case 0x4a91:
						v.s.R[cpu.DI] = 0x1572 + 2*count
					case 0x4aad:
						if seed.SI() == 0 {
							v.s.R[cpu.SI] = 0
							v.s.R[cpu.DI] = 0x1872 + 3*count
						} else {
							v.s.R[cpu.SI] = count * 2
							v.s.R[cpu.DI] = 0x1872
						}
					case 0x4abd, 0x4ae3:
						v.s.R[cpu.SI] = (count - 1) * 2
						v.s.R[cpu.DI] = 0x1872
						putRowWord(v, 0x48a10+0x1572+uint32(v.SI()), 255)
					}
					p := sourcePlan(v, 100)
					if p == nil {
						t.Fatalf("source rejected %x count%d flags%x", v.IP(), count, flags)
					}
					compare(t, v, p)
					if sourcePlan(v, p.Instructions-1) != nil {
						t.Fatal("short budget accepted")
					}
					compare(t, v, sourcePlan(v, p.Instructions))
					cases += 2
				}
			}
		}
	}
	v := &view{mem: append([]byte(nil), seeds[0].mem...), s: seeds[0].s}
	v.s.R[cpu.DI] = 0x1572
	compare(t, v, sourcePlan(v, 3))
	cases++
	t.Log("source boundary comparisons", cases)
}

func TestNativePaletteSourceRefusals(t *testing.T) {
	seeds := paletteSourceStates(t)
	cases := 0
	for _, seed := range seeds {
		for _, kind := range []string{"budget", "CS", "DS", "DF", "mode", "code"} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			b := uint64(100)
			switch kind {
			case "budget":
				b = 0
			case "CS":
				v.s.Seg[cpu.CS] = 0x4800
			case "DS":
				v.s.Seg[cpu.DS] = 0x4800
			case "DF":
				v.s.Flags |= cpu.DF
			case "mode":
				v.mem[0x449] = 0xd
			case "code":
				lo := uint32(0x4a91)
				if v.IP() == 0x4abd {
					lo = 0x4abd
				} else if v.IP() >= 0x4ae3 {
					lo = 0x4ae3
				}
				v.mem[0x48a10+lo] ^= 1
			}
			before := append([]byte(nil), v.mem...)
			regs := v.s
			if sourcePlan(v, b) != nil {
				t.Fatal("invalid source accepted", seed.IP(), kind)
			}
			if string(before) != string(v.mem) || regs != v.s {
				t.Fatal("refusal mutated input")
			}
			cases++
		}
	}
	for _, seed := range seeds {
		if seed.IP() == 0x4a91 {
			continue
		}
		for _, n := range []uint16{0, 257} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			putRowWord(v, 0x48a10+0x116f, n)
			if sourcePlan(v, 100) != nil {
				t.Fatal("invalid count")
			}
			cases++
		}
	}
	for _, seed := range seeds {
		if seed.IP() != 0x4abd && seed.IP() != 0x4ae3 {
			continue
		}
		for _, kind := range []string{"odd-index", "range-index", "cursor", "stack-wrap", "stack-video", "stack-driver", "component"} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			switch kind {
			case "odd-index":
				v.s.R[cpu.SI] |= 1
			case "range-index":
				v.s.R[cpu.SI] = 512
			case "cursor":
				v.s.R[cpu.DI]++
			case "stack-wrap":
				v.s.R[cpu.SP] = 65535
			case "stack-video":
				v.s.Seg[cpu.SS] = 0xa000
			case "stack-driver":
				v.s.Seg[cpu.SS] = 0x48a1
				v.s.R[cpu.SP] = 0x7000
			case "component":
				if v.IP() == 0x4abd {
					putRowWord(v, 0x48a10+0x1572+uint32(v.SI()), 256)
				} else {
					putRowWord(v, 0x48a10+0x117f, 64)
				}
			}
			if sourcePlan(v, 100) != nil {
				t.Fatal("invalid data accepted", seed.IP(), kind)
			}
			cases++
		}
	}
	t.Log("source refusals", cases)
}
