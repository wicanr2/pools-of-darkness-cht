package drawblocks

import (
	"bytes"
	"fmt"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func TestResearchImageControlOriginal(t *testing.T) {
	plans := 0
	for _, dos := range []bool{false, true} {
		for _, ip := range []uint16{0x6832, 0x689e} {
			v := imageSpanSeed(t)
			v.dos = dos
			v.s.IP = ip
			putRowWord(v, 0x48a10+0x1250, 0x8001)
			p := planImageControl(v, 100)
			if p == nil {
				t.Fatal("row finish rejected")
			}
			compare(t, v, p)
			plans++
		}
		for _, ip := range []uint16{0x6832, 0x6840, 0x6848} {
			v := imageSpanSeed(t)
			v.dos = dos
			v.s.IP = ip
			p := planImageControl(v, 100)
			if p == nil {
				t.Fatal("excluded span rejected")
			}
			compare(t, v, p)
			plans++
		}
		for _, bounds := range [][2]uint16{{0, 319}, {40, 120}} {
			for _, x := range []int16{-100, -7, 0, 41, 110, 310, 400} {
				v := imageSpanSeed(t)
				v.dos = dos
				putRowWord(v, 0x70808, uint16(x))
				putRowWord(v, 0x48a10+0xec1, bounds[0])
				putRowWord(v, 0x48a10+0xec5, bounds[1])
				if int32(x) >= int32(bounds[0]) && int32(x)+16 <= int32(bounds[1]) {
					continue
				}
				p := planImageControl(v, 100)
				if p == nil {
					t.Fatal("clipped span rejected", x, bounds)
				}
				compare(t, v, p)
				plans++
			}
		}
	}
	t.Log("image control original comparisons", plans)
}

func TestResearchImageClipRefusals(t *testing.T) {
	for _, kind := range []string{"budget", "code", "mode", "direction", "surface", "high-counter", "source-video", "source-system", "source-alias", "counter-negative", "counter-large", "packet", "destination", "row", "end"} {
		t.Run("run/"+kind, func(t *testing.T) {
			v := imageSpanSeed(t)
			v.s.IP = 0x69bd
			v.s.R[cpu.DX] = 0
			v.s.R[cpu.CX] = 0
			v.s.R[cpu.BX] = 0x102
			v.s.R[cpu.DI] = 3200
			v.mem[0x50100] = 3
			v.mem[0x50101] = 7
			putRowWord(v, 0x48a10+0x1230, 2)
			budget := uint64(100)
			switch kind {
			case "budget":
				budget = 1
			case "code":
				v.mem[0x48a10+0x691c] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "direction":
				v.s.Flags |= cpu.DF
			case "surface":
				v.s.Seg[cpu.ES] = 0x9000
			case "high-counter":
				v.s.R[cpu.CX] = 256
			case "source-video":
				v.s.Seg[cpu.DS] = 0xa000
			case "source-system":
				v.s.Seg[cpu.DS] = 0
				v.s.R[cpu.SI] = 0x100
				v.s.R[cpu.BX] = 0x102
			case "source-alias":
				v.s.Seg[cpu.DS] = 0x48a1
				v.s.R[cpu.SI] = 0x1230
				v.s.R[cpu.BX] = 0x1234
			case "counter-negative":
				putRowWord(v, 0x48a10+0x1230, 65535)
			case "counter-large":
				putRowWord(v, 0x48a10+0x1230, 321)
			case "packet":
				v.mem[0x50100] = 2
			case "destination":
				v.s.R[cpu.DI] = 63999
			case "row":
				v.s.R[cpu.DI] = 319
			case "end":
				v.s.R[cpu.BX] = v.s.R[cpu.SI]
			}
			before := append([]byte(nil), v.mem...)
			if planClippedRun(v, budget) != nil {
				t.Fatal("unproved clipped run accepted")
			}
			if !bytes.Equal(before, v.mem) {
				t.Fatal("refusal wrote memory")
			}
		})
	}
	for _, kind := range []string{"budget", "code", "mode", "direction", "source-alias", "argument-alias", "argument-range", "wrap", "bounds", "zero-width", "zero-payload"} {
		t.Run("header/"+kind, func(t *testing.T) {
			v := imageSpanSeed(t)
			putRowWord(v, 0x70808, uint16(65529))
			budget := uint64(100)
			switch kind {
			case "budget":
				budget = 1
			case "code":
				v.mem[0x48a10+0x689e] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "direction":
				v.s.Flags |= cpu.DF
			case "source-alias":
				v.s.Seg[cpu.DS] = 0x48a1
				v.s.R[cpu.SI] = 0x1250
			case "argument-alias":
				v.s.Seg[cpu.SS] = 0x48a1
				v.s.R[cpu.BP] = 0x1248
			case "argument-range":
				v.s.R[cpu.BP] = 65530
			case "wrap":
				v.s.R[cpu.SI] = 65534
			case "bounds":
				putRowWord(v, 0x48a10+0xec5, 320)
			case "zero-width":
				v.mem[0x50102] = 0
			case "zero-payload":
				v.mem[0x50103] = 0
			}
			before := append([]byte(nil), v.mem...)
			if planImageControl(v, budget) != nil {
				t.Fatal("unproved clipped header accepted")
			}
			if !bytes.Equal(before, v.mem) {
				t.Fatal("refusal wrote memory")
			}
		})
	}
}

func TestResearchClippedPacketsOriginal(t *testing.T) {
	plans := 0
	for _, left := range []bool{false, true} {
		for _, repeat := range []bool{false, true} {
			for _, run := range []uint16{1, 2, 7, 17, 128} {
				for _, remain := range []uint16{0, 1, 4, 17, 200} {
					if !left && remain == 0 {
						continue
					}
					for _, right := range []uint16{6, 319} {
						for _, extra := range []bool{false, true} {
							for _, dos := range []bool{false, true} {
								v := imageSpanSeed(t)
								v.dos = dos
								v.s.IP = 0x69bd
								v.s.R[cpu.DX] = 0
								v.s.R[cpu.DI] = 3200
								v.s.R[cpu.CX] = 0
								if left {
									v.s.IP = 0x6930
									v.s.R[cpu.DI] -= remain
								}
								putRowWord(v, 0x48a10+0x1230, remain)
								putRowWord(v, 0x48a10+0x1232, uint16(-int32(remain)))
								putRowWord(v, 0x48a10+0xec5, right)
								stream := []byte{byte((run - 1) * 2)}
								for n := uint16(0); n < run; n++ {
									stream = append(stream, byte(n*11+3))
								}
								if repeat {
									stream = []byte{byte((run-1)*2 + 1), 0x57}
								}
								if extra {
									stream = append(stream, 1, 0x99)
								}
								copy(v.mem[0x50100:], stream)
								v.s.R[cpu.BX] = 0x100 + uint16(len(stream))
								p := planClippedRun(v, 100)
								if p == nil {
									if left && int32(run)-int32(remain) <= 0 {
										t.Fatal("skip packet rejected")
									}
									continue
								}
								t.Run(testClipName(left, repeat, run, remain, right, extra, dos), func(t *testing.T) { compare(t, v, p) })
								plans++
							}
						}
					}
				}
			}
		}
	}
	t.Log("clipped packet original comparisons", plans)
}
func testClipName(left, repeat bool, run, remain, right uint16, extra, dos bool) string {
	return fmt.Sprintf("left%v/repeat%v/n%d/rem%d/right%d/extra%v/dos%v", left, repeat, run, remain, right, extra, dos)
}

func TestResearchClippedImageStates(t *testing.T) {
	seed := rleSeed(t)
	seen := map[uint16]int{}
	states := []*view{}
	for _, x := range []int16{-7, 7, -100, 100} {
		v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
		stack := uint32(v.SS())*16 + uint32(v.SP())
		putRowWord(v, stack+6, uint16(x))
		putRowWord(v, stack+10, 0)
		m := original(v)
		retIP, retCS := v.Read16(stack), v.Read16(stack+2)
		for m.Steps < 1000000 {
			if m.Steps > 0 && m.CPU.IP == retIP && m.CPU.Seg[cpu.CS] == retCS {
				break
			}
			ip := m.CPU.IP
			if m.CPU.Seg[cpu.CS] == 0x48a1 && (ip == 0x691c || ip == 0x6930 || ip == 0x69a9 || ip == 0x69ad || ip == 0x69bd || ip == 0x6832 || ip == 0x6840 || ip == 0x6848) && seen[ip] < 12 {
				seen[ip]++
				states = append(states, &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}})
			}
			if err := m.Step(); err != nil {
				t.Fatal(err)
			}
		}
	}
	plans := 0
	for _, v := range states {
		for _, dos := range []bool{false, true} {
			v.dos = dos
			p := planImageControl(v, 1024)
			if p == nil {
				p = planClipSetup(v, 1024)
			}
			if p == nil {
				p = planClippedRun(v, 1024)
			}
			if p == nil {
				t.Fatal("controlled clip state rejected", v.s)
			}
			compare(t, v, p)
			plans++
		}
	}
	if seen[0x6930] == 0 || seen[0x69bd] == 0 {
		t.Fatal("controlled clipping paths missing", seen)
	}
	t.Log("controlled clipped image state comparisons", plans, "states", len(states), "paths", seen)
}

