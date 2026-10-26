package drawblocks

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/wicanr2/dosgolem/internal/machine"
)

// Independent original-CPU evidence for spec025/#49, never a product path.
func TestOriginalRectangleEvidence(t *testing.T) {
	out := os.Getenv("POD_FILL_ORACLE")
	if out == "" {
		t.Skip("explicit local original-CPU export only")
	}
	var seed *view
	for _, v := range drawCalls(t) {
		if v.CS() == 0x48a1 && v.IP() == 0x41af {
			seed = v
		}
	}
	if seed == nil {
		t.Fatal("normal original rectangle entry absent")
	}
	type state struct {
		Steps, Cycles uint64
		CPU           machine.HostRegisters
	}
	type result struct {
		Name          string
		DOS           bool
		Width, Height int
		Rows          []state
		After         state
		Writes        []machine.HostWrite
	}
	results := []result{}
	for _, size := range [][2]int{{1, 1}, {2, 1}, {3, 2}, {17, 9}, {320, 200}, {64, 48}} {
		for _, dos := range []bool{false, true} {
			v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s, dos: dos}
			base := uint32(0x48a10)
			word := func(a uint32, x uint16) { v.mem[a] = byte(x); v.mem[a+1] = byte(x >> 8) }
			for n, x := range []uint16{0xa000, 4000, 0, 0, uint16(size[0] - 1), uint16(size[1] - 1), 32} {
				word(base+0xbd+uint32(n)*512, x)
			}
			stack := uint32(v.SS())*16 + uint32(v.SP())
			word(stack+4, 0)
			word(stack+6, 0x173c)
			m := original(v)
			r := result{Name: "root", DOS: dos, Width: size[0], Height: size[1]}
			m.WatchWrite(0, 0xfffff, func(m *machine.Machine, a uint32, old, b byte) {
				cs, ip := m.CPU.OpAddr()
				r.Writes = append(r.Writes, machine.HostWrite{Address: a, Value: b, Step: m.Steps, CS: cs, IP: ip})
			})
			readState := func() state {
				return state{m.Steps, m.CPU.Cycles, machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}}
			}
			retIP, retCS := v.Read16(stack), v.Read16(stack+2)
			for m.Steps < 100000 {
				if m.CPU.Seg[1] == 0x48a1 && m.CPU.IP == 0x4268 {
					r.Rows = append(r.Rows, readState())
				}
				if m.Steps > 0 && m.CPU.Seg[1] == retCS && m.CPU.IP == retIP {
					break
				}
				if err := m.Step(); err != nil {
					t.Fatal(err)
				}
			}
			if m.Steps >= 100000 {
				t.Fatal("rectangle did not return")
			}
			r.After = readState()
			results = append(results, r)
		}
	}
	b, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(out, append(b, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	t.Log("independent original cases", len(results))
}
