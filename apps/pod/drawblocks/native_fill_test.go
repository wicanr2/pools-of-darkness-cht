package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func rectangleCases(t *testing.T) []*view {
	t.Helper()
	var seed *view
	for _, v := range drawCalls(t) {
		if v.CS() == 0x48a1 && v.IP() == 0x41af {
			seed = v
		}
	}
	if seed == nil {
		t.Fatal("normal rectangle entry absent")
	}
	var cases []*view
	for _, shape := range [][3]int{{1, 1, 0}, {2, 1, 0}, {3, 2, 0}, {17, 9, 0}, {320, 200, 0}, {64, 48, 0}, {3, 2, 1}, {17, 9, 2}, {17, 9, 3}} {
		for _, dos := range []bool{false, true} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
			set := func(a uint32, x uint16) { v.mem[a] = byte(x); v.mem[a+1] = byte(x >> 8) }
			desc := func(slot int, values []uint16) {
				for i, x := range values {
					set(0x48a10+0xbd+uint32(slot*2+i*512), x)
				}
			}
			w, h, links := shape[0], shape[1], shape[2]
			desc(0, []uint16{0xa000, 4000, 0, 0, uint16(w - 1), uint16(h - 1), 32})
			if links > 0 {
				desc(0, []uint16{0xa000, 4000, 0, 0, 319, 199, 32})
				for i := 1; i <= links; i++ {
					desc(i, []uint16{uint16((i - 1) * 2), 7, 5, 6, uint16(5 + w - 1), uint16(6 + h - 1), 64})
				}
			}
			if links == 3 {
				desc(0, []uint16{0xa000, 4000, 10, 20, 73, 67, 32})
				for i := 1; i <= links; i++ {
					desc(i, []uint16{uint16((i - 1) * 2), 7, 12, 21, 28, 29, 64})
				}
				v.s.Flags |= 0x400
			}
			stack := uint32(v.SS())*16 + uint32(v.SP())
			set(stack+4, uint16(links))
			set(stack+6, 0x173c)
			cases = append(cases, v)
		}
	}
	return cases
}

func TestNativeRectangleRefusesUnprovedInputs(t *testing.T) {
	seed := rectangleCases(t)[0]
	for _, kind := range []string{"budget", "fingerprint", "mode", "slot", "cycle", "surface", "bounds", "stack"} {
		t.Run(kind, func(t *testing.T) {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: seed.dos}
			budget := uint64(1000000)
			set := func(a uint32, x uint16) { v.mem[a] = byte(x); v.mem[a+1] = byte(x >> 8) }
			stack := uint32(v.SS())*16 + uint32(v.SP())
			switch kind {
			case "budget":
				budget = 85
			case "fingerprint":
				v.mem[0x48a10+0x41af] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "slot":
				set(stack+4, 256)
			case "cycle":
				set(0x48a10+0xcbd, 64)
				set(0x48a10+0xbd, 0)
			case "surface":
				set(0x48a10+0xbd, 0x6000)
			case "bounds":
				set(0x48a10+0x8bd, 320)
			case "stack":
				v.s.R[4] = 8
			}
			if planRectangle(v, budget) != nil {
				t.Fatal("unproved input accepted")
			}
		})
	}
}

func TestNativeRectangleOracle(t *testing.T) {
	cases, plans := 0, 0
	for _, v := range rectangleCases(t) {
		for _, budget := range []uint64{86, 128, 1024, 1000000} {
			p := planRectangle(v, budget)
			if p == nil {
				continue
			}
			if !p.Algorithmic {
				t.Fatal("native rectangle missing algorithmic category")
			}
			compare(t, v, p)
			plans++
		}
		m := original(v)
		stack := uint32(v.SS())*16 + uint32(v.SP())
		retIP, retCS := v.Read16(stack), v.Read16(stack+2)
		for n := 0; n < 1000; n++ {
			state := machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}
			w := &view{mem: append([]byte(nil), m.Mem...), s: state, dos: v.dos}
			p := planRectangle(w, 128)
			if p == nil {
				t.Fatalf("no native rectangle continuation at %04x:%04x", w.CS(), w.IP())
			}
			compare(t, w, p)
			plans++
			if err := m.CommitHostBlock(p); err != nil {
				t.Fatal(err)
			}
			if m.CPU.IP == retIP && m.CPU.Seg[1] == retCS {
				break
			}
			if n == 999 {
				t.Fatal("rectangle did not finish")
			}
		}
		cases++
	}
	t.Log("complete original geometries/cost cases", cases, "independent plan comparisons", plans)
}
