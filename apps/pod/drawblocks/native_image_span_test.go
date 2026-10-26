package drawblocks

import (
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func imageSpanSeed(t *testing.T) *view {
	seed := rleSeed(t)
	v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
	v.s.IP = 0x689e
	v.s.Seg[cpu.DS] = 0x5000
	v.s.Seg[cpu.SS] = 0x7000
	v.s.Seg[cpu.ES] = 0xa000
	v.s.R[cpu.BP] = 0x800
	v.s.R[cpu.SI] = 0x100
	v.s.R[cpu.CX] = 0x5a
	v.s.Flags &^= cpu.DF
	for a, n := range map[int]uint16{0x48a10 + 0x1250: 0, 0x48a10 + 0xec1: 0, 0x48a10 + 0xec5: 319, 0x48a10 + 0x120c: 3200, 0x70808: 0} {
		v.mem[a], v.mem[a+1] = byte(n), byte(n>>8)
	}
	copy(v.mem[0x50100:], []byte{0, 0x80, 17, 4, 1, 3, 1, 4})
	return v
}

func TestNativeImageSpanOriginal(t *testing.T) {
	plans := 0
	for _, x := range []uint16{0, 1, 319} {
		for _, width := range []byte{1, 2, 17, 255} {
			if uint32(x)+uint32(width) > 320 {
				continue
			}
			for _, dos := range []bool{false, true} {
				v := imageSpanSeed(t)
				v.dos = dos
				v.mem[0x50100], v.mem[0x50101], v.mem[0x50102] = byte(x), byte(x>>8)|0x80, width
				p := planImageSpan(v, 26)
				if p == nil {
					t.Fatal("valid image span rejected")
				}
				compare(t, v, p)
				plans++
			}
		}
	}
	t.Log("image span synthetic original comparisons", plans)
}

func TestNativeImageSpanNormalStates(t *testing.T) {
	m := original(rleSeed(t))
	states := []*view{}
	for m.Steps < 1000000 && len(states) < 32 {
		if m.CPU.Seg[cpu.CS] == 0x48a1 && m.CPU.IP == 0x689e && m.Peek16(0x48a10+0x1250) < 0x8000 {
			states = append(states, &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}})
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if len(states) != 32 {
		t.Fatal("normal image span states absent")
	}
	plans := 0
	for _, v := range states {
		for _, dos := range []bool{false, true} {
			v.dos = dos
			p := planImageSpan(v, 26)
			if p == nil {
				t.Fatal("normal image span rejected")
			}
			compare(t, v, p)
			plans++
		}
	}
	t.Log("image span normal original comparisons", plans)
}

func TestNativeImageSpanRefusals(t *testing.T) {
	for _, kind := range []string{"budget", "code", "mode", "direction", "surface", "row-marker", "source", "clip", "destination", "zero-width", "truncated", "source-alias", "argument-alias", "stack", "bounds", "zero-payload"} {
		t.Run(kind, func(t *testing.T) {
			v := imageSpanSeed(t)
			budget := uint64(26)
			switch kind {
			case "budget":
				budget = 25
			case "code":
				v.mem[0x48a10+0x689e] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "direction":
				v.s.Flags |= cpu.DF
			case "surface":
				v.s.Seg[cpu.ES] = 0x9000
			case "row-marker":
				v.mem[0x48a10+0x1251] = 0x80
			case "source":
				v.s.Seg[cpu.DS] = 0xa000
			case "clip":
				v.mem[0x50100], v.mem[0x50101] = 0x3f, 0x81
			case "destination":
				v.mem[0x48a10+0x120c], v.mem[0x48a10+0x120d] = 0xff, 0xf9
			case "zero-width":
				v.mem[0x50102] = 0
			case "truncated":
				v.s.R[cpu.SI] = 0xfff0
				copy(v.mem[0x5fff0:], []byte{0, 0x80, 17, 32})
			case "source-alias":
				v.s.Seg[cpu.DS], v.s.R[cpu.SI] = 0x48a1, 0x1250
				copy(v.mem[0x48a10+0x1250:], []byte{0, 0, 17, 4})
			case "argument-alias":
				v.s.Seg[cpu.SS], v.s.R[cpu.BP] = 0x48a1, 0x1248
			case "stack":
				v.s.R[cpu.BP] = 65530
			case "bounds":
				v.mem[0x48a10+0xec1], v.mem[0x48a10+0xec2] = 0x40, 1
			case "zero-payload":
				v.mem[0x50103] = 0
			}
			if planImageSpan(v, budget) != nil {
				t.Fatal("unproved image span accepted")
			}
		})
	}
}
