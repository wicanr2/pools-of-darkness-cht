package pod

import "fmt"

// 本作v1.10定位及原始讀寫端見主專案規格017、研究029。
const CheatVerified = true

const (
	partyHead    = uint32(defaultDS)<<4 + 0x9CF2
	recNext      = 0x1E3
	recAffects   = 0x143
	affectNext   = 5
	recStatus    = 0x1ED
	recTeam      = 0x1EF
	recHP        = 0xB2
	recCombatHP  = 0x1FC
	maxAffects   = 64
	maxPartyRecs = 12
	cheatProgram = 0x16CA0 // runtime16AF:01B0
)

// 僅列本作已定位的失明、緩慢、麻痺、睡眠、中毒；未知效果保留。
var harmfulAffect = map[uint8]bool{0x21: true, 0x2A: true, 0x34: true, 0x35: true, 0x37: true}

type Memory interface {
	Peek8(uint32) uint8
	Poke8(uint32, uint8)
}

type Invincible struct {
	On                bool
	recs              []uint32
	Report            string
	Restored, Removed int
	Rejected          int
}

func far(peek func(uint32) uint8, a uint32) uint32 {
	off := uint32(peek(a)) | uint32(peek(a+1))<<8
	seg := uint32(peek(a+2)) | uint32(peek(a+3))<<8
	return seg*16 + off
}

func validCheatRecord(peek func(uint32) uint8, a uint32) bool {
	if a < 0x10000 || a+recCombatHP >= 0xA0000 {
		return false
	}
	n := peek(a)
	if n == 0 || n > 15 {
		return false
	}
	for i := uint32(1); i <= uint32(n); i++ {
		if v := peek(a + i); v < 32 || v > 126 {
			return false
		}
	}
	return peek(a+recTeam) <= 1 && peek(a+recStatus) <= 9
}

// 9CF2為全隊串列，9CEE只是目前角色。遇敵方正常停止；壞串列整批拒絕。
func party(peek func(uint32) uint8) []uint32 {
	var out []uint32
	seen := map[uint32]bool{}
	for a := far(peek, partyHead); a != 0; a = far(peek, a+recNext) {
		if seen[a] || len(out) >= maxPartyRecs || !validCheatRecord(peek, a) {
			return nil
		}
		if peek(a+recTeam) != 0 {
			return out
		}
		for _, prior := range out {
			if a < prior+recCombatHP+1 && a+recCombatHP+1 > prior {
				return nil
			}
		}
		seen[a] = true
		out = append(out, a)
	}
	return out
}

// 先驗證所有我方串列，再改任何連結。原九bytes節點不釋放或覆寫。
func affectNodes(mem Memory, rec uint32, recs []uint32) ([]uint32, bool) {
	var nodes []uint32
	seen := map[uint32]bool{}
	for node := far(mem.Peek8, rec+recAffects); node != 0; node = far(mem.Peek8, node+affectNext) {
		if node < 0x10000 || node+9 > 0xA0000 || seen[node] || len(nodes) >= maxAffects {
			return nil, false
		}
		for _, a := range recs {
			if node < a+recCombatHP+1 && node+9 > a {
				return nil, false
			}
		}
		for _, a := range nodes {
			if node < a+9 && node+9 > a {
				return nil, false
			}
		}
		seen[node] = true
		nodes = append(nodes, node)
	}
	return nodes, true
}

func (c *Invincible) GuardRange(mem Memory) (lo, hi uint32) {
	if c == nil {
		return 1, 0
	}
	c.recs = nil
	c.Report = ""
	if !c.On || !demoBytesMatch(mem.Peek8, cheatProgram, demoMainSignature.Bytes) {
		return 1, 0
	}
	recs := party(mem.Peek8)
	if len(recs) == 0 {
		return 1, 0
	}
	chains := make([][]uint32, len(recs))
	used := map[uint32]bool{}
	for i, a := range recs {
		var ok bool
		chains[i], ok = affectNodes(mem, a, recs)
		if !ok {
			c.Rejected++
			return 1, 0
		}
		for _, node := range chains[i] {
			for j := uint32(0); j < 9; j++ {
				if used[node+j] {
					c.Rejected++
					return 1, 0
				}
				used[node+j] = true
			}
		}
	}
	c.recs = recs
	lo = ^uint32(0)
	for i, a := range recs {
		link := a + recAffects
		for _, node := range chains[i] {
			if !harmfulAffect[mem.Peek8(node)] {
				link = node + affectNext
				continue
			}
			for j := uint32(0); j < 4; j++ {
				mem.Poke8(link+j, mem.Peek8(node+affectNext+j))
			}
			c.Removed++
		}
		name := make([]byte, mem.Peek8(a))
		for j := range name {
			name[j] = mem.Peek8(a + 1 + uint32(j))
		}
		c.Report += fmt.Sprintf("%s %d/%d 狀態%d；", name, mem.Peek8(a+recHP), mem.Peek8(a+recCombatHP), mem.Peek8(a+recStatus))
		lo = min(lo, a+recHP)
		hi = max(hi, a+recCombatHP)
	}
	return lo, hi
}

func (c *Invincible) GuardWrite(addr uint32, old, new uint8) bool {
	if c == nil || !c.On {
		return false
	}
	for _, a := range c.recs {
		restore := false
		switch addr - a {
		case recHP, recCombatHP:
			restore = new < old
		case recStatus, recTeam:
			restore = old == 0 && new != 0
		default:
			continue
		}
		if restore {
			c.Restored++
		}
		return restore
	}
	return false
}
