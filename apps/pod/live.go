package pod

import (
	"fmt"
	"strings"

	"github.com/wicanr2/dosgolem/runtime/tpcrt"
	"github.com/wicanr2/dosgolem/xlate"
)

// StepCPU 是每道指令前攔截點需要的機器狀態（session.StepView 與收據工具都滿足）。
type StepCPU interface {
	CPUView
	Steps() uint64
	CS() uint16
	IP() uint16
	SS() uint16
	SP() uint16
	DS() uint16
}

// Live 把路徑 A、B、C 組在同一個 xlate Layer 上（規格 001–003）。
//
// 攔截點的位址以簽章搜尋，但要等遊戲本體就位才搜：冷開機的第 0 步，記憶體裡的
// 主程式還沒自我展開（本作 GAME.EXE 是 EXEPACK，且由 CONTROL.EXE 以 EXEC 叫起），
// 遮罩簽章會對上一段不是程序本身的位元組（前作 2026-10-15 前端實測：
// 綁在 #0 的位址被執行時參數全不合法）。因此綁定在第一次視訊模式改變之後
// （ModeChange），或由呼叫端在已進入遊戲的狀態檔上呼叫 Arm。失敗就每 bindRetry 步再試；
// 搜尋結果不唯一就不綁定、不猜位址；最後一次失敗原因留在 BindErr。
type Live struct {
	*HDLayers        // 規格010 §6，僅讀原版畫面。
	prevB     []byte // 路徑 B 上一則字串（手札引用分兩則印）
	Layer     *xlate.Layer
	// HDItems 是 AddHD 載入的 HD 主題項數（主專案規格 010）；0 表示沒有 HD 素材（規格 011 切換列停用 HD）。
	HDItems   int
	Font      *xlate.Font
	glyphText *glyphText
	A         *PathA
	B         *PathB
	C         *PathC

	// Answers 是查驗題答案表（規格 006）；綁定前設定。
	Answers     map[QuizKey]string
	QuizFormat  QuizFormat
	hideAnswers bool
	// Journal 是手札面板（規格 007）；nil 時不啟用。
	Journal *JournalPanel
	// Guide 是攻略面板（規格 008）；nil 時不啟用。
	Guide *GuidePanel
	// Map 是地圖疊層（規格 009）；nil 時不啟用。
	Map *AutoMap
	// Music 是配樂選曲（規格 014）；nil 時不啟用。
	Music *MusicDirector
	ds    uint16 // 遊戲資料段（路徑 A 進入時取）
	// Inventory 為 true 時，綁定後建立 Inv 記錄印字路徑盤點（issue #11）。
	Inventory bool
	Inv       *Inventory
	// ControlDemo 為 true 時，綁定後建立 Demo（主專案規格 018）；定位失敗只記在 DemoErr，不影響覆繪。
	ControlDemo bool
	Demo        *DemoControl
	DemoErr     error
	// TPDelayLoop、TPDelayCount 是綁定時以簽章找到的 CRT.Delay（dosgolem 規格 429 §2.4；本作是 TP5 寫法，
	// 主專案規格 019）；找不到時為 0。
	TPDelayLoop, TPDelayCount uint32
	nextMapSample             uint64
	// Flash 是閃現量尺（主專案規格 021）；nil 時不量。由 AuditFlash 每畫格比對一次。
	Flash      *xlate.FlashAudit
	flashFrame int
	// hdHold、hdCells：HD 保留呈現與主題圖各墨跡格的指紋（主專案規格 021 §4.1；AddHD 時建立）
	hdHold         *xlate.Hold
	hdCells        map[int]map[uint64]*hdHoldCell
	hdLatch        *xlate.Latch // 戰鬥等圖層的 HD 鎖存（AddHD 載入圖層時建立）
	battleHold     *battleHold
	localized      *localizedHD
	bigpics        *bigpicHD
	Direct         *DirectCanvas
	deferredUnsafe bool

	ui      Lookuper
	ecl     *Catalog
	ready   bool
	armed   bool
	BindErr error  // 最近一次綁定失敗的原因；綁定成功後為 nil
	BoundAt uint64 // 綁定成功的步數
	indexed []byte
	pal     [256][3]uint8
	lang    *LangPack // 目前語言（主專案規格 015；SetLang 設定，nil 表示 NewLive 給的表）
	// journalKeep、guideKeep 保存面板物件：語言沒有該面板時 Journal／Guide 設為 nil（停用），換回時還原
	journalKeep *JournalPanel
	guideKeep   *GuidePanel
}

