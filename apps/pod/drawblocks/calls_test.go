package drawblocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dosgolem/internal/machine"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"
)

func drawCalls(t *testing.T) []*view {
	t.Helper()
	root := os.Getenv("POD_DRAW_CALLS")
	if root == "" {
		t.Skip("normal original entries absent")
	}
	raw, err := os.ReadFile(filepath.Join(root, "calls.json"))
	if err != nil {
		t.Fatal(err)
	}
	var j struct {
		DrawCaptures []struct {
			R         [8]uint16
			Seg       [4]uint16
			IP, Flags uint16
			Memory    string
		}
	}
	if err = json.Unmarshal(raw, &j); err != nil {
		t.Fatal(err)
	}
	var out []*view
	for _, c := range j.DrawCaptures {
		if !(c.Seg[1] == 0x1ed4 || c.Seg[1] == 0x1d1f || c.Seg[1] == 0x48a1 && (c.IP == 0x6733 || c.IP == 0x4944 || c.IP == 0x41af)) {
			continue
		}
		m, err := os.ReadFile(filepath.Join(root, c.Memory))
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, &view{mem: m, s: machine.HostRegisters{R: c.R, Seg: c.Seg, IP: c.IP, Flags: c.Flags}})
	}
	if len(out) != 6 {
		t.Fatal("real call count", len(out))
	}
	return out
}
func callMachine(v *view) *machine.Machine {
	m := original(v)
	m.SetNativeVideo(true)
	canvas := make([]byte, 64000)
	m.ObserveVideoWrites(func(w machine.VideoWrite) { _ = w.Value })
	m.ObserveVideoCommands(func(c machine.VideoCommand) {
		for i := 0; i < int(c.Count); i++ {
			off := int(c.Offset) + i*int(c.Width)
			if c.Backward {
				off = int(c.Offset) - i*int(c.Width)
			}
			for b := 0; b < int(c.Width); b++ {
				v := byte(c.Value >> uint(b*8))
				if c.Kind == machine.VideoCopy {
					v = c.Data[i*int(c.Width)+b]
				}
				canvas[off+b] = v
			}
		}
	})
	m.RunLimit = 1000000
	return m
}
func runCall(t *testing.T, m *machine.Machine, v *view, native bool) {
	t.Helper()
	ip := v.Read16(uint32(v.SS())*16 + uint32(v.SP()))
	cs := v.Read16(uint32(v.SS())*16 + uint32(v.SP()+2))
	used := false
	for m.Steps < 1000000 {
		if used && m.CPU.IP == ip && m.CPU.Seg[1] == cs {
			return
		}
		used = true
		if native {
			state := machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}
			w := &view{mem: m.Mem, s: state, dos: v.dos}
			if p := Plan(w, 1024); p != nil {
				if err := m.CommitHostBlock(p); err != nil {
					t.Fatal(err)
				}
				continue
			}
		}
		if err := m.Step(); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("whole function did not return")
}
func TestActualWholeDrawingCalls(t *testing.T) {
	cases := 0
	for _, v := range drawCalls(t) {
		for _, dos := range []bool{false, true} {
			v.dos = dos
			a, b := callMachine(v), callMachine(v)
			runCall(t, a, v, false)
			runCall(t, b, v, true)
			if a.CPU.R != b.CPU.R || a.CPU.Seg != b.CPU.Seg || a.CPU.IP != b.CPU.IP || a.CPU.Flags != b.CPU.Flags || a.CPU.Cycles != b.CPU.Cycles || a.Steps != b.Steps || !bytes.Equal(a.Mem, b.Mem) {
				t.Fatalf("whole call differs %04x:%04x", v.CS(), v.IP())
			}
			if a.DAC != b.DAC || !reflect.DeepEqual(a.VGA, b.VGA) || !reflect.DeepEqual(a.Ports, b.Ports) || !reflect.DeepEqual(a.PortLog, b.PortLog) {
				t.Fatalf("whole call device state/port log differs %04x:%04x", v.CS(), v.IP())
			}
			cases++
		}
	}
	t.Log("actual full functions including RAM, return and time", cases)
}
func TestPairedDrawingCost(t *testing.T) {
	out := os.Getenv("POD_DRAW_COST")
	if out == "" {
		t.Skip("explicit local measurement only")
	}
	rows := []map[string]any{}
	for _, v := range drawCalls(t) {
		old, newer := []int64{}, []int64{}
		var steps, host uint64
		for group := 0; group < 7; group++ {
			for _, mode := range []bool{group%2 == 0, group%2 != 0} {
				var elapsed time.Duration
				m := callMachine(v)
				for n := 0; n < 32; n++ {
					copy(m.Mem, v.mem)
					m.CPU.R = v.s.R
					m.CPU.Seg = v.s.Seg
					m.CPU.IP = v.s.IP
					m.CPU.Flags = v.s.Flags
					m.CPU.Cycles = 0
					m.CPU.Halted = false
					m.Steps = 0
					m.HostBlockInstructions = 0
					start := time.Now()
					runCall(t, m, v, mode)
					elapsed += time.Since(start)
					steps = m.Steps
					if mode {
						host = m.HostBlockInstructions
					}
				}
				if mode {
					newer = append(newer, elapsed.Nanoseconds()/32)
				} else {
					old = append(old, elapsed.Nanoseconds()/32)
				}
			}
		}
		sort.Slice(old, func(i, j int) bool { return old[i] < old[j] })
		sort.Slice(newer, func(i, j int) bool { return newer[i] < newer[j] })
		row := map[string]any{"cs": v.CS(), "ip": v.IP(), "virtual_steps": steps, "host_instructions": host, "original_ns": old[3], "native_ns": newer[3], "ratio": float64(old[3]) / float64(newer[3]), "original_batches": old, "native_batches": newer}
		rows = append(rows, row)
		t.Log(fmt.Sprintf("%04x:%04x %d -> %d ns %.2fx", v.CS(), v.IP(), old[3], newer[3], float64(old[3])/float64(newer[3])))
	}
	b, err := json.MarshalIndent(map[string]any{"scope": "six real complete functions; includes source guards, compatibility writes and canvas commands; excludes Owner and final RGB/FPS", "groups": 7, "calls_per_group": 32, "rows": rows}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(out, b, 0600); err != nil {
		t.Fatal(err)
	}
}
