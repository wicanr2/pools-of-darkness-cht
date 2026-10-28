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

func checkBoundedZero(t *testing.T, initial *view) (int, int) {
	t.Helper()
	m := original(initial)
	cursors, plans := 0, 0
	for n := 0; n < 12 && m.CPU.Seg[1] == 0x1d1f; n++ {
		v := &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}, dos: initial.dos}
		for budget := uint64(1); budget <= 12; budget++ {
			p := planBoundedZero(v, budget)
			if p == nil {
				t.Fatalf("refused cs:ip=%04x:%04x budget=%d regs=%+v", v.CS(), v.IP(), budget, v.s)
			}
			if !p.Algorithmic || p.Instructions > budget || p.Instructions == 0 {
				t.Fatal("invalid bounded plan")
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

func TestNativeBoundedZeroAllBoundaries(t *testing.T) {
	mem := input(t)
	cursors, plans := 0, 0
	for _, count := range []uint16{0, 1, 2, 3, 7, 64, 255, 256, 511} {
		for _, dos := range []bool{false, true} {
			v := &view{mem: append([]byte(nil), mem...), dos: dos, s: machine.HostRegisters{R: [8]uint16{0xab77, 23, 45, 12, 0xe000, 0xd000, 0x440, 0x880}, Seg: [4]uint16{0x6000, 0x1d1f, 0x8000, 0x6000}, IP: 0x46e, Flags: 0xfed7}}
			for off, val := range map[uint32]uint16{0: 0x1234, 2: 0x1111, 4: count, 6: 32, 8: 0xa000} {
				a := uint32(v.SS())*16 + uint32(v.SP()) + off
				v.mem[a] = byte(val)
				v.mem[a+1] = byte(val >> 8)
			}
			c, p := checkBoundedZero(t, v)
			cursors += c
			plans += p
		}
	}
	t.Logf("zero synthetic cursors=%d bounded plans=%d", cursors, plans)
}

func TestNativeBoundedZeroCapturedBoundaries(t *testing.T) {
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
		if !strings.HasPrefix(key, "1d1f:") {
			continue
		}
		mem, e := os.ReadFile(filepath.Join(dir, c.State.Memory))
		if e != nil {
			t.Fatal(e)
		}
		for _, dos := range []bool{false, true} {
			v := &view{mem: mem, dos: dos, s: machine.HostRegisters{R: c.State.R, Seg: c.State.Seg, IP: c.State.IP, Flags: c.State.Flags}}
			n, p := checkBoundedZero(t, v)
			cursors += n
			plans += p
		}
	}
	t.Logf("zero captured cursors=%d bounded plans=%d", cursors, plans)
}
