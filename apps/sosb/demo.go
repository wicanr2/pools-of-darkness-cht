package sosb

import "fmt"

// 開場 demo 控制（主專案規格 018，比照幽靈騎士規格 015）：PLAY／DEMO 選單不逾時；demo 中按鍵回選單。
// 只寫下列位址；經 session.WriteGuard 擋下選單逾時的非零寫入，在回合開頭寫入中止旗標，
// 並在中止期間把雙方存活數守在 0。
// 選單逾時必須在寫入當下擋下：幽靈騎士實測選單可能在設定逾時的同一回合內就判定逾時（該作規格 015 §2.1）。
// 位址與幽靈騎士逐項同構（主專案規格 018 §2）：
//   - 主程式 START.EXE 0x3B9 設 [8A63]=3000、0x5AF 設 1000，[8A67]='D' 為逾時預設選項；選單後清 0。
//   - 選單回傳字元存 [59A2]，等於 'D' 時 0x458 設 demo 旗標 [67E9]=1。
//   - ECL 迴圈（GAME.OVR 0x47FD）每圈檢查 [59B4]、[67F5]，進入與結束都把 [59B4] 清 0。
//   - 戰鬥雙方存活數 [87CC]、[87CD]：GAME.OVR 0x3AC6D 起歸零重算，0xC445 起與 demo 旗標 [67E9] 一起判斷戰鬥結束。
const (
	demoMenuTimeout = uint32(defaultDS)<<4 + 0x8A63 // 8A63..8A66：選單逾時（百分之一秒），0 為不逾時
	demoECLStop     = uint32(defaultDS)<<4 + 0x59B4 // ECL 迴圈停止旗標
	demoSideCounts  = uint32(defaultDS)<<4 + 0x87CC // 87CC..87CD：戰鬥雙方存活數
	demoFlag        = uint32(defaultDS)<<4 + 0x67E9 // 1 ＝ demo（只讀）
	bdaKeyTail      = 0x41C                         // BIOS 鍵盤緩衝尾；bdaKeyHead 在 live.go
)

// demoMainSignature 是主程式 `cmp [59A2],'D' / jne / mov [67E9],1`（START.EXE 檔案 0x451）。
// demo 與正式遊戲都由簽章 +0x116（檔案 0x567）的 `lcall 0050:0025` 進入，返回到 +0x11B；
// 之後的 `lcall 00A1:002A` 是收尾，與幽靈騎士 `lcall 002F:0025`／`0065:002A` 的配置相同。
var demoMainSignature = sig(0x80, 0x3E, 0xA2, 0x59, 0x44, 0x75, 0x05, 0xC6, 0x06, 0xE9, 0x67, 0x01)

const (
	demoEnterOffset  = 0x116
	demoReturnOffset = 0x11B
)

// GuardMemory 是 session.GuardMemory 的同形介面，讓本套件不依賴 session。
type GuardMemory interface {
	Peek8(addr uint32) uint8
	Poke8(addr uint32, v uint8)
}

// DemoControl 實作主專案規格 018 §3.2。
type DemoControl struct {
	enter, ret uint32
	inGame     bool // 已過進入點、未到返回點
	demo       bool // 進入時 67E9＝1
	request    bool // 前端要求中止，下一回合開頭生效
	aborting   bool // 已寫入停止旗標，等返回點
	Err        error

	// 收據用計數：擋下選單逾時寫入、開始中止、中止後回到返回點的次數。
	TimeoutsCleared, Aborts, AbortReturns int
	// LastAbortSteps 是最近一次從中止要求生效到返回點的指令數。
	LastAbortSteps uint64
	abortAt        uint64
	steps          uint64
}