// NewLive 建立尚未綁定位址的覆繪；ui 是介面字串查表（可為規格 004 的 Translator），
// ecl 可為 nil（不啟用路徑 B）。
func NewLive(font *xlate.Font, ui Lookuper, ecl *Catalog) *Live {
	AddHalfWidthPairs(font)
	if font == nil || font.Scaled[3].Font == nil { // 有外部 3 倍字模（<名稱>-3x.golemfnt）時不從 16×16 放大
		AddScaledVariants(font)
	}
	return &Live{HDLayers: &HDLayers{}, Layer: &xlate.Layer{W: 320, H: 200}, Font: font, ui: ui, ecl: ecl, glyphText: newGlyphText(font)}
}

const bindRetry = 1_000_000

// defaultDS 是本作 GAME.EXE v1.10 的資料段：照 START.BAT 的程式鏈（常駐驅動＋-root-tsr-shrink）時，
// 主選單 DS=2096（本專案 docs/re/003、004，dosgolem 觀測）。程式鏈不同時落點不同；路徑 A 進入時以當下 DS 更新。
const defaultDS = 0x2096

// mapSampleEvery 是地圖疊層讀記憶體的間隔（規格 009 §2.2）。
const mapSampleEvery = 1_000_000

// Arm 允許下一步起綁定攔截點（從已進入遊戲的狀態檔開始時使用）。
func (l *Live) Arm() { l.armed = true }

func (l *Live) bind(mem CPUView) error {
	var err error
	l.A, l.B, l.C = nil, nil, nil
	l.Layer.Frozen = nil
	if l.A, err = NewPathA(mem, l.ui, l.Font, l.Layer); err != nil {
		return err
	}
	l.A.Answers, l.A.QuizFormat, l.A.HideAnswers = l.Answers, l.QuizFormat, l.hideAnswers
	l.A.ASCIIWide = l.lang != nil && l.lang.Code == EnglishLang
	if l.C, err = NewPathC(mem, l.ui, l.Font, l.Layer); err != nil {
		return err
	}
	if l.ecl != nil {
		if l.B, err = NewPathB(mem, l.Font, l.Layer, l.ecl, l.ui); err != nil {
			return err
		}
		l.B.OnEntry = l.journalEntry
		if l.lang != nil {
			l.B.Ref = l.lang.Ref // 綁定前已 SetLang（主專案規格 015）
		}
	}
	if l.Music != nil {
		l.A.OnText = func(t []byte) { l.Music.Text(string(t)) }
		l.C.OnLine = func(_ int, line string) { l.Music.Menu(line) }
	}
	l.ready = true
	l.ds = defaultDS
	if l.Inventory {
		l.Inv = newInventory(l)
	}
	if l.ControlDemo {
		l.Demo, l.DemoErr = NewDemoControl(mem)
	}
	if loop, count, ok := tpcrt.FindDelay(mem.Read8, defaultDS); ok {
		l.TPDelayLoop, l.TPDelayCount = loop, count
	}
	return nil
}

// BeforeStep 在每道指令前呼叫。
func (l *Live) BeforeStep(v StepCPU) error {
	l.HDLayers.observeIntro(v)
	if !l.ready {
		if !l.armed || l.BindErr != nil && v.Steps()%bindRetry != 0 {
			return nil
		}
		if err := l.bind(v); err != nil {
			l.BindErr = fmt.Errorf("pod: 綁定攔截點（#%d）：%w", v.Steps(), err)
			return nil
		}
		l.BindErr, l.BoundAt = nil, v.Steps()
	}
	l.glyphText.before(v)
	if l.Journal.Open() {
		l.Journal.KeyTaken(v.Read16(bdaKeyHead)) // 規格 007 §2.4 條件 2：先比對再處理進入點
	}
	cs, ip, ss, sp := v.CS(), v.IP(), v.SS(), v.SP()
	if l.Demo != nil {
		l.Demo.Before(v.Steps(), linear(cs, ip), v)
	}
	if l.Map != nil {
		if linear(cs, ip) == l.A.Entry {
			l.ds = v.DS()
		}
		// 以門檻取代整除：TP Delay 快轉一次跳過許多步，整除點可能被跳過（主專案規格 019）；
		// 沒有快轉時每次都在整除點跨過門檻，取樣點與舊版相同。
		if s := v.Steps(); l.ds != 0 && s >= l.nextMapSample {
			if l.nextMapSample != 0 || s%mapSampleEvery == 0 {
				l.Map.Sample(v, l.ds)
				l.Music.MapKey(l.Map.Key)
			}
			l.nextMapSample = s - s%mapSampleEvery + mapSampleEvery
		}
	}
	na, nb := len(l.A.Hits), 0
	if l.Inv != nil {
		l.Inv.before(v.Steps(), cs, ip, ss, sp, v)
	}
	l.A.Before(v.Steps(), cs, ip, ss, sp, v)
	l.C.Before(cs, ip, ss, sp, v)
	if l.B != nil {
		nb = len(l.B.Hits)
		l.B.Before(v.Steps(), cs, ip, ss, sp, v.DS(), v)
	}
	if l.Inv != nil {
		if len(l.A.Hits) > na {
			l.Inv.attach("A", l.A.Hits[len(l.A.Hits)-1])
		}
		if l.B != nil && len(l.B.Hits) > nb {
			l.Inv.attach("B", l.B.Hits[len(l.B.Hits)-1])
		}
	}
	return nil
}

