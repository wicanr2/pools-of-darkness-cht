package pod

// 主專案規格010：本次正常街景已證實的第1組牆面及天空，其他VGA圖組不推定。
import (
	"fmt"
	"strconv"
	"strings"
)

func knownVGAWall(s wallAssetSpec) (bool, error) {
	if s.Key == "skygrnd-001" && s.Kind == "sky" && s.Size == [2]int{88, 88} && s.Algorithm == "" {
		return true, nil
	}
	if strings.HasPrefix(s.Key, "8x8d1-001@") {
		n, err := strconv.Atoi(strings.TrimPrefix(s.Key, "8x8d1-001@"))
		known := map[int]bool{21: true, 32: true, 36: true, 37: true, 38: true, 40: true, 41: true, 42: true, 43: true, 44: true, 46: true, 47: true, 55: true, 60: true, 61: true, 62: true, 193: true}
		if err == nil && s.Key == fmt.Sprintf("8x8d1-001@%d", n) && known[n] && s.Kind == "wall" && s.Size == [2]int{8, 8} && s.Algorithm == "" {
			return true, nil
		}
	}
	if strings.HasPrefix(s.Key, "8x8d1-") || strings.HasPrefix(s.Key, "skygrnd-") {
		return false, fmt.Errorf("%s：未證實的VGA牆面／天空來源", s.Key)
	}
	return false, nil
}

func vgaWallColor(c byte, wall bool) bool {
	return c <= 15 || wall && c >= 32 && c <= 143 || !wall && c >= 144 && c <= 175
}

func (h *wallHD) vgaPaletteOK(pal [256][3]uint8) bool {
	if paletteBankHash(pal, 32, 144) != "23c42b2c120d0b7740f503cc8d2c20ab29ddaa2ad0408eba17805a79c641ee2f" ||
		paletteBankHash(pal, 144, 176) != "f4780f94d70c97e8bd659ff668a8f8eab9599ba9e66233d4aa878d6a8fa1b609" {
		return false
	}
	// 有效來源色盤到達後才有可核對的RGB；壞接縫整組回退，不拿成品反推色盤。
	for _, a := range h.vgaAssets {
		if !a.wall {
			continue
		}
		for y := 0; y < 24; y++ {
			for x := 0; x < 24; x++ {
				if x != 0 && y != 0 && x != 23 && y != 23 {
					continue
				}
				c := a.ref[y/3*8+x/3]
				want := pal[c]
				if c == 13 {
					want = [3]uint8{85, 85, 85}
				}
				n := (y*24 + x) * 4
				for k := 0; k < 3; k++ {
					if a.pix[n+k] != want[k] {
						return false
					}
				}
			}
		}
	}
	return true
}
