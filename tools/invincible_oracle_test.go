package pod

// 受控原版函式／真實RAM抽樣；不是正常玩家路線。
import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/wicanr2/dosgolem/internal/cpu"
	"github.com/wicanr2/dosgolem/internal/machine"
	"os"
	"path/filepath"
	"testing"
)

type cheatOracleResult struct {
	Target            string
	On                bool
	Damage            uint16
	HP, Health        uint8
	Restored, Removed int
	Steps             uint64
}
type cheatOracleMemory struct{ *machine.Machine }

func (m cheatOracleMemory) Poke8(a uint32, v uint8) { m.Write8(a, v) }

func TestInvincibleOriginalDamage(t *testing.T) {
	dir := os.Getenv("POD_CHEAT_CASE")
	if dir == "" {
		t.Skip("需要既有正常戰場RAM及唯讀原版")
	}
	ram, e := os.ReadFile(filepath.Join(dir, "battle.ram"))
	if e != nil {
		t.Fatal(e)
	}
	codeRAM, e := os.ReadFile(filepath.Join(dir, "town.ram"))
	if e != nil {
		t.Fatal(e)
	}
	// 戰場快照的overlay已換出。保留街上已載入的原版程式區塊，
	// 明示複製戰場角色資料作受控fixture，不能當作正常玩家路線。
	copy(codeRAM[0x70000:0x90000], ram[0x70000:0x90000])
	copy(codeRAM[partyHead:partyHead+4], ram[partyHead:partyHead+4])
	codeRAM[uint32(defaultDS)<<4+0x880c] = 5
	ram = codeRAM
	ovr, e := os.ReadFile("/orig/GAME.OVR")
	if e != nil {
		t.Fatal(e)
	}
	const start = 0x45851
	if !bytes.Equal(ram[start:start+100], ovr[0x366ef:0x366ef+100]) {
		t.Fatal("傷害函式來源失配")
	}
	var results []cheatOracleResult
	for _, k := range []struct {
		on     bool
		damage uint16
		enemy  bool
	}{{false, 10, false}, {true, 10, false}, {false, 162, false}, {true, 162, false}, {false, 999, false}, {true, 999, false}, {true, 999, true}} {
		m := machine.New()
		copy(m.Mem, ram)
		m.IRQ0Every = 0
		m.VGAFrameEvery = 0
		target := uint32(0x77eb8)
		name := "player"
		want := uint8(162)
		if k.enemy {
			target = 0x79bc8
			want = 59
			name = "enemy"
		}
		c := &Invincible{On: k.on}
		lo, hi := c.GuardRange(cheatOracleMemory{m})
		restoring := false
		if lo <= hi {
			m.WatchWrite(lo, hi, func(m *machine.Machine, a uint32, old, v byte) {
				if restoring || old == v {
					return
				}
				if c.GuardWrite(a, old, v) {
					restoring = true
					m.Write8(a, old)
					restoring = false
				}
			})
		}
		m.CPU.Seg[cpu.CS] = 0x4300
		m.CPU.IP = 0x2851
		m.CPU.Seg[cpu.DS] = defaultDS
		m.CPU.Seg[cpu.SS] = 0x9000
		m.CPU.R[cpu.SP] = 0xfe0
		m.CPU.Flags = 2
		for i, v := range []uint16{0xeeee, 0x9000, k.damage, uint16(target & 15), uint16(target >> 4)} {
			binary.LittleEndian.PutUint16(m.Mem[0x90fe0+i*2:], v)
		}
		for m.Steps < 20000 && !(m.CPU.Seg[cpu.CS] == 0x9000 && m.CPU.IP == 0xeeee) {
			if e = m.Step(); e != nil {
				t.Fatal(e)
			}
		}
		if m.CPU.Seg[cpu.CS] != 0x9000 || m.CPU.IP != 0xeeee {
			t.Fatalf("原版函式沒有返回 %04x:%04x", m.CPU.Seg[cpu.CS], m.CPU.IP)
		}
		hp, status := m.Peek8(target+recCombatHP), m.Peek8(target+recStatus)
		if k.on && !k.enemy {
			if hp != want || status != 0 || c.Restored == 0 {
				t.Fatalf("保護失敗 hp%d status%d restored%d", hp, status, c.Restored)
			}
		} else {
			expected := uint8(0)
			if k.damage < uint16(want) {
				expected = want - uint8(k.damage)
			}
			if hp != expected || c.Restored != 0 {
				t.Fatalf("原版／敵方錯誤 hp%d want%d restored%d", hp, expected, c.Restored)
			}
		}
		results = append(results, cheatOracleResult{name, k.on, k.damage, hp, status, c.Restored, c.Removed, m.Steps})
	}
	data, _ := json.MarshalIndent(results, "", "  ")
	if e = os.WriteFile(filepath.Join(dir, "oracle.json"), append(data, '\n'), 0644); e != nil {
		t.Fatal(e)
	}
}
