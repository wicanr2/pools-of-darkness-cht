package pod

// 規格024：只讀片頭字幕區的已驗證色號圖樣，台詞取自目前語系表。
import (
	"crypto/sha256"
	"fmt"
)

type introCue struct {
	key        string
	background uint8
}

var introSources = map[string]introCue{
	"a3fcd07235724640d430c73838507b31a9d63fcb9adeed935872fce43ff2786f": {"intro.001", 32},
	"2c56a35a0aab3a9cd886167bfd028cabebf267cb736b3cb4aa0867d76feeac47": {"intro.002", 32},
	"513dabb86ca006cc28f6b5295769620e5b726e52d48fd60fd29266a6a495c073": {"intro.003", 32},
	"560a5b065c8e8e6bc23c2dee218cd998a21f5cf6539d224a1a8030f137568796": {"intro.004", 176},
	"12c51da1e763a7c154c9d1fda9e81850bc7eabdec8cbd1c92d48275f0d63014d": {"intro.005", 32},
	"03581843fa3b3f9b2f97040d301eaeb848d69096ac80f3f7386cb099b5231c1e": {"intro.006", 32},
	"760787555c1643f5ccd990367d6fc69f2d642a10a3aac75075cc5fba709da591": {"intro.007", 32},
	"2523926e02c3c538342953d63bbebe149a2494173f09441e4b80d8af146e6162": {"intro.008", 32},
}

func (l *Live) drawIntro(dst []byte, scale int, missing func(rune)) bool {
	if l.lang == nil || l.Font == nil || len(l.indexed) != 320*200 || (scale != 2 && scale != 3) || len(dst) != 320*200*scale*scale*4 {
		return false
	}
	cue, ok := introSources[fmt.Sprintf("%x", sha256.Sum256(l.indexed[168*320:]))]
	if !ok {
		return false
	}
	fg := l.pal[7]
	if l.pal[cue.background] != [3]uint8{} || fg[0] != fg[1] || fg[0] != fg[2] || fg[0] > 170 {
		return false
	}
	body := l.lang.Host[cue.key]
	if body == "" {
		return false
	}
	lines := wrapCols(body, 38)
	if len(lines) == 0 || len(lines) > 4 {
		return false
	}
	font, _, _, _ := l.Font.ForScale(scale)
	// Validate the complete cue before clearing anything; no clipped/missing text.
	for _, line := range lines {
		for _, r := range line {
			if r == ' ' {
				continue
			}
			if _, ok := font.Glyphs[r]; !ok {
				if missing != nil {
					missing(r)
				}
				return false
			}
		}
	}
	for y := 168 * scale; y < 200*scale; y++ {
		for x := 0; x < 320*scale; x++ {
			putRGB(dst, 320*scale, x, y, [3]uint8{})
		}
	}
	top := 168 + (32-8*len(lines))/2
	for i, line := range lines {
		drawLine(dst, scale, l.Font, (320-8*len(line))/2, top+8*i, 38, line, fg, missing)
	}
	return true
}
