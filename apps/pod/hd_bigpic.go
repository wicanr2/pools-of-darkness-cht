package pod

// 主專案規格010 b25：BIGPIC沿既有theme.json，另以實際使用RGB守衛。
import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/wicanr2/dosgolem/xlate"
)

var bigpicSources = map[string]localizedSource{
	"bigpic1-240": {"a9b0b5ec46b8a7843bc396a956d2305f41ffe32fdfb154e51b0e3027fe7b3e25", "d4b874f4b104dd3bbfc52af42f44712f6c10cc7690be135703990a9b81f5bb3f"},
	"bigpic1-241": {"3b94cba579ee81bf01289e66eadac799b784870ebb210cbffd4f6ed50ba1b0aa", "119d6584a3deccfa9aa523ccffd44b563a10ff1e1301c37bc7de635673e0b605"},
	"bigpic1-242": {"98a9650b8bcc7bef8822439837588ad88a0c7441cea58884d8d6ef6b440257bc", "99b2b6ec37a1a19bd97619b0f6e281bf893ea92fdace30c44cca43dc93388133"},
	"bigpic1-243": {"6e01f186c173e4c069866e768329e8aa29d2e5802403548ad82753e345294ba0", "c1a37316f25866530b2293734ae58fa521fcc98a925a72b1c4a0bb5d16d0fb49"},
	"bigpic1-244": {"04d166867708c4d61b9f05ff719ddfd8bf42aaa748f46f03077d725da262d33c", "e885f147b960484e95cddaa3af91a6948617cb094daee98109f6ca5724951bf7"},
	"bigpic1-245": {"70272d633fe952e63194c75663da87619dfe8c95967d5aa58e097f69557bd7f5", "a559474554368ffc2bd2a172e334cadd62df4876097049b06a67b2a441f62b00"},
	"bigpic1-246": {"2f400e237a5431a4b4d55331eec575e15ae73bf659c911de44a64ea6d226b683", "12d7331ae09ba966df9fe31cf8361c489815c78c61a564a00b523432b7c8bba1"},
}

type bigpicPattern struct {
	layer   *xlate.Layer
	used    []byte
	palette string
}
type bigpicCell struct {
	item   int
	pixels []byte
}
type bigpicHD struct {
	patterns []bigpicPattern
	cells    map[int]map[uint64][]bigpicCell
	active   uint8
	hold     *xlate.Hold
}

func loadBigpicHD(items []HDItem) (*bigpicHD, []HDItem, error) {
	h := &bigpicHD{cells: map[int]map[uint64][]bigpicCell{}}
	plain := make([]HDItem, 0, len(items))
	seen := map[string]bool{}
	for _, it := range items {
		if !strings.HasPrefix(it.Key, "bigpic1-") {
			plain = append(plain, it)
			continue
		}
		source, ok := bigpicSources[it.Key]
		if !ok || seen[it.Key] || it.X != 8 || it.Y != 8 || it.W != 304 || it.H != 120 || it.Scale != 3 || fmt.Sprintf("%x", sha256.Sum256(it.Reference)) != source.indexed {
			return nil, nil, fmt.Errorf("%s：BIGPIC來源、位置或尺寸不符", it.Key)
		}
		seen[it.Key] = true
		for n := 3; n < len(it.Pix); n += 4 {
			if it.Pix[n] != 255 {
				return nil, nil, fmt.Errorf("%s：BIGPIC含透明點", it.Key)
			}
		}
		var used [256]bool
		for _, v := range it.Reference {
			used[v] = true
		}
		p := bigpicPattern{layer: &xlate.Layer{W: 320, H: 200}, palette: source.palette}
		for n, yes := range used {
			if yes {
				p.used = append(p.used, byte(n))
			}
		}
		p.layer.Watch(it.Watcher())
		number := len(h.patterns)
		h.patterns = append(h.patterns, p)
		for cy := 0; cy < it.H/8; cy++ {
			for cx := 0; cx < it.W/8; cx++ {
				ink, sum := xlate.CellContent(it.Reference, it.W, cx, cy)
				if !ink {
					continue
				}
				cell := (it.Y/8+cy)*40 + it.X/8 + cx
				if h.cells[cell] == nil {
					h.cells[cell] = map[uint64][]bigpicCell{}
				}
				pixels := make([]byte, 0, 24*24*4)
				for row := 0; row < 24; row++ {
					off := ((cy*24+row)*it.W*3 + cx*24) * 4
					pixels = append(pixels, it.Pix[off:off+24*4]...)
				}
				h.cells[cell][sum] = append(h.cells[cell][sum], bigpicCell{number, pixels})
			}
		}
	}
	if len(h.patterns) == 0 {
		return nil, plain, nil
	}
	h.reset()
	return h, plain, nil
}

func (h *bigpicHD) reset() {
	if h != nil {
		h.active = 0
		h.hold = xlate.NewHold(320, 200, hdHoldFrames)
	}
}
func (h *bigpicHD) frame(indexed []byte, pal [256][3]uint8) {
	if h == nil {
		return
	}
	h.hold.Commit()
	var active uint8
	if len(indexed) == 64000 {
		for i, p := range h.patterns {
			signature := make([]byte, 0, len(p.used)*4)
			for _, v := range p.used {
				signature = append(signature, v)
				signature = append(signature, pal[v][:]...)
			}
			if fmt.Sprintf("%x", sha256.Sum256(signature)) == p.palette {
				active |= 1 << i
			}
		}
	}
	if active != h.active {
		h.hold = xlate.NewHold(320, 200, hdHoldFrames)
	}
	h.active = active
	if active == 0 {
		return
	}
	rgb := make([]byte, 64000*3)
	for i, v := range indexed {
		copy(rgb[i*3:i*3+3], pal[v][:])
	}
	for i, p := range h.patterns {
		if active&(1<<i) != 0 {
			p.layer.Frame(indexed, rgb)
		}
	}
}
func (h *bigpicHD) region(cx, cy int, sum uint64) ([]byte, bool) {
	if h == nil || h.active == 0 {
		return nil, false
	}
	var pixels []byte
	for _, cell := range h.cells[cy*40+cx][sum] {
		if h.active&(1<<cell.item) == 0 {
			continue
		}
		if pixels == nil {
			pixels = cell.pixels
		} else if !bytes.Equal(pixels, cell.pixels) {
			return nil, true
		}
	}
	return pixels, pixels != nil
}
func (h *bigpicHD) art(out []byte, scale int) {
	if h == nil || scale != 3 {
		return
	}
	for i, p := range h.patterns {
		if h.active&(1<<i) != 0 {
			p.layer.Draw(out, scale, nil)
		}
	}
}
func (h *bigpicHD) draw(indexed, base, out []byte, scale int) {
	if h != nil && scale == 3 && len(indexed) == 64000 && len(out) == 960*600*4 {
		h.hold.Apply(indexed, base, out, scale, h.region)
	}
}
