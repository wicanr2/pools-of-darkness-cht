package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/wicanr2/dosgolem/apps/pod"
)

func (g *game) frameSize() (int, int) {
	w := 320 * g.scale
	if g.posters != nil {
		w += 160 * g.scale
	}
	return w, 200 * g.scale
}

func (g *game) canvasPoint(x, y int) (int, int) {
	if g.posters != nil {
		x -= 80 * g.scale
	}
	return dosPoint(x, y, g.scale)
}

func (g *game) posterLocale() string {
	if !g.chinese {
		return "en"
	}
	if g.lang >= 0 && g.lang < len(g.langs) {
		return g.langs[g.lang].Code
	}
	return pod.DefaultLang
}

func (g *game) windowRGBA(game []byte) []byte {
	if g.posters == nil {
		return game
	}
	g.windowPixels = g.posters.Compose(g.windowPixels, game, g.scale, g.posterLocale())
	return g.windowPixels
}

func (g *game) saveWindowPNG(path string, game []byte) error {
	rgba := g.windowRGBA(game)
	if rgba == nil {
		return fmt.Errorf("尚無完整視窗畫格")
	}
	w, h := g.frameSize()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = png.Encode(f, &image.NRGBA{Pix: rgba, Stride: w * 4, Rect: image.Rect(0, 0, w, h)})
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	return err
}

func (g *game) windowShot(name string, rgba []byte) {
	if g.frameShotDir == "" {
		return
	}
	if err := os.MkdirAll(g.frameShotDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	copyGame := append([]byte(nil), rgba...)
	g.drawBar(copyGame)
	if err := g.saveWindowPNG(filepath.Join(g.frameShotDir, name), copyGame); err != nil {
		fmt.Fprintln(os.Stderr, "完整視窗截圖：", err)
	}
}