// VideoWrite 處理 A000 寫入（offset 相對 A000）。
func (l *Live) VideoWrite(offset uint32) {
	l.glyphText.write(offset)
	if !l.ready {
		return
	}
	l.C.VideoWrite(offset)
	if l.B != nil {
		l.B.VideoWrite(offset)
	}
}

// SetHideAnswers 切換查驗題答案顯示（規格 006 §2.2）。
func (l *Live) SetHideAnswers(hide bool) {
	l.hideAnswers = hide
	l.Layer.ForgetGhosts() // 同樣的原文改成另一種呈現（xlate spec 202 §2.7）
	if l.A != nil {
		l.A.SetHideAnswers(hide)
	}
}

// HideAnswers 回目前是否隱藏查驗題答案。
func (l *Live) HideAnswers() bool { return l.hideAnswers }

// bdaKeyHead 是 BIOS 鍵盤緩衝區 head 指標（0040:001A）。
const bdaKeyHead = 0x41A

// journalEntry 是路徑 B 進入時的手札處理（規格 007 §2.2、§2.4 條件 1）。
func (l *Live) journalEntry(text []byte, mem CPUView) {
	l.Music.Text(string(text))
	if l.Journal == nil {
		return
	}
	// 酒店傳聞的編號後單獨印一個句點（`YOU HEAR TAVERN TALE `、`40`、`.`；本專案 docs/re/008）：純標點不關面板
	if strings.Trim(string(text), " .,!?") == "" {
		return
	}
	l.Journal.CloseAuto()
	prev := append([]byte(nil), l.prevB...) // 複製：下一行會重用 l.prevB 的底層陣列
	l.prevB = append(l.prevB[:0], text...)
	if n, _, ok := JournalRefOf(prev, text); ok {
		l.Journal.OpenAuto(n, mem.Read16(bdaKeyHead))
	}
}

// ModeChange 在視訊模式改變時呼叫。
func (l *Live) ModeChange() {
	if l.HDLayers != nil {
		l.intro.reset()
	}
	l.glyphText.reset()
	l.Direct.Reset()
	l.bigpics.reset()
	if l.localized != nil {
		l.localized.hold = xlate.NewHold(320, 200, hdHoldFrames)
	}
	if l.battleHold != nil {
		l.battleHold = &battleHold{}
	}
	l.Journal.closeAutoSafe()
	l.Layer.ForgetGhosts()
	if l.hdLatch != nil {
		l.hdLatch.Forget()
	}
	if !l.ready {
		l.armed = true
		return
	}
	l.C.Reset()
	if l.B != nil {
		l.B.Reset()
	}
}

// Frame 在每個 retrace 呼叫；palette 須是與 indexed 配對的色盤（machine.FramePalette）。
func (l *Live) Frame(indexed []byte, palette [256][3]uint8) {
	if l.Direct != nil {
		if err := l.Direct.Frame(indexed); err != nil {
			panic(err)
		}
	}
	l.bigpics.frame(indexed, palette)
	l.localized.frame(indexed, palette)
	l.battleHold.commit()
	if l.hdHold != nil {
		l.hdHold.Commit() // 上一幀最後一次合成的結果成為保留的依據
	}
	if l.hdLatch != nil {
		l.hdLatch.Tick()
	}
	l.HDLayers.Frame(indexed, palette)
	l.glyphText.frame(indexed)
	l.indexed = append(l.indexed[:0], indexed...)
	l.pal = palette
	if len(indexed) != 320*200 {
		return
	}
	// 綁定前也跑 Layer.Frame：HD 圖面（hd.go）的標題畫面出現在綁定之前，那時還沒有疊字。
	if l.ready {
		l.A.Pal = palette
		l.A.PrepareFrame(indexed, 320, 200)
		l.C.SetPalette(palette)
		if l.B != nil {
			l.B.SetPalette(palette)
		}
	} else if l.Layer.Watchers() == 0 {
		return
	}
	rgb := make([]uint8, len(indexed)*3)
	for i, c := range indexed {
		rgb[i*3], rgb[i*3+1], rgb[i*3+2] = palette[c][0], palette[c][1], palette[c][2]
	}
	l.Layer.Frame(indexed, rgb)
}

// Compose 回最後一個畫格放大 scale 倍的 RGBA；overlay 為 false 時是原版畫面。
// missing 收字型缺字（可為 nil）。尚無 320×200 畫格時回 nil。
func (l *Live) Compose(scale int, overlay bool, missing func(rune)) []byte {
	if l.Direct != nil && overlay && !l.deferredUnsafe {
		return l.composeDirect(scale, missing)
	}
	if l.Direct != nil && overlay {
		l.Direct.stats.LegacyRGBFrames++
	}
	return l.composeLegacy(scale, overlay, missing)
}

