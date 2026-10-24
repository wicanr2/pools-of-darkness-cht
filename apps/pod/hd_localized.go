package pod

// 主專案規格010開場語系圖層：沿用theme.json，僅改呈現，不把文字寫回原版。
import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/dosgolem/xlate"
)

type localizedSource struct{ indexed, palette string }

var localizedSources = map[string]localizedSource{
	"title-001": {"f61a14b25fe8e4921b9da854e15f1cc60475ec75ce2147e15b7d1dfc5525db0c", "17dc586aa5241390980763e8b5abf9db044d3ad788a2b2bdd5d6d285325b2acf"},
	"title-002": {"f5aff74e221772af84d2883bb091c0facf82e3995c8096ba5814c81caceb1795", "2f4985c33a587ff45d028fbdd1ba2fa34f3f5c0b9cd8c67a870b07e7604b741a"},
	"title-003": {"2c2f516bb9bd70302ee85b953673eb82f94c3ded401ac1e5968ae65e1e207396", "17dc586aa5241390980763e8b5abf9db044d3ad788a2b2bdd5d6d285325b2acf"},
}

type localizedCell struct {
	palette   string
	pixels    []byte
	ambiguous bool
}
type localizedLocale struct {
	groups map[string]*xlate.Layer
	cells  map[int]map[uint64]*localizedCell
}
type localizedHD struct {
	locales       map[string]*localizedLocale
	code, palette string
	selected      *localizedLocale
	active        *xlate.Layer
	hold          *xlate.Hold
	indexed       []byte
	lastPal       [256][3]uint8
}

func loadLocalizedHD(dir string) (*localizedHD, int, error) {
	entries, err := os.ReadDir(filepath.Join(dir, "localized"))
	if os.IsNotExist(err) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	h := &localizedHD{locales: map[string]*localizedLocale{}}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		code := entry.Name()
		items, err := LoadHDTheme(filepath.Join(dir, "localized", code))
		if err != nil {
			return nil, 0, err
		}
		if len(items) != 3 {
			return nil, 0, fmt.Errorf("開場語系%s：需要三張TITLE", code)
		}
		locale := &localizedLocale{groups: map[string]*xlate.Layer{}, cells: map[int]map[uint64]*localizedCell{}}
		seen := map[string]bool{}
		for _, it := range items {
			source, ok := localizedSources[it.Key]
			if !ok || seen[it.Key] || it.X != 0 || it.Y != 0 || it.W != 320 || it.H != 200 || it.Scale != 3 || fmt.Sprintf("%x", sha256.Sum256(it.Reference)) != source.indexed {
				return nil, 0, fmt.Errorf("開場語系%s：%s來源、位置或尺寸不符", code, it.Key)
			}
			seen[it.Key] = true
			for n := 3; n < len(it.Pix); n += 4 {
				if it.Pix[n] != 255 {
					return nil, 0, fmt.Errorf("開場語系%s：PNG含透明點", code)
				}
			}
			group := locale.groups[source.palette]
			if group == nil {
				group = &xlate.Layer{W: 320, H: 200}
				locale.groups[source.palette] = group
			}
			group.Watch(it.Watcher())
			for cy := 0; cy < 25; cy++ {
				for cx := 0; cx < 40; cx++ {
					ink, sum := xlate.CellContent(it.Reference, 320, cx, cy)
					if !ink {
						continue
					}
					cell := cy*40 + cx
					if locale.cells[cell] == nil {
						locale.cells[cell] = map[uint64]*localizedCell{}
					}
					pixels := make([]byte, 0, 24*24*4)
					for row := 0; row < 24; row++ {
						o := ((cy*24+row)*960 + cx*24) * 4
						pixels = append(pixels, it.Pix[o:o+24*4]...)
					}
					previous := locale.cells[cell][sum]
					if previous == nil {
						locale.cells[cell][sum] = &localizedCell{palette: source.palette, pixels: pixels}
					} else if previous.palette != source.palette || !bytes.Equal(previous.pixels, pixels) {
						previous.ambiguous = true
					}
				}
			}
		}
		h.locales[code] = locale
	}
	if h.locales["base"] == nil {
		return nil, 0, fmt.Errorf("開場語系缺無字base主題")
	}
	h.setLocale(DefaultLang)
	return h, 3, nil
}

func (h *localizedHD) setLocale(code string) {
	if h == nil || h.code == code && h.selected != nil {
		return
	}
	h.code = code
	h.selected = h.locales[code]
	if h.selected == nil {
		h.selected = h.locales["base"]
	}
	h.hold = xlate.NewHold(320, 200, hdHoldFrames)
	h.frame(h.indexed, h.lastPal)
}

func (h *localizedHD) frame(indexed []byte, pal [256][3]uint8) {
	if h == nil {
		return
	}
	if h.hold != nil {
		h.hold.Commit()
	}
	palette := paletteBankHash(pal, 0, 256)
	if palette != h.palette {
		h.hold = xlate.NewHold(320, 200, hdHoldFrames)
	}
	h.palette = palette
	h.lastPal = pal
	h.indexed = append(h.indexed[:0], indexed...)
	h.active = h.selected.groups[palette]
	if h.active == nil || len(indexed) != 64000 {
		return
	}
	rgb := make([]byte, 64000*3)
	for i, c := range indexed {
		copy(rgb[i*3:i*3+3], pal[c][:])
	}
	h.active.Frame(indexed, rgb)
}

func (h *localizedHD) region(cx, cy int, sum uint64) ([]byte, bool) {
	if h == nil || h.active == nil {
		return nil, false
	}
	cell := h.selected.cells[cy*40+cx][sum]
	if cell == nil {
		return nil, false
	}
	if cell.ambiguous || cell.palette != h.palette {
		return nil, true
	}
	return cell.pixels, true
}

func (h *localizedHD) art(out []byte, scale int) {
	if h != nil && scale == 3 && h.active != nil {
		h.active.Draw(out, scale, nil)
	}
}

func (h *localizedHD) draw(indexed, base, out []byte, scale int) {
	if h == nil || scale != 3 || len(indexed) != 64000 || len(out) != 960*600*4 {
		return
	}
	if h.hold != nil {
		h.hold.Apply(indexed, base, out, scale, h.region)
	}
}
