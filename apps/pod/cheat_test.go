package pod

import (
	"reflect"
	"testing"
)

type cheatMem map[uint32]uint8

func (m cheatMem) Peek8(a uint32) uint8    { return m[a] }
func (m cheatMem) Poke8(a uint32, v uint8) { m[a] = v }

// fakeParty 在 64 KB 以上擺三筆紀錄：兩名隊員、一隻怪物，串成隊伍串列。
func fakeParty() cheatMem {
	mem := cheatMem{}
	for i, b := range demoMainSignature.Bytes {
		mem[cheatProgram+uint32(i)] = b
	}
	put := func(a uint32, seg uint16) {
		mem[a], mem[a+1], mem[a+2], mem[a+3] = 0, 0, uint8(seg), uint8(seg>>8)
	}
	put(partyHead, 0x5219)       // 隊員 1：0x52190
	put(0x52190+recNext, 0x5240) // 隊員 2：0x52400，完整紀錄不重疊
	put(0x52400+recNext, 0x57BA) // 怪物：0x57BA0
	for _, a := range []uint32{0x52190, 0x52400, 0x57BA0} {
		mem[a] = 1
		mem[a+1] = 'A'
	}
	mem[0x57BA0+recTeam] = 1
	put(0x57BA0+recNext, 0x5219) // 怪物再指回隊員 1，走訪要在怪物就停
	return mem
}

func TestInvinciblePartyWalk(t *testing.T) {
	mem := fakeParty()
	peek := func(a uint32) uint8 { return mem[a] }
	got := party(peek)
	if len(got) != 2 || got[0] != 0x52190 || got[1] != 0x52400 {
		t.Fatalf("隊伍串列走訪 %x，要 52190、52400（遇怪物停）", got)
	}
	// 迴圈：隊員 2 指回隊員 1，不可無限走
	mem[0x52400+recNext+2], mem[0x52400+recNext+3] = 0x19, 0x52
	if got := party(peek); len(got) != 0 {
		t.Fatalf("迴圈串列必須整批拒絕，得到%x", got)
	}
}

func TestInvincibleGuard(t *testing.T) {
	mem := fakeParty()
	c := &Invincible{}
	if lo, hi := c.GuardRange(mem); lo <= hi {
		t.Fatal("關閉時不監看")
	}
	if c.GuardWrite(0x52190+recHP, 50, 40) {
		t.Fatal("關閉時不還原")
	}
	c.On = true
	lo, hi := c.GuardRange(mem)
	if lo != 0x52190+recHP || hi != 0x52400+recCombatHP {
		t.Fatalf("範圍 %x–%x", lo, hi)
	}
	cases := []struct {
		addr     uint32
		old, new uint8
		restore  bool
		why      string
	}{
		{0x52190 + recHP, 50, 40, true, "扣血"},
		{0x52400 + recHP, 50, 51, false, "回血"},
		{0x52400 + recCombatHP, 162, 0, true, "戰鬥致命傷害"},
		{0x52400 + recCombatHP, 162, 163, false, "戰鬥治療"},
		{0x52190 + recStatus, 0, 4, true, "正常變異常"},
		{0x52190 + recStatus, 6, 0, false, "復活"},
		{0x52190 + recTeam, 0, 1, true, "魅惑（陣營改成敵方）"},
		{0x52190 + recTeam - 1, 0, 1, false, "其他欄位"},
		{0x57BA0 + recHP, 9, 0, false, "怪物"},
	}
	for _, k := range cases {
		if got := c.GuardWrite(k.addr, k.old, k.new); got != k.restore {
			t.Errorf("%s：還原 %v，要 %v", k.why, got, k.restore)
		}
	}
}

