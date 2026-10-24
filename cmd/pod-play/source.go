package main

import (
	"errors"
	"os"
)

// saveSource只讀取最後呈現的原版畫格；不讀目前可能已往前執行的VGA記憶體。
func (g *game) saveSource(prefix string) error {
	if len(g.live.Indexed()) != 64000 {
		return errors.New("尚無完整原版色號畫格")
	}
	if err := os.WriteFile(prefix+".idx", g.live.Indexed(), 0600); err != nil {
		return err
	}
	pal := g.live.Palette()
	flat := make([]byte, 0, 768)
	for _, rgb := range pal {
		flat = append(flat, rgb[:]...)
	}
	return os.WriteFile(prefix+".pal", flat, 0600)
}
