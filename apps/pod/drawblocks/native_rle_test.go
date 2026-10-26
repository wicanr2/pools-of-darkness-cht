package drawblocks

import (
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

func rleSeed(t *testing.T) *view {
	t.Helper()
	for _, v := range drawCalls(t) {
		if v.CS() == 0x48a1 && v.IP() == 0x6733 {
			return v
		}
	}
	t.Fatal("normal original image entry absent")
	return nil
}

func TestNativeRLEPacketsOriginal(t *testing.T) {
	seed := rleSeed(t)
	var streams [][]byte
	for _, n := range []int{1, 2, 3, 127, 128} {
		literal := []byte{byte((n - 1) * 2)}
		for i := 0; i < n; i++ {
			literal = append(literal, byte(i*11))
		}
		streams = append(streams, literal, []byte{byte((n-1)*2 + 1), 0xa5})
	}
	streams = append(streams, []byte{1, 0xff, 4, 1, 2, 3, 3, 0x80, 0, 0x17})
	cases, plans := 0, 0
	for _, stream := range streams {
		for _, dos := range []bool{false, true} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
			v.s.IP = 0x68f4
			v.s.Seg[cpu.DS] = 0x5000
			v.s.Seg[cpu.ES] = 0xa000
			v.s.R[cpu.SI] = 0x100
			v.s.R[cpu.BX] = 0x100 + uint16(len(stream))
			v.s.R[cpu.DI] = 0x1234
			v.s.R[cpu.CX] = 0x7f
			v.s.R[cpu.AX] = 0x3700
			v.s.Flags &^= cpu.DF
			copy(v.mem[0x50100:], stream)
			for _, budget := range []uint64{11, 12, 13, 14, 16, 64, 1024} {
				if p := planRLERuns(v, budget); p != nil {
					compare(t, v, p)
					plans++
				}
			}
			cases++
		}
	}
	t.Log("RLE packet original cases", cases, "plan comparisons", plans)
}

func TestNativeRLENormalImageStates(t *testing.T) {
	seed := rleSeed(t)
	m := original(seed)
	states := []*view{}
	for m.Steps < 1000000 && len(states) < 32 {
		if m.CPU.Seg[cpu.CS] == 0x48a1 && m.CPU.IP == 0x68f4 {
			states = append(states, &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}})
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if len(states) != 32 {
		t.Fatal("normal RLE states missing")
	}
	plans := 0
	for _, v := range states {
		for _, dos := range []bool{false, true} {
			v.dos = dos
			for _, budget := range []uint64{16, 64, 1024} {
				p := planRLERuns(v, budget)
				if p == nil {
					t.Fatal("normal RLE plan rejected")
				}
				compare(t, v, p)
				plans++
			}
		}
	}
	t.Log("normal original image RLE state comparisons", plans)
}

func TestNativeRLERefusals(t *testing.T) {
	seed := rleSeed(t)
	for _, kind := range []string{"budget", "fingerprint", "mode", "direction", "surface", "source", "packet", "destination", "counter"} {
		t.Run(kind, func(t *testing.T) {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
			v.s.IP = 0x68f4
			v.s.Seg[cpu.DS] = 0x5000
			v.s.Seg[cpu.ES] = 0xa000
			v.s.R[cpu.SI] = 0x100
			v.s.R[cpu.BX] = 0x102
			v.s.R[cpu.DI] = 100
			v.s.R[cpu.CX] = 0
			v.s.Flags &^= cpu.DF
			v.mem[0x50100] = 1
			v.mem[0x50101] = 0x55
			budget := uint64(100)
			switch kind {
			case "budget":
				budget = 13
			case "fingerprint":
				v.mem[0x48a10+0x68f4] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "direction":
				v.s.Flags |= cpu.DF
			case "surface":
				v.s.Seg[cpu.ES] = 0x9000
			case "source":
				v.s.Seg[cpu.DS] = 0xa000
			case "packet":
				v.s.R[cpu.BX] = 0x101
			case "destination":
				v.s.R[cpu.DI] = 64000
			case "counter":
				v.s.R[cpu.CX] = 0x100
			}
			if planRLERuns(v, budget) != nil {
				t.Fatal("unproved RLE input accepted")
			}
		})
	}
}
