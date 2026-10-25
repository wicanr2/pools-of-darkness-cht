// Paired complete CALL cost, including qualification, RAM and pixel notifications.
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/wicanr2/dosgolem/apps/pod"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type view struct{ m *machine.Machine }

func (v view) Read8(a uint32) byte    { return v.m.Peek8(a) }
func (v view) Read16(a uint32) uint16 { return uint16(v.Read8(a)) | uint16(v.Read8(a+1))<<8 }
func (v view) Steps() uint64          { return v.m.Steps }
func (v view) AX() uint16             { return v.m.CPU.R[0] }
func (v view) CX() uint16             { return v.m.CPU.R[1] }
func (v view) DX() uint16             { return v.m.CPU.R[2] }
func (v view) BX() uint16             { return v.m.CPU.R[3] }
func (v view) SP() uint16             { return v.m.CPU.R[4] }
func (v view) BP() uint16             { return v.m.CPU.R[5] }
func (v view) SI() uint16             { return v.m.CPU.R[6] }
func (v view) DI() uint16             { return v.m.CPU.R[7] }
func (v view) ES() uint16             { return v.m.CPU.Seg[0] }
func (v view) CS() uint16             { return v.m.CPU.Seg[1] }
func (v view) SS() uint16             { return v.m.CPU.Seg[2] }
func (v view) DS() uint16             { return v.m.CPU.Seg[3] }
func (v view) IP() uint16             { return v.m.CPU.IP }
func (v view) Flags() uint16          { return v.m.CPU.Flags }
func (v view) DOSBoxCost() bool       { return v.m.CPU.DOSBoxCost }

type capture struct {
	R         [8]uint16
	Seg       [4]uint16
	IP, Flags uint16
	Memory    string
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func read(p string) []byte { b, e := os.ReadFile(p); must(e); return b }
func main() {
	root := flag.String("root", "", "normal call corpus")
	out := flag.String("out", "", "new report")
	flag.Parse()
	if _, e := os.Stat(*out); !os.IsNotExist(e) {
		panic("report exists")
	}
	var input struct{ Captures []capture }
	must(json.Unmarshal(read(filepath.Join(*root, "calls.json")), &input))
	type row struct {
		Source                       string
		Instructions                 uint64
		LegacyNS, HostNS             []int64
		LegacyMedianNS, HostMedianNS int64
		Ratio                        float64
		CallsPerSample               uint64
	}
	rows := []row{}
	for _, index := range []int{0, 10, 20} {
		c := input.Captures[index]
		mem := read(filepath.Join(*root, c.Memory))
		sp := int(c.Seg[2])*16 + int(c.R[4])
		args := append([]byte(nil), mem[sp+4:sp+14]...)
		type fixture struct {
			m      *machine.Machine
			canvas *pod.DirectCanvas
			per    uint64
		}
		build := func(native bool) fixture {
			m := machine.New()
			m.SetVideoMode(0x13)
			copy(m.Mem, mem)
			m.IRQ0Every = 0
			m.VGAFrameEvery = 0
			m.CPU.Model = cpu.Model80386
			m.CPU.R = c.R
			m.CPU.Seg = c.Seg
			m.CPU.Seg[cpu.CS] = 0x9000
			m.CPU.R[cpu.SP] += 14
			m.CPU.IP = 0x100
			m.CPU.Flags = c.Flags
			m.CPU.Cycles = 0
			m.SetNativeVideo(true)
			code := []byte{}
			for i := 4; i >= 0; i-- {
				word := binary.LittleEndian.Uint16(args[i*2:])
				code = append(code, 0xb8, byte(word), byte(word>>8), 0x50)
			}
			code = append(code, 0x9a, 0xa6, 0x11, 0x2e, 0x1b)
			code = append(code, 0xe9, 0, 0)
			delta := uint16(int16(-len(code)))
			binary.LittleEndian.PutUint16(code[len(code)-2:], delta)
			copy(m.Mem[0x90100:], code)
			originalView := view{m: m}
			b := m.CPU.R
			m.CPU.R[cpu.SP] -= 14
			m.CPU.Seg[cpu.CS] = 0x1b2e
			m.CPU.IP = 0x11a6
			p := pod.PlanGlyphCall(originalView, 4096)
			if p == nil {
				panic("unsupported corpus")
			}
			per := p.Instructions + 12
			m.CPU.R = b
			m.CPU.Seg[cpu.CS] = 0x9000
			m.CPU.IP = 0x100
			canvas := pod.NewDirectCanvas()
			m.ObserveVideoWrites(func(machine.VideoWrite) {})
			m.ObserveVideoCommands(func(cmd machine.VideoCommand) { must(canvas.Apply(cmd)) })
			if native {
				m.ObserveHostCalls(func() {
					p := pod.PlanGlyphCall(view{m}, m.HostWorkBudget(m.RunLimit-m.Steps))
					if p == nil {
						panic("bench declined")
					}
					must(m.CommitHostWork(p))
				})
				m.SetHostCallTargets(pod.GlyphHostCallTargets())
			}
			return fixture{m, canvas, per}
		}
		legacy, native := build(false), build(true)
		run := func(f fixture, calls uint64) int64 {
			budget := f.per * calls
			f.m.RunLimit = f.m.Steps + budget
			start := time.Now()
			_, e := f.m.RunUntilObserved(nil, budget, nil)
			must(e)
			return time.Since(start).Nanoseconds() / int64(calls)
		}
		run(legacy, 64)
		run(native, 64)
		if legacy.m.Steps != native.m.Steps || legacy.m.CPU.R != native.m.CPU.R || legacy.m.CPU.Seg != native.m.CPU.Seg || legacy.m.CPU.IP != native.m.CPU.IP || legacy.m.CPU.Flags != native.m.CPU.Flags || legacy.m.CPU.Cycles != native.m.CPU.Cycles || !bytes.Equal(legacy.m.Mem, native.m.Mem) {
			panic("paired CALL state mismatch")
		}
		r := row{Source: c.Memory, Instructions: legacy.per - 12, CallsPerSample: 512}
		for i := 0; i < 7; i++ {
			var a, b int64
			if i%2 == 0 {
				a = run(legacy, 512)
				b = run(native, 512)
			} else {
				b = run(native, 512)
				a = run(legacy, 512)
			}
			r.LegacyNS = append(r.LegacyNS, a)
			r.HostNS = append(r.HostNS, b)
		}
		median := func(v []int64) int64 {
			c := append([]int64(nil), v...)
			sort.Slice(c, func(i, j int) bool { return c[i] < c[j] })
			return c[len(c)/2]
		}
		r.LegacyMedianNS = median(r.LegacyNS)
		r.HostMedianNS = median(r.HostNS)
		r.Ratio = float64(r.LegacyMedianNS) / float64(r.HostMedianNS)
		if native.m.HostCallCommits != 64+7*512 {
			panic("not every benchmark call used")
		}
		rows = append(rows, r)
	}
	b, e := json.MarshalIndent(struct {
		Scope string
		Rows  []row
	}{"三個真實字模來源；七組交錯、每組512次完整CALL，含客體參數準備、入口篩選、簽章、計畫、RAM及逐點／命令通知；事件關閉；不含Owner.BeforeStep、遊戲規則、RGB、GPU或視窗，不外推FPS", rows}, "", "  ")
	must(e)
	must(os.WriteFile(*out, append(b, '\n'), 0600))
	fmt.Println(string(b))
}