func TestResearchClipCostPaths(t *testing.T) {
	for _, c := range []struct {
		left, repeat, extra bool
		run, rem            uint16
	}{{false, false, false, 2, 1}, {true, false, false, 2, 0}, {true, false, true, 2, 0}, {true, true, false, 2, 0}, {true, true, true, 2, 0}} {
		v := imageSpanSeed(t)
		v.s.IP = 0x69bd
		v.s.R[cpu.DX] = 0
		v.s.R[cpu.DI] = 3200
		v.s.R[cpu.CX] = 0
		if c.left {
			v.s.IP = 0x6930
		}
		putRowWord(v, 0x48a10+0x1230, c.rem)
		putRowWord(v, 0x48a10+0x1232, 0)
		stream := []byte{byte((c.run - 1) * 2), 3, 4}
		if c.repeat {
			stream = []byte{byte((c.run-1)*2 + 1), 0x57}
		}
		if c.extra {
			stream = append(stream, 1, 0x99)
		}
		copy(v.mem[0x50100:], stream)
		v.s.R[cpu.BX] = 0x100 + uint16(len(stream))
		p := planClippedRun(v, 100)
		m := original(v)
		t.Logf("case left%v repeat%v extra%v N%d", c.left, c.repeat, c.extra, p.Instructions)
		for n := uint64(0); n < p.Instructions; n++ {
			ip, before := m.CPU.IP, m.CPU.Cycles
			if err := m.Step(); err != nil {
				t.Fatal(err)
			}
			t.Logf("%04x +%d", ip, m.CPU.Cycles-before)
		}
	}
}
