package pod

import (
	"fmt"
	"github.com/wicanr2/dosgolem/runtime/tpcrt"
)

// 本作v1.10，主專案規格018。位址來自本作正常程式鏈，前作值不再沿用。
const (
	demoMenuTimeout  = uint32(defaultDS)<<4 + 0xAAE8
	demoECLStop      = uint32(defaultDS)<<4 + 0x7DE6
	demoGameStop     = uint32(defaultDS)<<4 + 0x8814
	demoSideCounts   = uint32(defaultDS)<<4 + 0xA7CE
	demoFlag         = uint32(defaultDS)<<4 + 0x880E
	bdaKeyTail       = 0x41C
	demoEnterOffset  = 0xC6
	demoReturnOffset = 0xCB
)

// 16AF:01B0：索引1為demo；主迴圈在0276進入、027B返回。
var demoMainSignature = sig(0x80, 0x3E, 0xDE, 0x7D, 0x01, 0x75, 0x05, 0xC6, 0x06, 0x0E, 0x88, 0x01)
var demoSoundSignature = sig(0x0e, 0xe8, 0xce, 0x00, 0x08, 0xc0, 0x74, 0x02, 0xeb, 0xf6, 0xbf, 0x6a, 0xfb, 0x1e, 0x57)

type GuardMemory interface {
	Peek8(uint32) uint8
	Poke8(uint32, uint8)
}

type demoCodePatch struct {
	at            uint32
	before, after []byte
}

// DemoControl 只改選單與中斷中的demo，正式遊戲不改等待或角色資料。
type DemoControl struct {
	enter, ret                            uint32
	inGame, demo, request, aborting       bool
	patches                               []demoCodePatch
	patched, restorePending               bool
	Err                                   error
	TimeoutsCleared, Aborts, AbortReturns int
	LastAbortSteps                        uint64
	abortAt, steps                        uint64
}

func demoBytesMatch(read func(uint32) uint8, at uint32, want []byte) bool {
	for i, b := range want {
		if read(at+uint32(i)) != b {
			return false
		}
	}
	return true
}

func NewDemoControl(mem CPUView) (*DemoControl, error) {
	hits := demoMainSignature.Find(mem.Read8, 0, 0xA0000)
	if len(hits) != 1 {
		return nil, fmt.Errorf("pod: demo 主程式簽章命中%d處，要恰好1處", len(hits))
	}
	d := &DemoControl{enter: hits[0] + demoEnterOffset, ret: hits[0] + demoReturnOffset}
	if !demoBytesMatch(mem.Read8, d.enter, []byte{0x9a, 0x25, 0, 0xef, 0x16}) ||
		!demoBytesMatch(mem.Read8, d.ret, []byte{0x9a, 0x2a, 0, 0x28, 0x17}) {
		return nil, fmt.Errorf("pod: demo 主迴圈進出點不符本作")
	}
	loop, _, ok := tpcrt.FindDelay(mem.Read8, defaultDS)
	if !ok || !demoBytesMatch(mem.Read8, loop, []byte{0x26, 0x3a, 5, 0x75, 2, 0xe2, 0xf9, 0xc3}) {
		return nil, fmt.Errorf("pod: demo CRT等待簽章不唯一或內圈不符")
	}
	sound := demoSoundSignature.Find(mem.Read8, 0, 0xA0000)
	if len(sound) != 1 || !demoBytesMatch(mem.Read8, sound[0]+0xa3, []byte{0x89, 0xec, 0x5d, 0xca, 2, 0}) {
		return nil, fmt.Errorf("pod: demo音效呼叫端或返回點不符")
	}
	d.patches = []demoCodePatch{
		{loop, []byte{0x26}, []byte{0xc3}},
		{sound[0] + 6, []byte{0x74}, []byte{0xeb}},
		{sound[0] + 10, []byte{0xbf, 0x6a, 0xfb}, []byte{0xe9, 0x96, 0}},
	}
	return d, nil
}

func (d *DemoControl) Before(step uint64, at uint32, mem CPUView) {
	d.steps = step
	switch at {
	case d.enter:
		d.inGame, d.demo = true, mem.Read8(demoFlag) == 1
	case d.ret:
		if d.aborting {
			d.AbortReturns++
			d.LastAbortSteps = step - d.abortAt
			d.restorePending = d.patched
		}
		d.inGame, d.demo, d.request, d.aborting = false, false, false, false
	}
}

