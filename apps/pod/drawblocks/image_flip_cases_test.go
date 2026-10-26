package drawblocks

import (
	"fmt"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestResearchMirrorPacketsOriginal(t *testing.T) {
	plans := 0
	for _, repeat := range []bool{false, true} {
		for _, count := range []uint16{1, 2, 3, 7, 8, 9, 17, 64, 127, 128} {
			for _, extra := range []bool{false, true} {
				for _, dos := range []bool{false, true} {
					for _, flags := range []uint16{0x7246, 0xf002} {
						v := imageSpanSeed(t)
						v.dos = dos
						v.s.IP = 0x6c8f
						v.s.R[cpu.CX] = 0
						v.s.R[cpu.SP] = 0x900
						v.s.R[cpu.DI] = uint16(3200 + count - 1)
						v.s.Flags = flags
						stream := []byte{byte((count - 1) * 2)}
						for i := uint16(0); i < count; i++ {
							stream = append(stream, byte(i*11+3))
						}
						if repeat {
							stream = []byte{byte((count-1)*2 + 1), 0x57}
						}
						if extra {
							stream = append(stream, 1, 0x99)
						}
						copy(v.mem[0x50100:], stream)
						v.s.R[cpu.BX] = 0x100 + uint16(len(stream))
						p := planMirrorPacket(v, 1000)
						if p == nil {
							t.Fatal("mirror packet rejected")
						}
						compare(t, v, p)
						plans++
					}
				}
			}
		}
	}
	t.Log("mirror full packet original comparisons", plans)
}

func TestResearchFlipControlledStates(t *testing.T) {
	seed := rleSeed(t)
	seen := map[uint16]int{}
	selected := map[uint32]int{}
	states := []*view{}
	targets := map[uint16]bool{0x6a0c: true, 0x6a58: true, 0x6abe: true, 0x6b33: true, 0x6b5d: true, 0x6b71: true, 0x6bea: true, 0x6bee: true, 0x6bfe: true, 0x6c8f: true, 0x6a3a: true, 0x6a48: true, 0x6a50: true, 0x6c4d: true, 0x6c50: true, 0x6c53: true, 0x6cf9: true, 0x6cf6: true, 0x6d08: true, 0x6dc0: true, 0x6dc4: true, 0x6dd2: true}
	for _, flags := range []uint16{1, 2, 3} {
		for _, x := range []int16{-7, 0, 7} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			stack := uint32(v.SS())*16 + uint32(v.SP())
			putRowWord(v, stack+6, uint16(x))
			putRowWord(v, stack+10, flags)
			m := original(v)
			retIP, retCS := v.Read16(stack), v.Read16(stack+2)
			for m.Steps < 1000000 {
				if m.Steps > 0 && m.CPU.IP == retIP && m.CPU.Seg[cpu.CS] == retCS {
					break
				}
				ip := m.CPU.IP
				key := uint32(flags)<<16 | uint32(ip)
				if m.CPU.Seg[cpu.CS] == 0x48a1 && targets[ip] && selected[key] < 4 {
					selected[key]++
					seen[ip]++
					states = append(states, &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}})
				}
				if err := m.Step(); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	plans := 0
	for _, v := range states {
		for _, dos := range []bool{false, true} {
			v.dos = dos
			p := planFlipInitialize(v, 1024)
			if p == nil {
				p = planFlipRow(v, 1024)
			}
			if p == nil {
				p = planFlipHeader(v, 1024)
			}
			if p == nil {
				p = planMirrorPacket(v, 1024)
			}
			if p == nil {
				p = planFlipControl(v, 1024)
			}
			if p == nil {
				p = planMirrorSpan(v, 1024)
			}
			if p == nil {
				p = planMirrorClipSetup(v, 1024)
			}
			if p == nil {
				p = planMirrorClippedPacket(v, 1024)
			}
			if p == nil {
				p = planRLERunsProgram(v, 1024, true)
			}
			if p == nil {
				p = planClipSetupProgram(v, 1024, flippedClipProgram)
			}
			if p == nil {
				p = planClippedRunProgram(v, 1024, flippedClipProgram)
			}
			if p == nil {
				t.Fatal("flip state rejected", v.s)
			}
			compare(t, v, p)
			plans++
		}
	}
	if seen[0x6a0c] == 0 || seen[0x6b33] == 0 || seen[0x6c8f] == 0 {
		t.Fatal("flip paths missing", seen)
	}
	t.Log("controlled flip state comparisons", plans, "states", len(states), "paths", seen)
}

func TestResearchFlipGeometryOriginal(t *testing.T) {
	plans := 0
	for _, flags := range []uint16{1, 2, 3} {
		for _, shape := range [][2]uint16{{1, 1}, {304, 120}, {320, 200}} {
			for _, xy := range [][2]int16{{-7, -5}, {0, 0}, {7, 7}} {
				for _, dos := range []bool{false, true} {
					v := imageRowSeed(t)
					v.s.IP = 0x6a0c
					v.dos = dos
					putRowWord(v, 0x50100, shape[0])
					putRowWord(v, 0x50102, shape[1])
					putRowWord(v, 0x70808, uint16(xy[0]))
					putRowWord(v, 0x7080a, uint16(xy[1]))
					putRowWord(v, 0x7080c, flags)
					p := planFlipInitialize(v, 100)
					if p == nil {
						t.Fatal("valid flip geometry rejected")
					}
					compare(t, v, p)
					plans++
					for _, raw := range []byte{0, 1, 119, 199, 255} {
						w := imageRowSeed(t)
						w.dos = dos
						w.s.IP = 0x6a58
						w.mem[0x50100] = raw
						putRowWord(w, 0x7080c, flags)
						putRowWord(w, 0x48a10+0x1204, uint16(xy[1]))
						putRowWord(w, 0x48a10+0x1206, uint16(int32(xy[1])+int32(shape[1])-1))
						p := planFlipRow(w, 100)
						if p == nil {
							t.Fatal("valid flip row rejected")
						}
						compare(t, w, p)
						plans++
					}
				}
			}
		}
	}
	for _, x := range []int16{-100, -1, 0, 7, 310, 319, 320, 400} {
		for _, width := range []byte{1, 2, 17, 255} {
			for _, dos := range []bool{false, true} {
				v := imageSpanSeed(t)
				v.dos = dos
				v.s.IP = 0x6c53
				v.s.R[cpu.AX] = uint16(x)
				v.mem[0x50100], v.mem[0x50101] = width, 4
				p := planMirrorSpan(v, 100)
				if p == nil {
					t.Fatal("valid mirror span rejected", x, width)
				}
				compare(t, v, p)
				plans++
			}
		}
	}
	t.Log("flip geometry original comparisons", plans)
}

func TestResearchMirrorClippedPackets(t *testing.T) {
	plans := 0
	for _, leading := range []bool{false, true} {
		for _, repeat := range []bool{false, true} {
			for _, run := range []uint16{1, 2, 3, 8, 9, 17, 64, 127, 128} {
				for _, rem := range []uint16{1, 4, 17, 128, 200} {
					for _, left := range []uint16{0, 311} {
						for _, extra := range []bool{false, true} {
							for _, dos := range []bool{false, true} {
								v := imageSpanSeed(t)
								v.dos = dos
								v.s.IP = 0x6dd2
								v.s.R[cpu.CX] = 0
								v.s.R[cpu.DX] = 0
								v.s.R[cpu.SP] = 0x900
								v.s.R[cpu.DI] = 3519
								if leading {
									v.s.IP = 0x6d08
									v.s.R[cpu.DI] += rem
								}
								putRowWord(v, 0x48a10+0x1230, rem)
								putRowWord(v, 0x48a10+0x1232, 319+rem)
								putRowWord(v, 0x48a10+0xec1, left)
								stream := []byte{byte((run - 1) * 2)}
								for i := uint16(0); i < run; i++ {
									stream = append(stream, byte(i*11+3))
								}
								if repeat {
									stream = []byte{byte((run-1)*2 + 1), 0x57}
								}
								if extra {
									stream = append(stream, 1, 0x99)
								}
								copy(v.mem[0x50100:], stream)
								v.s.R[cpu.BX] = 0x100 + uint16(len(stream))
								p := planMirrorClippedPacket(v, 2000)
								if p == nil {
									continue
								}
								t.Run(fmt.Sprintf("leading%v/repeat%v/run%d/rem%d/left%d/extra%v/dos%v", leading, repeat, run, rem, left, extra, dos), func(t *testing.T) { compare(t, v, p) })
								plans++
							}
						}
					}
				}
			}
		}
	}
	t.Log("mirror clipped packet original comparisons", plans)
}

type imageModelView struct {
	*view
	model cpu.Model
	known bool
}

func (v imageModelView) CPUModel() (cpu.Model, bool) { return v.model, v.known }

func TestImageMirrorCPUModels(t *testing.T) {
	for _, model := range []cpu.Model{cpu.Model8086, cpu.Model80186, cpu.Model80386} {
		for _, dos := range []bool{false, true} {
			for _, clipped := range []bool{false, true} {
				v := imageSpanSeed(t)
				v.dos = dos
				v.s.IP = 0x6c8f
				v.s.R[cpu.CX], v.s.R[cpu.SP], v.s.R[cpu.DI] = 0, 0x900, 3519
				v.s.Flags = 0x7246
				copy(v.mem[0x50100:], []byte{16, 1, 2, 3, 4, 5, 6, 7, 8, 9})
				v.s.R[cpu.BX] = 0x10a
				mv := imageModelView{v, model, true}
				var p *machine.HostWorkPlan
				if clipped {
					v.s.IP = 0x6dd2
					v.s.R[cpu.DX] = 0
					putRowWord(v, 0x48a10+0x1230, 8)
					p = planMirrorClippedPacket(mv, 1000)
				} else {
					p = planMirrorPacket(mv, 1000)
				}
				if p == nil {
					t.Fatal("model rejected", model, clipped)
				}
				compare(t, v, p, model)
				if clipped {
					p = planMirrorClippedPacket(imageModelView{v, model, false}, 1000)
				} else {
					p = planMirrorPacket(imageModelView{v, model, false}, 1000)
				}
				if p != nil {
					t.Fatal("unknown model admitted")
				}
			}
		}
	}
}

func TestImageMirrorRefusals(t *testing.T) {
	seed := func() *view {
		v := imageSpanSeed(t)
		v.s.IP = 0x6c8f
		v.s.R[cpu.CX], v.s.R[cpu.SP], v.s.R[cpu.DI] = 0, 0x900, 3519
		copy(v.mem[0x50100:], []byte{6, 1, 2, 3, 4})
		v.s.R[cpu.BX] = 0x105
		return v
	}
	cases := []struct {
		name   string
		change func(*view)
	}{
		{"code", func(v *view) { v.mem[0x48a10+0x6a0c] ^= 1 }},
		{"mode", func(v *view) { v.mem[0x449] = 0x0d }},
		{"CS", func(v *view) { v.s.Seg[cpu.CS]++ }},
		{"ES", func(v *view) { v.s.Seg[cpu.ES]-- }},
		{"DF", func(v *view) { v.s.Flags |= cpu.DF }},
		{"count", func(v *view) { v.s.R[cpu.CX] = 0x100 }},
		{"truncated", func(v *view) { v.s.R[cpu.BX]-- }},
		{"empty", func(v *view) { v.s.R[cpu.BX] = v.SI() }},
		{"source video", func(v *view) { v.s.Seg[cpu.DS] = 0xa000 }},
		{"source low", func(v *view) { v.s.Seg[cpu.DS] = 0 }},
		{"source scratch", func(v *view) { v.s.Seg[cpu.DS] = 0x48a1; v.s.R[cpu.SI] = 0x1230; v.s.R[cpu.BX] = 0x1235 }},
		{"destination", func(v *view) { v.s.R[cpu.DI] = 64000 }},
		{"row wrap", func(v *view) { v.s.R[cpu.DI] = 3200 }},
		{"stack wrap", func(v *view) { v.s.R[cpu.SP] = 1 }},
		{"stack video", func(v *view) { v.s.Seg[cpu.SS] = 0xa000 }},
		{"stack code", func(v *view) { v.s.Seg[cpu.SS] = 0x48a1 }},
		{"stack source", func(v *view) { v.s.Seg[cpu.SS] = v.DS(); v.s.R[cpu.SP] = 0x104 }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			v := seed()
			c.change(v)
			before := append([]byte(nil), v.mem...)
			regs := v.s
			if p := planMirrorPacket(v, 1000); p != nil {
				t.Fatal("invalid packet admitted")
			}
			if string(v.mem) != string(before) || v.s != regs {
				t.Fatal("decline mutated input")
			}
		})
	}
	v := seed()
	p := planMirrorPacket(v, 1000)
	if p == nil {
		t.Fatal("valid packet rejected")
	}
	if planMirrorPacket(v, p.Instructions-1) != nil {
		t.Fatal("short budget admitted")
	}
	compare(t, v, planMirrorPacket(v, p.Instructions))
}

