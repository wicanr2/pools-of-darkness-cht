package drawblocks

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"os"
	"path/filepath"
	"testing"
)

func surfaceSetupSeed(t *testing.T, ip uint16) *view {
	t.Helper()
	root := os.Getenv("POD_ROOT")
	if root == "" {
		t.Skip("original surface setup fixtures absent")
	}
	folder := filepath.Join(root, "workplace/probe/draw-intercept-v1/native-fixed-control-v1/trace-audit")
	raw, e := os.ReadFile(filepath.Join(folder, "fixed-records.json"))
	if e != nil {
		t.Fatal(e)
	}
	var records []struct {
		Before struct {
			R                    [8]uint16
			Seg                  [4]uint16
			IP, Flags            uint16
			Memory, MemorySHA256 string
		}
	}
	if e = json.Unmarshal(raw, &records); e != nil {
		t.Fatal(e)
	}
	for _, r := range records {
		b := r.Before
		if b.IP != ip {
			continue
		}
		mem, e := os.ReadFile(filepath.Join(folder, b.Memory))
		if e != nil {
			t.Fatal(e)
		}
		if fmt.Sprintf("%x", sha256.Sum256(mem)) != b.MemorySHA256 {
			t.Fatal("fixture hash")
		}
		return &view{mem: mem, s: machine.HostRegisters{R: b.R, Seg: b.Seg, IP: b.IP, Flags: b.Flags}}
	}
	t.Fatal("missing entry")
	return nil
}
func surfaceSetupCompare(t *testing.T, v *view, p *machine.HostWorkPlan) {
	t.Helper()
	if p == nil || !p.Algorithmic {
		t.Fatal("valid setup refused")
	}
	compare(t, v, p)
	a, b := callMachine(v), callMachine(v)
	runCall(t, a, v, false)
	if e := b.CommitHostBlock(p); e != nil {
		t.Fatal(e)
	}
	paletteMachinesEqual(t, a, b)
}
func TestSurfaceRootSetupOriginal(t *testing.T) {
	cases := 0
	for _, pointer := range []uint32{0x500, 0x169f0, 0x6000f, 0x9ffef} {
		for _, length := range []uint32{0, 1, 15, 16, 17, 65535, 65536, 0x100001, 0xffffffff} {
			for _, dos := range []bool{false, true} {
				v := surfaceSetupSeed(t, 0x2786)
				v.dos = dos
				v.s.Flags |= cpu.DF
				stack := uint32(v.SS())*16 + uint32(v.SP())
				putRowWord(v, stack+4, uint16(pointer&15))
				putRowWord(v, stack+6, uint16(pointer>>4))
				putRowWord(v, stack+8, uint16(length))
				putRowWord(v, stack+10, uint16(length>>16))
				for slot := uint32(2); slot < 256; slot++ {
					putRowWord(v, 0x48a10+0xcbd+slot*2, uint16(slot<<8)|3)
				}
				p := planSurfaceSetup(v, 837)
				surfaceSetupCompare(t, v, p)
				if planSurfaceSetup(v, 836) != nil {
					t.Fatal("short root budget accepted")
				}
				cases++
			}
		}
	}
	t.Logf("surface-root ordered original and whole SaveState+PortLog=%d unused flag high bytes preserved", cases)
}
func TestSurfaceChildSetupOriginal(t *testing.T) {
	cases := 0
	for _, free := range []uint16{2, 3, 17, 255} {
		for _, links := range []uint16{0, 1, 3, 8} {
			if links >= free {
				continue
			}
			for _, point := range []bool{false, true} {
				for _, dos := range []bool{false, true} {
					v := surfaceSetupSeed(t, 0x28d5)
					v.dos = dos
					v.s.Flags |= cpu.DF
					stack := uint32(v.SS())*16 + uint32(v.SP())
					for slot := uint16(0); slot < 256; slot++ {
						flags := uint16(0)
						if slot >= free {
							flags = 0x80
						}
						putRowWord(v, 0x48a10+0xcbd+uint32(slot)*2, flags)
					}
					putRowWord(v, 0x48a10+0xbd, 0xa000)
					for slot := uint16(0); slot <= links; slot++ {
						for off, value := range map[uint16]uint16{0x4bd: 3, 0x6bd: 4, 0x8bd: 300, 0xabd: 180} {
							putRowWord(v, 0x48a10+uint32(off)+uint32(slot)*2, value)
						}
						if slot != 0 {
							putRowWord(v, 0x48a10+0xcbd+uint32(slot)*2, 0x40)
							putRowWord(v, 0x48a10+0xbd+uint32(slot)*2, (slot-1)*2)
						}
					}
					right, bottom := uint16(280), uint16(170)
					if point {
						right, bottom = 7, 8
					}
					for off, value := range map[uint32]uint16{4: links, 6: 7, 8: 8, 10: right, 12: bottom} {
						putRowWord(v, stack+off, value)
					}
					p := planSurfaceSetup(v, 4096)
					surfaceSetupCompare(t, v, p)
					if planSurfaceSetup(v, p.Instructions-1) != nil {
						t.Fatal("short child budget accepted")
					}
					cases++
				}
			}
		}
	}
	t.Logf("surface-child ordered original and whole SaveState+PortLog=%d free slots through255 parents through8", cases)
}
func TestSurfaceSetupRefusals(t *testing.T) {
	cases := 0
	for _, ip := range []uint16{0x2786, 0x28d5} {
		for _, kind := range []string{"budget", "code", "mode", "CS", "IP", "stack-low", "stack-wrap", "stack-code", "unknown-model", "8086", "no-model", "pointer", "slot", "free", "cycle", "rectangle", "parent"} {
			if ip == 0x2786 && kind != "pointer" && (kind == "slot" || kind == "free" || kind == "cycle" || kind == "rectangle" || kind == "parent") {
				continue
			}
			if ip == 0x28d5 && kind == "pointer" {
				continue
			}
			v := surfaceSetupSeed(t, ip)
			stack := uint32(v.SS())*16 + uint32(v.SP())
			budget := uint64(4096)
			switch kind {
			case "budget":
				budget = 0
			case "code":
				v.mem[0x48a10+uint32(ip)] ^= 1
			case "mode":
				v.mem[0x449] = 0xd
			case "CS":
				v.s.Seg[cpu.CS]++
			case "IP":
				v.s.IP++
			case "stack-low":
				v.s.Seg[cpu.SS] = 0
				v.s.R[cpu.SP] = 2
			case "stack-wrap":
				v.s.R[cpu.SP] = 65530
			case "stack-code":
				v.s.Seg[cpu.SS] = 0x48a1
				v.s.R[cpu.SP] = 0x1000
			case "pointer":
				putRowWord(v, stack+6, 0xa000)
			case "slot":
				putRowWord(v, stack+4, 256)
			case "free":
				for i := uint32(0); i < 256; i++ {
					putRowWord(v, 0x48a10+0xcbd+i*2, 0)
				}
			case "cycle":
				putRowWord(v, 0x48a10+0xcbd, 0x40)
				putRowWord(v, 0x48a10+0xbd, 0)
			case "rectangle":
				putRowWord(v, stack+6, 320)
			case "parent":
				putRowWord(v, 0x48a10+0xbd, 0x6000)
			}
			var w CPUView = v
			switch kind {
			case "unknown-model":
				w = primitiveUnknownModel{v}
			case "8086":
				w = primitive8086Model{v}
			case "no-model":
				w = struct{ CPUView }{v}
			}
			before := append([]byte(nil), v.mem...)
			regs := v.s
			if planSurfaceSetup(w, budget) != nil {
				t.Fatal("unverified setup accepted", ip, kind)
			}
			if !bytes.Equal(before, v.mem) || regs != v.s {
				t.Fatal("refusal mutated input")
			}
			cases++
		}
	}
	t.Logf("surface setup unchanged refusals=%d", cases)
}
