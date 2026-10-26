// Typed Go image rows and ABI notices, spec025/#49.
package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func bitmap51At(t *testing.T, planar bool, ip uint16) *view {
	v := bitmap51RowSeed(t, 17, 1, planar, true)
	m := original(v)
	for m.Steps < 10000 {
		if m.CPU.IP == ip {
			return &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: ip, Flags: m.CPU.Flags}}
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("missing bitmap cursor", ip)
	return nil
}
func TestBitmap51Refusals(t *testing.T) {
	cases := 0
	mutations := []func(*view){func(v *view) { v.mem[0x449] = 3 }, func(v *view) { v.s.Flags |= cpu.DF }, func(v *view) { v.mem[0x48a10+0x51cc] ^= 1 }, func(v *view) { v.s.R[cpu.SP]++ }, func(v *view) { v.s.R[cpu.BP]++ }, func(v *view) { v.s.Seg[cpu.DS]++ }, func(v *view) { putRowWord(v, 0x48a10+0x11fb, 0) }, func(v *view) { putRowWord(v, 0x48a10+0x11f7, 3) }, func(v *view) { putRowWord(v, 0x48a10+0x1135, 1) }, func(v *view) { putRowWord(v, 0x48a10+0x11ff, 0) }, func(v *view) { putRowWord(v, 0x48a10+0xec3, 200) }, func(v *view) { putRowWord(v, 0x48a10+0xedd, 319) }, func(v *view) { putRowWord(v, 0x48a10+0xbd+2, 0xa001) }, func(v *view) { putRowWord(v, 0x48a10+0xbd, 0) }, func(v *view) { putRowWord(v, 0x48a10+0x11f3, 256) }}
	for _, cursor := range []struct {
		planar bool
		ip     uint16
	}{{false, 0x5373}, {false, 0x53f6}, {false, 0x53fa}, {false, 0x53fc}, {true, 0x5373}, {true, 0x53c5}, {true, 0x53d8}, {true, 0x53fc}} {
		seed := bitmap51At(t, cursor.planar, cursor.ip)
		if bitmap51(seed, 10000) == nil {
			t.Fatalf("refusal seed invalid%x", cursor.ip)
		}
		for i, mutate := range mutations {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			mutate(v)
			if bitmap51(v, 10000) != nil {
				t.Fatalf("bitmap bad cursor accepted %x mutation%d", cursor.ip, i)
			}
			cases++
		}
		if bitmap51(seed, 0) != nil {
			t.Fatal("zero budget")
		}
		cases++
	}
	for _, cursor := range []struct {
		planar bool
		ip     uint16
	}{{false, 0x53f6}, {false, 0x53fa}, {true, 0x53c5}, {true, 0x53d8}} {
		seed := bitmap51At(t, cursor.planar, cursor.ip)
		extra := []func(*view){func(v *view) { v.s.R[cpu.DI]++ }, func(v *view) { v.s.R[cpu.SI]++ }, func(v *view) { v.s.Seg[cpu.ES]++ }}
		if cursor.planar {
			extra = append(extra, func(v *view) { v.s.R[cpu.DX] ^= 1 }, func(v *view) { putRowWord(v, 0x48a10+0x11f9, 1) }, func(v *view) { putRowWord(v, 0x48a10+0x11fd, 0) })
		} else {
			extra = append(extra, func(v *view) { v.s.R[cpu.CX]++ }, func(v *view) { v.s.R[cpu.AX]++ })
		}
		for i, mutate := range extra {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			mutate(v)
			if bitmap51(v, 10000) != nil {
				t.Fatalf("bitmap span guard accepted %x mutation%d", cursor.ip, i)
			}
			cases++
		}
	}
	t.Log("bitmap51 refused invalid states", cases)
}
