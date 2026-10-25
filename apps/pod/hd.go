package pod

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"github.com/wicanr2/dosgolem/xlate"
)

// HD 主題（death-knights-of-krynn-cht 規格 010 §2.2；xlate spec 204 圖面）。
// 主題目錄由主專案 tools/hd_theme.py 產生：theme.json 加每項的 HD PNG 與原版色號基準 .idx。
// 每項登記一個圖面 watcher：Want 是整張原版色號，畫面上整塊吻合才啟用；啟用後逐 8×8 格
// 與基準比對，被改動的格子退回原版。HD 圖只有 theme.json 的倍率一種，其他倍率不畫（spec 204 §2.3）。

type hdTheme struct {
	Scale int `json:"scale"`
	Items []struct {
		Key        string `json:"key"`
		ScreenRect [4]int `json:"screen_rect"`
		PNG        string `json:"png"`
		PNGSHA256  string `json:"png_sha256"`
		Indexed    string `json:"indexed"`
		IdxSHA256  string `json:"indexed_sha256"`
	} `json:"items"`
}

// HDItem 是載入後的一項：畫面矩形（原版像素）、HD RGBA（非預乘，已放大）與原版色號基準。
type HDItem struct {
	Key        string
	X, Y, W, H int
	Scale      int
	Pix        []uint8
	Reference  []uint8
}

const hdCell = 8

// LoadHDTheme 讀主題目錄並核對雜湊、尺寸與 8 像素格對齊；任一項不合就整個回錯。
func LoadHDTheme(dir string) ([]HDItem, error) {
	b, err := os.ReadFile(filepath.Join(dir, "theme.json"))
	if err != nil {
		return nil, err
	}
	var t hdTheme
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, fmt.Errorf("theme.json：%w", err)
	}
	if t.Scale <= 0 {
		return nil, fmt.Errorf("theme.json：倍率 %d", t.Scale)
	}
	var out []HDItem
	for _, it := range t.Items {
		x, y, w, h := it.ScreenRect[0], it.ScreenRect[1], it.ScreenRect[2], it.ScreenRect[3]
		if it.Key == "" || x < 0 || y < 0 || w <= 0 || h <= 0 || x+w > 320 || y+h > 200 ||
			x%hdCell != 0 || y%hdCell != 0 || w%hdCell != 0 || h%hdCell != 0 {
			return nil, fmt.Errorf("%s：矩形 %v 不在畫面內或沒對齊 %d 像素格", it.Key, it.ScreenRect, hdCell)
		}
		ref, err := readChecked(filepath.Join(dir, it.Indexed), it.IdxSHA256)
		if err != nil {
			return nil, fmt.Errorf("%s：%w", it.Key, err)
		}
		if len(ref) != w*h {
			return nil, fmt.Errorf("%s：基準 %d 點，應為 %d", it.Key, len(ref), w*h)
		}
		raw, err := readChecked(filepath.Join(dir, it.PNG), it.PNGSHA256)
		if err != nil {
			return nil, fmt.Errorf("%s：%w", it.Key, err)
		}
		img, err := png.Decode(bytes.NewReader(raw))
		if err != nil {
			return nil, fmt.Errorf("%s：%w", it.Key, err)
		}
		if img.Bounds().Dx() != w*t.Scale || img.Bounds().Dy() != h*t.Scale {
			return nil, fmt.Errorf("%s：HD 圖 %v，應為 %d×%d", it.Key, img.Bounds().Size(), w*t.Scale, h*t.Scale)
		}
		nrgba := image.NewNRGBA(image.Rect(0, 0, w*t.Scale, h*t.Scale))
		draw.Draw(nrgba, nrgba.Bounds(), img, img.Bounds().Min, draw.Src)
		out = append(out, HDItem{Key: it.Key, X: x, Y: y, W: w, H: h, Scale: t.Scale, Pix: nrgba.Pix, Reference: ref})
	}
	return out, nil
}

func readChecked(path, want string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(b)
	if got := hex.EncodeToString(sum[:]); got != want {
		return nil, fmt.Errorf("%s 雜湊 %s，theme.json 記 %s", filepath.Base(path), got, want)
	}
	return b, nil
}

// Rows 把一項切成 8 像素高的橫排圖面 Stamp（spec 204 §2.2），Key 為 hd:<key>:<列>。
func (it HDItem) Rows() []*xlate.Stamp {
	s := it.Scale
	rows := make([]*xlate.Stamp, 0, it.H/hdCell)
	for r := 0; r < it.H/hdCell; r++ {
		rowPix := it.W * s * 4 * hdCell * s
		rows = append(rows, &xlate.Stamp{
			Key: fmt.Sprintf("hd:%s:%d", it.Key, r), X: it.X, Y: it.Y + r*hdCell,
			Cells: it.W / hdCell, CellW: hdCell, CellH: hdCell,
			Art: true, PixScale: s,
			Pix:       it.Pix[r*rowPix : (r+1)*rowPix],
			Reference: it.Reference[r*hdCell*it.W : (r+1)*hdCell*it.W],
		})
	}
	return rows
}

