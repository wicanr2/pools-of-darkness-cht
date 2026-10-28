// Spec025/#49: budgeted Go memory/image spans.
package drawblocks

import (
	"encoding/json"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"os"
	"path/filepath"
	"testing"
)

func checkBoundedRLE(t *testing.T, initial *view, header uint16) (int, int) {
	t.Helper()
	m := original(initial)
	cursors, plans := 0, 0
	for n := 0; n < 20 && m.CPU.Seg[1] == 0x48a1 && m.CPU.IP != header; n++ {
		if n > 0 && (m.CPU.IP == 0x68f4 || m.CPU.IP == 0x6b33) {
			break
		}
		v := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: initial.dos}
		for _, budget := range []uint64{1, 2, 4, 7, 17} {
			p := planBoundedRLE(v, budget)
			if p == nil {
				t.Fatalf("refused %04x:%04x budget=%d regs=%+v", v.CS(), v.IP(), budget, v.s)
			}
			if !p.Algorithmic || p.Instructions > budget || p.Instructions == 0 {
				t.Fatal("invalid plan")
			}
			compare(t, v, p)
			plans++
		}
		cursors++
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if cursors == 0 {
		t.Fatal("no original instruction compared")
	}
	if m.CPU.IP != header && m.CPU.IP != 0x68f4 && m.CPU.IP != 0x6b33 {
		t.Fatal("packet did not finish")
	}
	return cursors, plans
}
func TestNativeBoundedRLEAllBoundaries(t *testing.T) {
	seed := &view{mem: input(t), s: machine.HostRegisters{Seg: [4]uint16{0xa000, 0x48a1, 0x8000, 0x5000}, Flags: 0xfed7}}
	cursors, plans := 0, 0
	for _, n := range []int{1, 2, 3, 7, 127, 128} {
		for _, repeat := range []bool{false, true} {
			stream := []byte{byte((n - 1) * 2)}
			if repeat {
				stream[0]++
				stream = append(stream, 0xa5)
			} else {
				for i := 0; i < n; i++ {
					stream = append(stream, byte(i*11))
				}
			}
			for _, l := range boundedRLELayouts() {
				for _, dos := range []bool{false, true} {
					v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
					v.s.IP = l.entry
					v.s.Seg[cpu.DS] = 0x5000
					v.s.Seg[cpu.ES] = 0xa000
					v.s.R[cpu.SI] = 0x100
					v.s.R[cpu.BX] = 0x100 + uint16(len(stream))
					v.s.R[cpu.DI] = 0x1234
					v.s.R[cpu.CX] = 0x7f
					v.s.R[cpu.AX] = 0x3700
					v.s.Flags &^= cpu.DF
					copy(v.mem[0x50100:], stream)
					c, p := checkBoundedRLE(t, v, l.header)
					cursors += c
					plans += p
				}
			}
		}
	}
	t.Logf("RLE synthetic cursors=%d bounded plans=%d", cursors, plans)
}
func TestNativeBoundedRLECapturedBoundaries(t *testing.T) {
	dir := filepath.Join(os.Getenv("POD_ROOT"), "workplace/probe/draw-intercept-v1/native-remainder-capture-v1/trace")
	raw, err := os.ReadFile(filepath.Join(dir, "fixed-entries.json"))
	if err != nil {
		t.Fatal(err)
	}
	var entries map[string]struct {
		State struct {
			R         [8]uint16
			Seg       [4]uint16
			IP, Flags uint16
			Memory    string
		}
	}
	if err = json.Unmarshal(raw, &entries); err != nil {
		t.Fatal(err)
	}
	cursors, plans := 0, 0
	for _, key := range []string{"48a1:68f4-68fb", "48a1:68fd-6907", "48a1:690b-6918", "48a1:691a-691a"} {
		c := entries[key]
		mem, e := os.ReadFile(filepath.Join(dir, c.State.Memory))
		if e != nil {
			t.Fatal(e)
		}
		for _, dos := range []bool{false, true} {
			v := &view{mem: mem, dos: dos, s: machine.HostRegisters{R: c.State.R, Seg: c.State.Seg, IP: c.State.IP, Flags: c.State.Flags}}
			n, p := checkBoundedRLE(t, v, 0x689e)
			cursors += n
			plans += p
		}
	}
	t.Logf("RLE captured cursors=%d bounded plans=%d", cursors, plans)
}
