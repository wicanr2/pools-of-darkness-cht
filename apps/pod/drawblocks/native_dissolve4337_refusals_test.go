// Go像素轉場，規格025/#49；VGA.DRV 0x4337，保留已驗收語意。
package drawblocks

import (
	"bytes"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func dissolve4337At(t *testing.T, ip uint16) *view {
	v := dissolve4337Seed(t, 9, 9)
	if ip == 0x45ce {
		v = dissolve4337Seed(t, 17, 9)
	}
	m := original(v)
	for m.Steps < 1000000 {
		if m.CPU.IP == ip {
			return &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}}
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("missing dissolve refusal cursor", ip)
	return nil
}
func TestNativeDissolve4337Refusals(t *testing.T) {
	cases := 0
	for _, ip := range []uint16{0x456c, 0x4569, 0x45bb, 0x45c4, 0x45ce} {
		seed := dissolve4337At(t, ip)
		if dissolve4337(seed, 10000) == nil {
			t.Fatalf("bad dissolve refusal seed%x", ip)
		}
		mutations := []func(*view){func(v *view) { v.mem[0x449] = 3 }, func(v *view) { v.mem[0x48a10+0x4337] ^= 1 }, func(v *view) { v.s.R[cpu.SP]++ }, func(v *view) { v.s.R[cpu.BP]++ }, func(v *view) { v.s.Seg[cpu.DS]++ }, func(v *view) { putRowWord(v, 0x48a10+0x1171, 0) }, func(v *view) { putRowWord(v, 0x48a10+0x1135, 0) }, func(v *view) { putRowWord(v, 0x48a10+0x1139, 0) }, func(v *view) { putRowWord(v, 0x48a10+0x1165, 0) }, func(v *view) { putRowWord(v, 0x48a10+0x1167, 0); putRowWord(v, 0x48a10+0x1169, 0) }, func(v *view) { v.mem[0x48a10+0x116b] ^= 1 }, func(v *view) { putRowWord(v, 0x48a10+0xbd+4, 0xb000) }, func(v *view) { putRowWord(v, 0x48a10+0xbd+4, 4); putRowWord(v, 0x48a10+0xcbd+4, 0xc0) }}
		for i, mutate := range mutations {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			mutate(v)
			before := append([]byte(nil), v.mem...)
			if dissolve4337(v, 10000) != nil {
				t.Fatalf("dissolve guard accepted%x mutation%d", ip, i)
			}
			if !bytes.Equal(before, v.mem) {
				t.Fatal("refusal changed input")
			}
			cases++
		}
		if dissolve4337(seed, 0) != nil {
			t.Fatal("dissolve zero budget")
		}
		cases++
		if ip == 0x4569 || ip == 0x45bb {
			for _, reg := range []int{cpu.AX, cpu.CX, cpu.SI, cpu.DI, cpu.BX} {
				v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
				v.s.R[reg]++
				if dissolve4337(v, 10000) != nil {
					t.Fatalf("dissolve pixel guard accepted%x reg%d", ip, reg)
				}
				cases++
			}
		}
	}
	seed := dissolve4337At(t, 0x456c)
	p := planDissolve4337Batch(seed, 4096)
	if p == nil || len(p.Transfers) == 0 {
		t.Fatal("batch table fixture missing")
	}
	row := uint16((p.Transfers[0].Address-0xa0000)/320) - 8
	for _, off := range []uint16{0x1252, 0x13e2} {
		v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
		v.mem[0x48a10+uint32(off+row*2)] ^= 1
		if planDissolve4337Batch(v, 4096) != nil {
			t.Fatal("bad used row table accepted", off, row)
		}
		cases++
	}
	seed = dissolve4337Seed(t, 9, 9)
	stack := uint32(seed.SS())*16 + uint32(seed.SP())
	for _, mutate := range []func(*view){func(v *view) { putRowWord(v, stack+4, 256) }, func(v *view) { putRowWord(v, stack+6, 256) }, func(v *view) { v.s.R[cpu.SP] = 0 }, func(v *view) { putRowWord(v, 0x48a10+0x8bd, 19) }} {
		v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
		mutate(v)
		if dissolve4337(v, 10000) != nil {
			t.Fatal("bad dissolve prefix accepted")
		}
		cases++
	}
	t.Log("dissolve4337 invalid state refusals", cases)
}