// Watcher 回這一項的圖面 watcher：Want 為整張原版色號，Owner 為 hd:<key>。
func (it HDItem) Watcher() *xlate.Watcher {
	rows := it.Rows()
	keys := make([]string, len(rows))
	for i, s := range rows {
		keys[i] = s.Key
	}
	// Sticky：啟用後只遮被原版改掉的格，不因另一張圖開始畫就整張撤下（xlate spec 204 §6；主專案規格 021 §4.1）
	return &xlate.Watcher{Key: "hd:" + it.Key, X: it.X, Y: it.Y, W: it.W, H: it.H, Want: it.Reference,
		Art: true, ArtKeys: keys, Sticky: true, Make: func() []*xlate.Stamp { return it.Rows() }}
}

// AddHD 載入主題目錄並登記全部圖面 watcher。
func (l *Live) AddHD(dir string) error {
	items, err := LoadHDTheme(dir)
	if err != nil {
		return err
	}
	layers, count, err := LoadHDLayers(dir)
	if err != nil {
		return err
	}
	localized, localizedCount, err := loadLocalizedHD(dir)
	if err != nil {
		return err
	}
	bigpics, plain, err := loadBigpicHD(items)
	if err != nil {
		return err
	}
	l.deferredUnsafe = !deferredThemeSafe(dir)
	for _, it := range plain {
		l.Layer.Watch(it.Watcher())
	}
	l.HDLayers = layers
	l.localized = localized
	l.bigpics = bigpics
	if l.lang != nil {
		l.localized.setLocale(l.lang.Code)
	}
	if layers.combat != nil && len(layers.combat.vgaGeometry) > 0 {
		l.battleHold = &battleHold{}
	}
	l.HDItems += len(items) + count + localizedCount
	l.armHDHold(plain)
	if count > 0 {
		l.hdLatch = xlate.NewLatch(320, 200, hdLatchKeep, hdLatchFrames)
	}
	return nil
}

// hdLatchKeep、hdLatchFrames：HD 鎖存每格最多記幾種內容、閒置多少幀（retrace）後移除（xlate spec 202 §2.9）。
const hdLatchKeep, hdLatchFrames = 4, 600

// hdHoldFrames 是 HD 保留呈現的連續上限（幀；主專案規格 021 §4.1）。實測標題背景從開始畫到整張吻合超過 10 幀
// （冷開機第 487–498 格），8 幀不夠；保留只作用在內容等於主題圖對應格的格子，原版動畫不受影響。
const hdHoldFrames = 60

// hdCell 是某個畫面格、某種原版內容對應的 HD 畫面；同一位置同樣內容對到不同 HD 畫面時 pix 為 nil（有歧義）。
type hdHoldCell struct{ pix []uint8 }

// armHDHold 建立 HD 保留呈現（xlate spec 202 §2.8）：記下每張主題圖各墨跡格的色號指紋與 HD 畫面。
// 畫面某格內容等於某張圖在同一位置的格，就是原版正在畫這張圖：換圖前直接畫那一格的 HD；
// 同一位置同樣內容對到多張不同 HD 時，退回保留上一幀的呈現。
func (l *Live) armHDHold(items []HDItem) {
	cells := map[int]map[uint64]*hdHoldCell{}
	for _, it := range items {
		if it.Scale != hdHoldScale {
			continue
		}
		s, W := it.Scale, it.W*it.Scale
		for cy := 0; cy < it.H/hdCell; cy++ {
			for cx := 0; cx < it.W/hdCell; cx++ {
				ink, sum := xlate.CellContent(it.Reference, it.W, cx, cy)
				if !ink {
					continue
				}
				pix := make([]uint8, 0, hdCell*s*hdCell*s*4)
				for y := cy * hdCell * s; y < (cy+1)*hdCell*s; y++ {
					o := (y*W + cx*hdCell*s) * 4
					pix = append(pix, it.Pix[o:o+hdCell*s*4]...)
				}
				c := (it.Y/hdCell+cy)*(320/hdCell) + it.X/hdCell + cx
				if cells[c] == nil {
					cells[c] = map[uint64]*hdHoldCell{}
				}
				switch old := cells[c][sum]; {
				case old == nil:
					cells[c][sum] = &hdHoldCell{pix: pix}
				case old.pix != nil && !bytes.Equal(old.pix, pix):
					old.pix = nil
				}
			}
		}
	}
	if len(cells) == 0 {
		return
	}
	l.hdCells = cells
	l.hdHold = xlate.NewHold(320, 200, hdHoldFrames)
}

// hdHoldScale 是 HD 圖的倍率：保留呈現只在這個倍率作用。
const hdHoldScale = 3

func (l *Live) hdRegion(cx, cy int, sum uint64) ([]uint8, bool) {
	c := l.hdCells[cy*(320/hdCell)+cx][sum]
	if c == nil {
		return nil, false
	}
	return c.pix, true
}