func TestInvincibleStripsHarmfulAffects(t *testing.T) {
	mem := fakeParty()
	// 隊員 1 的效果串列：麻痺 → 有益(0x12) → 睡眠 → 失明 → 有益(0x5F)
	nodes := []struct {
		at  uint32
		typ uint8
	}{{0x60000, 0x34}, {0x60010, 0x12}, {0x60020, 0x35}, {0x60030, 0x21}, {0x60040, 0x5F}}
	link := uint32(0x52190 + recAffects)
	for _, n := range nodes {
		mem[link], mem[link+1], mem[link+2], mem[link+3] = 0, 0, uint8((n.at>>4)&0xFF), uint8(n.at>>12)
		mem[n.at] = n.typ
		link = n.at + affectNext
	}
	c := &Invincible{On: true}
	c.GuardRange(mem)
	var got []uint8
	for a := far(mem.Peek8, 0x52190+recAffects); a != 0; a = far(mem.Peek8, a+affectNext) {
		got = append(got, mem[a])
	}
	if len(got) != 2 || got[0] != 0x12 || got[1] != 0x5F || c.Removed != 3 {
		t.Fatalf("剩下 %x、移除 %d，要剩 12、5F，移除 3", got, c.Removed)
	}
	off := &Invincible{}
	mem[0x52400+recAffects+2], mem[0x52400+recAffects+3] = 0x00, 0x60 // 隊員 2 指向麻痺
	off.GuardRange(mem)
	if far(mem.Peek8, 0x52400+recAffects) != 0x60000 {
		t.Fatal("關閉時不可改效果串列")
	}
}

func TestInvincibleRejectsBeforeAnyMutation(t *testing.T) {
	for _, kind := range []string{"program", "effect-cycle", "shared-effect", "record-overlap", "bad-party", "bad-name"} {
		t.Run(kind, func(t *testing.T) {
			m := fakeParty()
			c := &Invincible{On: true}
			put := func(at, node uint32) {
				m[at] = uint8(node & 15)
				m[at+1] = 0
				m[at+2] = uint8(node >> 4)
				m[at+3] = uint8(node >> 12)
			}
			put(0x52190+recAffects, 0x60000)
			m[0x60000] = 0x37
			switch kind {
			case "program":
				m[cheatProgram] ^= 1
			case "effect-cycle":
				put(0x52400+recAffects, 0x60100)
				put(0x60105, 0x60100)
			case "shared-effect":
				put(0x52400+recAffects, 0x60000)
			case "record-overlap":
				put(0x52400+recAffects, 0x52190)
			case "bad-party":
				put(0x52400+recNext, 0xA0000)
			case "bad-name":
				m[0x52400] = 16
			}
			before := cheatMem{}
			for k, v := range m {
				before[k] = v
			}
			lo, hi := c.GuardRange(m)
			if lo <= hi || !reflect.DeepEqual(before, m) || c.Removed != 0 || c.GuardWrite(0x52190+recHP, 40, 0) {
				t.Fatalf("%s未原子拒絕", kind)
			}
		})
	}
}

func TestInvincibleOnlyVerifiedEffects(t *testing.T) {
	m := fakeParty()
	link := uint32(0x52190 + recAffects)
	types := []uint8{0x1F, 0x21, 0x2A, 0x33, 0x34, 0x35, 0x37, 0x5F}
	for i, typ := range types {
		node := uint32(0x60000 + i*16)
		m[link], m[link+1], m[link+2], m[link+3] = 0, 0, uint8(node>>4), uint8(node>>12)
		m[node] = typ
		link = node + 5
	}
	c := &Invincible{On: true}
	c.GuardRange(m)
	var got []uint8
	for a := far(m.Peek8, 0x52190+recAffects); a != 0; a = far(m.Peek8, a+5) {
		got = append(got, m[a])
	}
	if !reflect.DeepEqual(got, []uint8{0x1F, 0x33, 0x5F}) || c.Removed != 5 {
		t.Fatalf("未知／有益效果遭改：%x，移除%d", got, c.Removed)
	}
	c.On = false
	c.GuardRange(m)
	if c.GuardWrite(0x52190+recHP, 40, 0) {
		t.Fatal("關閉後仍保護")
	}
}
