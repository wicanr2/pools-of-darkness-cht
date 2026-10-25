package pod

import (
	"bytes"
	"encoding/json"
	"image/png"
	"os"
	"path/filepath"
)

// Only central game assets enter the sparse compositor. Poster alpha belongs
// to the outer window and must not disable central deferred rendering.
func deferredThemeSafe(dir string) bool {
	names := []string{"theme.json", "actors.json", "characters.json", "effects.json", "combat.json", "encounters.json", "walls.json", "exact.json"}
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return false
		}
		var root any
		if json.Unmarshal(raw, &root) != nil {
			return false
		}
		seen := map[string]bool{}
		var visit func(any) bool
		visit = func(v any) bool {
			switch v := v.(type) {
			case map[string]any:
				for k, x := range v {
					if k == "png" {
						s, ok := x.(string)
						if !ok {
							return false
						}
						if seen[s] {
							continue
						}
						seen[s] = true
						b, e := os.ReadFile(filepath.Join(dir, s))
						if e != nil {
							return false
						}
						img, e := png.Decode(bytes.NewReader(b))
						if e != nil {
							return false
						}
						r := img.Bounds()
						for y := r.Min.Y; y < r.Max.Y; y++ {
							for x := r.Min.X; x < r.Max.X; x++ {
								_, _, _, a := img.At(x, y).RGBA()
								if a != 65535 {
									return false
								}
							}
						}
					} else if !visit(x) {
						return false
					}
				}
			case []any:
				for _, x := range v {
					if !visit(x) {
						return false
					}
				}
			}
			return true
		}
		if !visit(root) {
			return false
		}
	}
	// loadLocalizedHD already rejects every non-opaque localized pixel.
	return true
}
