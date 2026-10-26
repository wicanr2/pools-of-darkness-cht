package drawblocks

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func compareResearchPalette(t *testing.T, v *view, phase int) {
	t.Helper()
	j := planPaletteOutput(v, 16)
	if j == nil {
		t.Fatal("valid RGB job rejected")
	}
	a, b := original(v), original(v)
	for _, m := range []*machine.Machine{a, b} {
		m.Out8(0x3c8, 17)
		for n := 0; n < phase; n++ {
			m.Out8(0x3c9, byte(n+11))
		}
		m.Out8(0x3c7, 8)
		m.In8(0x3c9)
		m.PortLog = nil
	}
	for n := 0; n < 16; n++ {
		if err := a.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if err := b.CommitHostBlock(j); err != nil {
		t.Fatal(err)
	}
	if a.CPU.R != b.CPU.R || a.CPU.Seg != b.CPU.Seg || a.CPU.IP != b.CPU.IP || a.CPU.Flags != b.CPU.Flags || a.CPU.Cycles != b.CPU.Cycles || a.Steps != b.Steps || !bytes.Equal(a.Mem, b.Mem) || a.DAC != b.DAC || !reflect.DeepEqual(a.VGA, b.VGA) || !reflect.DeepEqual(a.Ports, b.Ports) || !reflect.DeepEqual(a.PortLog, b.PortLog) {
		t.Fatalf("RGB job differs dos%v state want%+v got%+v cycles %d/%d ports want%+v got%+v", v.dos, machine.HostRegisters{R: a.CPU.R, Seg: a.CPU.Seg, IP: a.CPU.IP, Flags: a.CPU.Flags}, j.After, a.CPU.Cycles, b.CPU.Cycles, a.PortLog, b.PortLog)
	}
	for _, m := range []*machine.Machine{a, b} {
		m.Out8(0x3c9, 23)
	}
	if a.DAC != b.DAC || a.In8(0x3c9) != b.In8(0x3c9) {
		t.Fatal("DAC phase/read state differs")
	}
}

func TestNativePaletteOutputOriginal(t *testing.T) {
	plans := 0
	for _, index := range []uint16{0, 1, 127, 255} {
		for _, rgb := range [][3]uint16{{0, 0, 0}, {63, 63, 63}, {1, 17, 31}, {63, 0, 63}} {
			for _, flags := range []uint16{0xf002, 0xfad7} {
				for _, dos := range []bool{false, true} {
					for phase := 0; phase < 3; phase++ {
						v := imageSpanSeed(t)
						v.s.IP = 0x2aa8
						v.s.R[cpu.SP] = 0x8000
						v.s.Flags = flags
						v.dos = dos
						for i, n := range []uint16{0x9999, 0x1234, index, rgb[0], rgb[1], rgb[2]} {
							putRowWord(v, 0x78000+uint32(i*2), n)
						}
						compareResearchPalette(t, v, phase)
						plans++
					}
				}
			}
		}
	}
	t.Log("RGB job original comparisons", plans)
}

func TestNativePaletteOutputEntry(t *testing.T) {
	root := os.Getenv("POD_DRAW_CALLS")
	if root == "" {
		t.Skip("original entry fixture absent")
	}
	raw, err := os.ReadFile(filepath.Join(root, "calls.json"))
	if err != nil {
		t.Fatal(err)
	}
	var table struct {
		DrawCaptures []struct {
			R         [8]uint16
			Seg       [4]uint16
			IP, Flags uint16
			Memory    string
		}
	}
	if err = json.Unmarshal(raw, &table); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range table.DrawCaptures {
		if row.Seg[cpu.CS] != 0x48a1 || row.IP != 0x2aa8 {
			continue
		}
		found = true
		mem, err := os.ReadFile(filepath.Join(root, row.Memory))
		if err != nil {
			t.Fatal(err)
		}
		for _, dos := range []bool{false, true} {
			for phase := 0; phase < 3; phase++ {
				v := &view{mem: mem, s: machine.HostRegisters{R: row.R, Seg: row.Seg, IP: row.IP, Flags: row.Flags}, dos: dos}
				compareResearchPalette(t, v, phase)
			}
		}
	}
	if !found {
		t.Fatal("original RGB entry absent")
	}
	t.Log("RGB job original entry comparisons 6")
}

func TestNativePaletteOutputRefusals(t *testing.T) {
	for _, kind := range []string{"budget", "code", "mode", "entry", "segment", "stack-low", "stack-wrap", "stack-video", "stack-system", "stack-driver", "index", "red", "green", "blue"} {
		t.Run(kind, func(t *testing.T) {
			v := imageSpanSeed(t)
			v.s.IP = 0x2aa8
			v.s.R[cpu.SP] = 0x8000
			budget := uint64(16)
			for i, n := range []uint16{0x9999, 0x1234, 1, 2, 3, 4} {
				putRowWord(v, 0x78000+uint32(i*2), n)
			}
			switch kind {
			case "budget":
				budget = 15
			case "code":
				v.mem[0x48a10+0x2aa8] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "entry":
				v.s.IP++
			case "segment":
				v.s.Seg[cpu.CS]++
			case "stack-low":
				v.s.R[cpu.SP] = 1
			case "stack-wrap":
				v.s.R[cpu.SP] = 65525
			case "stack-video":
				v.s.Seg[cpu.SS] = 0xa000
			case "stack-system":
				v.s.Seg[cpu.SS] = 0
				v.s.R[cpu.SP] = 0x100
			case "stack-driver":
				v.s.Seg[cpu.SS] = 0x48a1
				v.s.R[cpu.SP] = 0x1200
			case "index":
				putRowWord(v, 0x78004, 256)
			case "red":
				putRowWord(v, 0x78006, 64)
			case "green":
				putRowWord(v, 0x78008, 64)
			case "blue":
				putRowWord(v, 0x7800a, 64)
			}
			before := append([]byte(nil), v.mem...)
			if planPaletteOutput(v, budget) != nil {
				t.Fatal("unsupported RGB job accepted")
			}
			if !bytes.Equal(before, v.mem) {
				t.Fatal("refusal changed RAM")
			}
		})
	}
}