func (l *Live) composeLegacy(scale int, overlay bool, missing func(rune)) []byte {
	if len(l.indexed) != 320*200 || scale <= 0 {
		return nil
	}
	w := 320 * scale
	out := make([]uint8, w*200*scale*4)
	for y := 0; y < 200*scale; y++ {
		row := l.indexed[(y/scale)*320:]
		for x := 0; x < w; x++ {
			c := l.pal[row[x/scale]]
			o := (y*w + x) * 4
			out[o], out[o+1], out[o+2], out[o+3] = c[0], c[1], c[2], 0xFF
		}
	}
	if overlay {
		l.HDLayers.Draw(out, l.indexed, scale)
		var base []byte
		if l.hdLatch != nil && scale == hdHoldScale {
			// 戰鬥等圖層整件比對，一件被蓋到一小塊就整件退回；原版沒變的格沿用上次的 HD（xlate spec 202 §2.9）
			base = l.Compose(scale, false, nil)
			l.hdLatch.Apply(l.indexed, l.palSig(), base, out, scale)
		}
		l.localized.art(out, scale)
		l.bigpics.art(out, scale)
		l.glyphText.draw(out, scale, l.pal)
		l.Layer.Draw(out, scale, missing)
		if l.localized != nil && scale == hdHoldScale {
			if base == nil {
				base = l.Compose(scale, false, nil)
			}
			l.localized.draw(l.indexed, base, out, scale)
		}
		if l.bigpics != nil && scale == hdHoldScale {
			if base == nil {
				base = l.Compose(scale, false, nil)
			}
			l.bigpics.draw(l.indexed, base, out, scale)
		}
		if l.hdHold != nil && scale == hdHoldScale {
			if base == nil {
				base = l.Compose(scale, false, nil)
			}
			l.hdHold.Apply(l.indexed, base, out, scale, l.hdRegion)
		}
		if l.battleHold != nil && scale == hdHoldScale {
			if base == nil {
				base = l.Compose(scale, false, nil)
			}
			l.battleHold.apply(l.indexed, base, out, l.palSig(), l.HDLayers.pendingBattleCells())
		}
		l.HDLayers.drawIntroArt(out, scale)
		l.drawIntro(out, scale, missing)
		if l.Flash != nil && l.flashFrame != 0 {
			if base == nil {
				base = l.Compose(scale, false, nil)
			}
			l.Flash.Observe(l.flashFrame, l.indexed, base, out, scale)
			l.flashFrame = 0
		}
		if l.Journal.Open() {
			l.Journal.Draw(out, scale, l.Font, missing)
		}
		if l.Guide.Open() {
			l.Guide.Draw(out, scale, l.Font, missing)
		}
		if l.Map.Open() {
			l.Map.Draw(out, scale, l.Font, missing)
		}
	}
	return out
}

// AuditFlash 以閃現量尺比對這一畫格的呈現（主專案規格 021；xlate spec 202 §2.6）：疊字與 HD 合成一次，面板不列入。
// l.Flash 為 nil 時不做事。
func (l *Live) AuditFlash(frame, scale int) {
	if l.Flash == nil {
		return
	}
	l.flashFrame = frame
	l.Compose(scale, true, nil)
	l.flashFrame = 0
}

// palSig 是目前全部256色的指紋（HD鎖存與戰場保留判斷色盤沒變）。
func (l *Live) palSig() uint64 {
	h := uint64(14695981039346656037)
	for _, c := range l.pal {
		for _, v := range c {
			h = (h ^ uint64(v)) * 1099511628211
		}
	}
	return h
}

// Indexed 回最後一個畫格的原版色號（收據用）。
func (l *Live) Indexed() []byte { return l.indexed }

// Palette 回與 Indexed 配對的最後呈現畫格色盤副本（原始素材比對收據）。
func (l *Live) Palette() [256][3]uint8 { return l.pal }

// AddScaledVariants 為 16×16 字型登記 3× 變體：最近鄰放成 22×22，置於 24 像素格內偏移 (1,1)，
// 比照拯救地球（apps/buckrogers manualThreeXFont、manualGlyphOffset）。
// 規格見 death-knights-of-krynn-cht docs/spec/010 §2.3；xlate spec 202 §2.4.1。須在半形字對登記之後呼叫。
func AddScaledVariants(font *xlate.Font) {
	if font == nil || font.W != 16 || font.H != 16 {
		return
	}
	font.Scaled = map[int]xlate.ScaledFont{3: {Font: xlate.DeriveNearest(font, 22, 22), X: 1, Y: 1}}
}
