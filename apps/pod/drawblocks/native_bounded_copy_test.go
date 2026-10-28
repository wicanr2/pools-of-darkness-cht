// Spec025/#49: budgeted Go memory/image spans.
package drawblocks

import (
	"encoding/json"
	"github.com/wicanr2/dosgolem/internal/machine"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func checkBoundedCopy(t *testing.T, initial *view) (int, int) {
	t.Helper()
	m := original(initial)
	cursors, plans := 0, 0
	for n := 0; n < 17 && m.CPU.Seg[1] == 0x1ed4; n++ {
		v := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: initial.dos}
		for budget := uint64(1); budget <= 17; budget++ {
			p := planBoundedCopy(v, budget)
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
	return cursors, plans
}
func TestNativeBoundedCopyAllBoundaries(t *testing.T) {
	mem := input(t)
	cursors, plans := 0, 0
	for _, count := range []uint16{0, 1, 2, 3, 31, 128, 255} {
		for _, pair := range [][2]uint16{{32, 32}, {8, 32}, {128, 32}, {32, 512}} {
			for _, dos := range []bool{false, true} {
				v := &view{mem: append([]byte(nil), mem...), dos: dos, s: machine.HostRegisters{R: [8]uint16{0xab77, 23, 45, 12, 0xe000, 0xd000, 0x440, 0x880}, Seg: [4]uint16{0x6000, 0x1ed4, 0x8000, 0x7000}, IP: 0x1a5c, Flags: 0xfed7}}
				for off, val := range map[uint32]uint16{0: 0x1234, 2: 0x1111, 4: count, 6: pair[1], 8: 0x6000, 10: pair[0], 12: 0x6000} {
					a := uint32(v.SS())*16 + uint32(v.SP()) + off
					v.mem[a] = byte(val)
					v.mem[a+1] = byte(val >> 8)
				}
				c, p := checkBoundedCopy(t, v)
				cursors += c
				plans += p
			}
		}
	}
	t.Logf("copy synthetic cursors=%d bounded plans=%d", cursors, plans)
}
func TestNativeBoundedCopyCapturedBoundaries(t *testing.T) {
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
	for key, c := range entries {
		if !strings.HasPrefix(key, "1ed4:") {
			continue
		}
		mem, e := os.ReadFile(filepath.Join(dir, c.State.Memory))
		if e != nil {
			t.Fatal(e)
		}
		for _, dos := range []bool{false, true} {
			v := &view{mem: mem, dos: dos, s: machine.HostRegisters{R: c.State.R, Seg: c.State.Seg, IP: c.State.IP, Flags: c.State.Flags}}
			n, p := checkBoundedCopy(t, v)
			cursors += n
			plans += p
		}
	}
	t.Logf("copy captured cursors=%d bounded plans=%d", cursors, plans)
}
