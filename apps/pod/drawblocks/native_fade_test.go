package drawblocks

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
)

func fadeSeed(t *testing.T, ip, count uint16) *view {
	seed := paletteSeed(t)
	v := &view{mem: append([]byte(nil), seed.mem...), s: seed.s}
	v.s.IP = ip
	v.s.Seg[cpu.DS] = 0x48a1
	v.s.R[cpu.SI], v.s.R[cpu.CX], v.s.R[cpu.BX] = 0, count, 0x1200
	v.s.Flags &^= cpu.DF
	return v
}

func TestNativeFadeDifferenceOriginal(t *testing.T) {
	plans := 0
	pairs := [][2]byte{{0, 63}, {63, 0}, {17, 17}, {20, 21}, {21, 20}, {0, 0}}
	for _, count := range []uint16{1, 3, 17, 768} {
		for _, maximum := range []byte{0, 17, 63} {
			for _, dos := range []bool{false, true} {
				v := fadeSeed(t, 0x4b19, count)
				v.dos = dos
				v.s.R[cpu.BX] |= uint16(maximum)
				for x := 0; x < int(count); x++ {
					v.mem[0x48a10+0x1872+x], v.mem[0x48a10+0x1b72+x] = pairs[x%len(pairs)][0], pairs[x%len(pairs)][1]
				}
				for _, budget := range []uint64{10, 13, 64, 1024} {
					if p := planFadeDifference(v, budget); p != nil {
						compare(t, v, p)
						plans++
					}
				}
			}
		}
	}
	t.Log("fade difference synthetic original comparisons", plans)
}

func TestNativeFadeBlendOriginal(t *testing.T) {
	plans := 0
	errors, distances := []byte{0, 31, 30, 16, 0}, []byte{0, 1, 2, 63, 31}
	for _, count := range []uint16{1, 2, 3, 17, 768} {
		for _, dos := range []bool{false, true} {
			v := fadeSeed(t, 0x4b6c, count)
			v.dos = dos
			v.mem[0x48a10+0x1173] = 31
			for x := 0; x < int(count); x++ {
				v.mem[0x48a10+0x1e72+x], v.mem[0x48a10+0x2172+x] = errors[x%len(errors)], distances[x%len(distances)]
				v.mem[0x48a10+0x1872+x] = 32
				v.mem[0x48a10+0x2472+x] = 1
				if x%2 == 0 {
					v.mem[0x48a10+0x2472+x] = 255
				}
			}
			for _, budget := range []uint64{7, 10, 64, 1024} {
				p := planFadeBlend(v, budget)
				if p == nil {
					t.Fatal("valid fade blend rejected")
				}
				compare(t, v, p)
				plans++
			}
		}
	}
	t.Log("fade blend synthetic original comparisons", plans)
}

func TestNativeFadeNormalStates(t *testing.T) {
	root := os.Getenv("POD_FADE_CALLS")
	if root == "" {
		t.Skip("complete normal-DEMO fade entry fixtures absent")
	}
	raw, err := os.ReadFile(filepath.Join(root, "calls.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		DrawCaptures []struct {
			R                    [8]uint16
			Seg                  [4]uint16
			IP, Flags            uint16
			Memory, MemorySHA256 string
		}
	}
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	states := map[uint16][]*view{}
	for _, row := range report.DrawCaptures {
		if row.Seg[cpu.CS] != 0x48a1 || (row.IP != 0x4b19 && row.IP != 0x4b6c) {
			continue
		}
		mem, err := os.ReadFile(filepath.Join(root, row.Memory))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(mem)
		if hex.EncodeToString(sum[:]) != row.MemorySHA256 || row.R[cpu.CX] < 32 {
			t.Fatal("normal original fade fixture identity or channel count invalid")
		}
		m := original(&view{mem: mem, s: machine.HostRegisters{R: row.R, Seg: row.Seg, IP: row.IP, Flags: row.Flags}})
		for m.Steps < 1000000 && len(states[row.IP]) < 32 {
			if m.CPU.Seg[cpu.CS] == 0x48a1 && m.CPU.IP == row.IP {
				states[row.IP] = append(states[row.IP], &view{mem: append([]byte(nil), m.Mem...), s: machine.HostRegisters{R: m.CPU.R, Seg: m.CPU.Seg, IP: m.CPU.IP, Flags: m.CPU.Flags}})
			}
			if err := m.Step(); err != nil {
				t.Fatal(err)
			}
		}
	}
	plans := 0
	for _, ip := range []uint16{0x4b19, 0x4b6c} {
		if len(states[ip]) != 32 {
			t.Fatal("normal fade states absent", ip, len(states[ip]))
		}
		for _, v := range states[ip] {
			for _, dos := range []bool{false, true} {
				v.dos = dos
				for _, budget := range []uint64{13, 64, 1024} {
					var p *machine.HostWorkPlan
					if ip == 0x4b19 {
						p = planFadeDifference(v, budget)
					} else {
						p = planFadeBlend(v, budget)
					}
					if p == nil {
						t.Fatalf("normal fade rejected %04x", ip)
					}
					compare(t, v, p)
					plans++
				}
			}
		}
	}
	t.Log("fade normal original comparisons", plans)
}

func TestNativeFadeRefusals(t *testing.T) {
	for _, ip := range []uint16{0x4b19, 0x4b6c} {
		kinds := []string{"budget", "code", "mode", "direction-flag", "segment", "counter", "entry"}
		if ip == 0x4b19 {
			kinds = append(kinds, "from", "to", "maximum")
		} else {
			kinds = append(kinds, "error", "distance", "threshold", "direction", "increment", "decrement")
		}
		for _, kind := range kinds {
			t.Run(kind, func(t *testing.T) {
				v := fadeSeed(t, ip, 1)
				v.mem[0x48a10+0x1872], v.mem[0x48a10+0x1b72] = 32, 30
				v.mem[0x48a10+0x1e72], v.mem[0x48a10+0x2172] = 31, 1
				v.mem[0x48a10+0x1173], v.mem[0x48a10+0x2472] = 31, 1
				budget := uint64(1024)
				switch kind {
				case "budget":
					budget = 1
				case "code":
					v.mem[0x48a10+uint32(ip)] ^= 1
				case "mode":
					v.mem[0x449] = 0xd
				case "direction-flag":
					v.s.Flags |= cpu.DF
				case "segment":
					v.s.Seg[cpu.DS] = 0x5000
				case "counter":
					v.s.R[cpu.SI] = 768
				case "entry":
					v.s.IP++
				case "from":
					v.mem[0x48a10+0x1872] = 64
				case "to":
					v.mem[0x48a10+0x1b72] = 64
				case "maximum":
					v.s.R[cpu.BX] |= 64
				case "error":
					v.mem[0x48a10+0x1e72] = 64
				case "distance":
					v.mem[0x48a10+0x2172] = 64
				case "threshold":
					v.mem[0x48a10+0x1173] = 64
				case "direction":
					v.mem[0x48a10+0x2472] = 0
				case "increment":
					v.mem[0x48a10+0x1872] = 63
				case "decrement":
					v.mem[0x48a10+0x1872], v.mem[0x48a10+0x2472] = 0, 255
				}
				var p *machine.HostWorkPlan
				if ip == 0x4b19 {
					p = planFadeDifference(v, budget)
				} else {
					p = planFadeBlend(v, budget)
				}
				if p != nil {
					t.Fatal("unproved fade accepted")
				}
			})
		}
	}
}