func (d *DemoControl) Running() bool { return d != nil && d.inGame && d.demo }
func (d *DemoControl) RequestAbort() {
	if d.Running() && !d.aborting && d.Err == nil {
		d.request = true
	}
}

var demoWritable = [][2]uint32{
	{demoMenuTimeout, demoMenuTimeout + 3}, {demoECLStop, demoECLStop},
	{demoGameStop, demoGameStop}, {demoSideCounts, demoSideCounts + 1}, {bdaKeyTail, bdaKeyTail + 1},
}

func (d *DemoControl) fail(format string, args ...any) {
	if d.Err == nil {
		d.Err = fmt.Errorf(format, args...)
	}
}

func (d *DemoControl) poke(mem GuardMemory, addr uint32, v uint8) {
	for _, r := range demoWritable {
		if addr >= r[0] && addr <= r[1] {
			mem.Poke8(addr, v)
			return
		}
	}
	for _, p := range d.patches {
		if addr >= p.at && addr < p.at+uint32(len(p.before)) {
			i := addr - p.at
			if v == p.before[i] || v == p.after[i] {
				mem.Poke8(addr, v)
				return
			}
		}
	}
	d.fail("pod: demo控制拒絕寫入未宣告位址%05X或位元組%02X", addr, v)
}

func (d *DemoControl) patchWaits(mem GuardMemory) {
	if d.patched || d.Err != nil {
		return
	}
	// 先全部驗證，避免失配時只改了一半。
	for _, p := range d.patches {
		if !demoBytesMatch(mem.Peek8, p.at, p.before) {
			d.fail("pod: demo等待原始位元組失配%05X", p.at)
			return
		}
	}
	for _, p := range d.patches {
		for i, b := range p.after {
			d.poke(mem, p.at+uint32(i), b)
		}
	}
	d.patched = true
}

func (d *DemoControl) restoreWaits(mem GuardMemory) {
	for _, p := range d.patches {
		for i, b := range p.before {
			at := p.at + uint32(i)
			switch mem.Peek8(at) {
			case p.after[i]:
				d.poke(mem, at, b)
			case b: // 已是原始位元組。
			default:
				d.fail("pod: demo等待還原位元組失配%05X", at)
			}
		}
	}
	d.patched, d.restorePending = false, false
}

// GuardRange 在每次Advance前執行。中斷返回後先還原等待，再讓下一批玩家輸入執行。
func (d *DemoControl) GuardRange(mem GuardMemory) (uint32, uint32) {
	if d.restorePending {
		d.restoreWaits(mem)
	}
	if !d.inGame {
		for a := demoMenuTimeout; a <= demoMenuTimeout+3; a++ {
			if mem.Peek8(a) != 0 {
				d.poke(mem, a, 0)
			}
		}
	}
	if d.request && d.Running() {
		d.request, d.aborting = false, true
		d.Aborts++
		d.abortAt = d.steps
		d.poke(mem, demoSideCounts, 0)
		d.poke(mem, demoSideCounts+1, 0)
		d.poke(mem, bdaKeyTail, mem.Peek8(bdaKeyHead))
		d.poke(mem, bdaKeyTail+1, mem.Peek8(bdaKeyHead+1))
	}
	if d.aborting {
		d.poke(mem, demoECLStop, 1)
		d.poke(mem, demoGameStop, 1)
		d.patchWaits(mem)
		return demoECLStop, demoSideCounts + 1
	}
	if d.inGame {
		return 1, 0
	}
	return demoMenuTimeout, demoMenuTimeout + 3
}

func (d *DemoControl) GuardWrite(addr uint32, old, new uint8) bool {
	if new == 0 {
		return false
	}
	if !d.inGame && addr >= demoMenuTimeout && addr <= demoMenuTimeout+3 {
		if addr == demoMenuTimeout {
			d.TimeoutsCleared++
		}
		return true
	}
	return d.aborting && addr >= demoSideCounts && addr <= demoSideCounts+1
}
