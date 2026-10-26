// Typed Go integer surface scaling and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func scale5615At(t *testing.T, ip uint16) *view {
	v := scale5615Seed(t, 9, 7, 2, 25, 15)
	m := original(v)
	for m.Steps < 100000 {
		if m.CPU.IP == ip {
			return &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}}
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("missing scaling cursor", ip)
	return nil
}
func TestScale5615Refusals(t *testing.T) {
	cases := 0
	for _, ip := range []uint16{0x579c, 0x581a, 0x5813, 0x5856, 0x585a} {
		seed := scale5615At(t, ip)
		if scale5615(seed, 10000) == nil {
			t.Fatalf("bad scaling refusal seed%x", ip)
		}
		mutations := []func(*view){func(v *view) { v.mem[0x449] = 3 }, func(v *view) { v.s.Flags |= cpu.DF }, func(v *view) { v.mem[0x48a10+0x5615] ^= 1 }, func(v *view) { v.s.R[cpu.SP]++ }, func(v *view) { v.s.R[cpu.BP]++ }, func(v *view) { v.s.Seg[cpu.DS]++ }, func(v *view) { putRowWord(v, 0x48a10+0x1212, 0) }, func(v *view) { putRowWord(v, 0x48a10+0x1141, 0) }, func(v *view) { putRowWord(v, 0x48a10+0x113f, 0) }, func(v *view) { putRowWord(v, 0x48a10+0x1171, 300) }, func(v *view) { putRowWord(v, 0x48a10+0xbd+4, 0xb000) }, func(v *view) { putRowWord(v, 0x48a10+0xbd+4, 4); putRowWord(v, 0x48a10+0xcbd+4, 0xc0) }}
		for i, mutate := range mutations {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			mutate(v)
			if scale5615(v, 10000) != nil {
				t.Fatalf("scaling cursor accepted%x mutation%d", ip, i)
			}
			cases++
		}
		if scale5615(seed, 0) != nil {
			t.Fatal("scaling zero budget")
		}
		cases++
		if ip == 0x5813 || ip == 0x5856 || ip == 0x585a {
			for _, reg := range []int{cpu.AX, cpu.CX, cpu.SI, cpu.DI, cpu.DX} {
				v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
				v.s.R[reg]++
				if scale5615(v, 10000) != nil {
					t.Fatalf("scaling span accepted%x reg%d", ip, reg)
				}
				cases++
			}
		}
	}
	seed := scale5615Seed(t, 9, 7, 2, 25, 15)
	stack := uint32(seed.SS())*16 + uint32(seed.SP())
	mutations := []func(*view){func(v *view) { putRowWord(v, stack+14, 0) }, func(v *view) { putRowWord(v, stack+14, 9) }, func(v *view) { putRowWord(v, stack+8, 480) }, func(v *view) { putRowWord(v, stack+10, 300) }, func(v *view) { putRowWord(v, stack+4, 256) }, func(v *view) { putRowWord(v, stack+6, 256) }, func(v *view) { putRowWord(v, 0x48a10+0x8bd+2, 18) }, func(v *view) { v.mem[0x48a10+0x5615] ^= 1 }}
	for i, mutate := range mutations {
		v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
		mutate(v)
		if scale5615(v, 10000) != nil {
			t.Fatal("scaling prefix bad state", i)
		}
		cases++
	}
	t.Log("scale5615 invalid state refusals", cases)
}
