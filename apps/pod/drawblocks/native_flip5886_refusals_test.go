// Typed Go in-place surface flips and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func flip5886At(t *testing.T, ip uint16) *view {
	h := true
	if ip == 0x59cf || ip == 0x59d3 || ip == 0x59f8 || ip == 0x59fc {
		h = false
	}
	v := flip5886Seed(t, 9, 7, h, true)
	m := original(v)
	for m.Steps < 100000 {
		if m.CPU.IP == ip {
			return &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}}
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("missing flip refusal cursor", ip)
	return nil
}
func TestFlip5886Refusals(t *testing.T) {
	cases := 0
	for _, ip := range []uint16{0x5947, 0x59a9, 0x59c8, 0x59cf, 0x59d3, 0x59f1, 0x59f8, 0x59fc, 0x5a0c, 0x5a10} {
		seed := flip5886At(t, ip)
		if flip5886(seed, 10000) == nil {
			t.Fatalf("invalid flip refusal seed%x", ip)
		}
		mutations := []func(*view){func(v *view) { v.mem[0x449] = 3 }, func(v *view) { v.s.Flags |= cpu.DF }, func(v *view) { v.mem[0x48a10+0x5886] ^= 1 }, func(v *view) { v.s.R[cpu.SP]++ }, func(v *view) { v.s.R[cpu.BP]++ }, func(v *view) { v.s.Seg[cpu.DS]++ }, func(v *view) { v.s.Seg[cpu.ES]++ }, func(v *view) { putRowWord(v, 0x48a10+0x1171, 0) }, func(v *view) { putRowWord(v, 0x48a10+0xecb, 0) }, func(v *view) { putRowWord(v, 0x48a10+0xec3, 11) }, func(v *view) { putRowWord(v, 0x48a10+0xbd+4, 0xb000) }, func(v *view) { putRowWord(v, 0x48a10+0xbd+4, 4); putRowWord(v, 0x48a10+0xcbd+4, 0xc0) }}
		for i, mutate := range mutations {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			mutate(v)
			if flip5886(v, 10000) != nil {
				t.Fatalf("flip guard accepted%x mutation%d", ip, i)
			}
			cases++
		}
		if flip5886(seed, 0) != nil {
			t.Fatal("flip zero budget")
		}
		cases++
		if ip != 0x5947 && ip != 0x59a9 {
			for _, reg := range []int{cpu.CX, cpu.SI, cpu.DI, cpu.BX, cpu.DX} {
				v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
				v.s.R[reg]++
				if flip5886(v, 10000) != nil {
					t.Fatalf("flip span accepted%x reg%d", ip, reg)
				}
				cases++
			}
		}
	}
	seed := flip5886Seed(t, 9, 7, true, true)
	stack := uint32(seed.SS())*16 + uint32(seed.SP())
	mutations := []func(*view){func(v *view) { putRowWord(v, stack+4, 256) }, func(v *view) { putRowWord(v, 0x48a10+0xbd+4, 0xb000) }, func(v *view) { putRowWord(v, 0x48a10+0x8bd, 19) }, func(v *view) { v.mem[0x48a10+0x5886] ^= 1 }}
	for i, mutate := range mutations {
		v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
		mutate(v)
		if flip5886(v, 10000) != nil {
			t.Fatal("flip prefix guard accepted", i)
		}
		cases++
	}
	t.Log("flip5886 invalid state refusals", cases)
}