func TestImageExhaustedClipOriginal(t *testing.T) {
	for _, mirror := range []bool{false, true} {
		for _, repeat := range []bool{false, true} {
			for _, dos := range []bool{false, true} {
				v := imageSpanSeed(t)
				v.dos = dos
				v.s.IP = 0x69bd
				if mirror {
					v.s.IP = 0x6dd2
				}
				v.s.R[cpu.CX], v.s.R[cpu.DX], v.s.R[cpu.DI] = 0, 0, 3519
				v.s.R[cpu.BX] = 0x102
				v.mem[0x50100], v.mem[0x50101] = 0, 3
				if repeat {
					v.mem[0x50100] = 1
				}
				putRowWord(v, 0x48a10+0x1230, 0)
				var p *machine.HostWorkPlan
				if mirror {
					p = planMirrorClippedPacket(v, 1000)
				} else {
					p = planClippedRun(v, 1000)
				}
				if p == nil {
					t.Fatal("exhausted packet rejected", mirror, repeat, dos)
				}
				compare(t, v, p)
			}
		}
	}
}

func TestResearchImageInitializeCost(t *testing.T) {
	for _, flags := range []uint16{0, 1} {
		for _, missing := range []bool{false, true} {
			v := rleSeed(t)
			stack := uint32(v.SS())*16 + uint32(v.SP())
			putRowWord(v, stack+10, flags)
			if missing {
				putRowWord(v, stack+16, 65535)
			}
			for _, dos := range []bool{false, true} {
				v.dos = dos
				m := original(v)
				for m.Steps < 1024 {
					if m.CPU.IP == 0x6850 || m.CPU.IP == 0x6a0c || m.CPU.Seg[cpu.CS] != 0x48a1 {
						break
					}
					if err := m.Step(); err != nil {
						t.Fatal(err)
					}
				}
				t.Logf("flags%d missing%v dos%v N%d C%d R%v Seg%v IP%x F%x source%04x:%04x slot%d", flags, missing, dos, m.Steps, m.CPU.Cycles, m.CPU.R, m.CPU.Seg, m.CPU.IP, m.CPU.Flags, v.Read16(stack+14), v.Read16(stack+12), v.Read16(stack+4))
			}
		}
	}
}