// NewDemoControl 以主程式簽章定位進入點與返回點；不唯一或進入點不是遠呼叫就回錯誤。
func NewDemoControl(mem CPUView) (*DemoControl, error) {
	hits := demoMainSignature.Find(mem.Read8, 0, 0xA0000)
	if len(hits) != 1 {
		return nil, fmt.Errorf("sosb: demo 主程式簽章命中 %d 處，要恰好 1 處", len(hits))
	}
	d := &DemoControl{enter: hits[0] + demoEnterOffset, ret: hits[0] + demoReturnOffset}
	// 返回點是遠呼叫的下一道指令（本作為 `cmp [741C],0`，不像幽靈騎士剛好接另一個遠呼叫），只檢查進入點
	if mem.Read8(d.enter) != 0x9A {
		return nil, fmt.Errorf("sosb: demo 進入點 %05X 不是遠呼叫", d.enter)
	}
	return d, nil
}

// Before 在每道指令前呼叫，追蹤是否在遊戲本體（demo 或正式遊戲）裡。
func (d *DemoControl) Before(step uint64, at uint32, mem CPUView) {
	d.steps = step
	switch at {
	case d.enter:
		d.inGame, d.demo = true, mem.Read8(demoFlag) == 1
	case d.ret:
		if d.aborting {
			d.AbortReturns++
			d.LastAbortSteps = step - d.abortAt
		}
		d.inGame, d.demo, d.request, d.aborting = false, false, false, false
	}
}

// Running 回 demo 是否正在進行（標題與選單期間為 false）。
func (d *DemoControl) Running() bool { return d != nil && d.inGame && d.demo }

// RequestAbort 由前端在 demo 中收到按鍵時呼叫。
func (d *DemoControl) RequestAbort() {
	if d.Running() {
		d.request = true
	}
}

var demoWritable = [][2]uint32{
	{demoMenuTimeout, demoMenuTimeout + 3},
	{demoECLStop, demoECLStop},
	{demoSideCounts, demoSideCounts + 1},
	{bdaKeyTail, bdaKeyTail},
}

func (d *DemoControl) poke(mem GuardMemory, addr uint32, v uint8) {
	for _, r := range demoWritable {
		if addr >= r[0] && addr <= r[1] {
			mem.Poke8(addr, v)
			return
		}
	}
	d.Err = fmt.Errorf("sosb: demo 控制拒絕寫入未宣告位址 %05X", addr)
}

// GuardRange 是 session.WriteGuard 的回合開頭：清掉綁定前留下的選單逾時，處理中止要求，
// 回監看範圍（平時只有選單逾時；中止期間擴到雙方存活數，兩者都在資料段 87CC..8A66 內）。
func (d *DemoControl) GuardRange(mem GuardMemory) (lo, hi uint32) {
	for a := demoMenuTimeout; a <= demoMenuTimeout+3; a++ {
		if mem.Peek8(a) != 0 {
			d.poke(mem, a, 0)
		}
	}
	if d.request && d.Running() {
		d.request, d.aborting = false, true
		d.Aborts++
		d.abortAt = d.steps
		d.poke(mem, demoSideCounts, 0)
		d.poke(mem, demoSideCounts+1, 0)
		d.poke(mem, bdaKeyTail, mem.Peek8(bdaKeyHead))
	}
	if d.aborting {
		// 每回合重寫：ECL 迴圈進入時（GAME.OVR 0x480F）會把停止旗標清 0，
		// demo 若有下一段 ECL 也要一併停下。
		d.poke(mem, demoECLStop, 1)
		return demoSideCounts, demoMenuTimeout + 3
	}
	return demoMenuTimeout, demoMenuTimeout + 3
}

// GuardWrite 擋下選單逾時的非零寫入；中止期間另把雙方存活數守在 0，讓進行中的戰鬥經正常收尾結束。
func (d *DemoControl) GuardWrite(addr uint32, old, new uint8) bool {
	if new == 0 {
		return false
	}
	if addr >= demoMenuTimeout && addr <= demoMenuTimeout+3 {
		if addr == demoMenuTimeout {
			d.TimeoutsCleared++
		}
		return true
	}
	return d.aborting && addr >= demoSideCounts && addr <= demoSideCounts+1
}
