// Spec025/#49: full original-machine and negative controls.
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func boundedMemorySeedTest(t *testing.T, kind string) *view {
	v := &view{mem: input(t), s: machine.HostRegisters{R: [8]uint16{0x3700, 0x7f, 45, 0x102, 0xe000, 0xd000, 0x100, 0x1234}, Seg: [4]uint16{0xa000, 0x1d1f, 0x8000, 0x5000}, IP: 0x46e, Flags: 0xfed7}}
	if kind == "copy" {
		v.s.Seg[cpu.CS] = 0x1ed4
		v.s.IP = 0x1a5c
	}
	if kind == "rle" {
		v.s.Seg[cpu.CS] = 0x48a1
		v.s.IP = 0x68f4
		v.s.Flags &^= cpu.DF
		v.mem[0x50100] = 1
		v.mem[0x50101] = 0x55
		return v
	}
	for off, val := range map[uint32]uint16{0: 0x1234, 2: 0x1111, 4: 31, 6: 32, 8: 0xa000, 10: 8, 12: 0x6000} {
		a := uint32(v.SS())*16 + uint32(v.SP()) + off
		v.mem[a] = byte(val)
		v.mem[a+1] = byte(val >> 8)
	}
	return v
}
func TestNativeBoundedRefusals(t *testing.T) {
	cases := 0
	for _, kind := range []string{"zero", "copy", "rle"} {
		for _, bad := range []string{"budget", "mode", "signature", "unknown-entry", "range", "source-or-stack"} {
			v := boundedMemorySeedTest(t, kind)
			if boundedMemoryPlanTest(v, 100) == nil {
				t.Fatal("valid baseline refused", kind)
			}
			budget := uint64(100)
			switch bad {
			case "budget":
				budget = 0
			case "mode":
				v.mem[0x449] = 0xd
			case "signature":
				v.mem[uint32(v.CS())*16+uint32(v.IP())] ^= 1
			case "unknown-entry":
				v.s.IP++
			case "range":
				if kind == "rle" {
					v.s.R[cpu.DI] = 64000
				} else {
					a := uint32(v.SS())*16 + uint32(v.SP()) + 6
					v.mem[a] = 0xff
					v.mem[a+1] = 0xff
				}
			case "source-or-stack":
				if kind == "rle" {
					v.s.R[cpu.BX] = v.SI()
				} else {
					v.s.R[cpu.SP] = 0xfff8
				}
			}
			before := append([]byte(nil), v.mem...)
			state := v.s
			if p := boundedMemoryPlanTest(v, budget); p != nil {
				t.Fatalf("unproved state accepted %s %s", kind, bad)
			}
			if !bytes.Equal(before, v.mem) || state != v.s {
				t.Fatal("refusal mutated input")
			}
			cases++
		}
	}
	t.Logf("bounded refusal controls=%d unchanged inputs", cases)
}
func TestNativeBoundedLargeSpans(t *testing.T) {
	cases := 0
	for _, kind := range []string{"zero", "copy"} {
		for _, dos := range []bool{false, true} {
			v := boundedMemorySeedTest(t, kind)
			v.dos = dos
			stack := uint32(v.SS())*16 + uint32(v.SP())
			for off, val := range map[uint32]uint16{4: 64000, 6: 0, 10: 0} {
				v.mem[stack+off] = byte(val)
				v.mem[stack+off+1] = byte(val >> 8)
			}
			p := boundedMemoryPlanTest(v, 100)
			if p == nil {
				t.Fatal("large span refused")
			}
			compare(t, v, p)
			cases++
		}
	}
	t.Logf("full-screen spans original comparisons=%d", cases)
}
