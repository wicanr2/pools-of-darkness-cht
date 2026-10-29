// 研究用契約測試。由toolbar-acceptance-receipt.sh複製到隔離cmd副本。
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/apps/pod"
	"github.com/wicanr2/dosgolem/host"
	"github.com/wicanr2/dosgolem/session"
)

func TestAcceptanceNativeToolbarAndPreferenceClicks(t *testing.T) {
	f, err := pod.LoadFontScaled("/font/pod-rounded.golemfnt")
	if err != nil {
		t.Fatal(err)
	}
	fonts, err := loadBarFonts(f, "/font/pod-rounded.golemfnt")
	if err != nil {
		t.Fatal(err)
	}
	o, err := session.New(session.Config{InitialScale: host.OutputScale2, ClockPercent: 50})
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()
	dir := t.TempDir()
	if output := os.Getenv("POD_TOOLBAR_PREF_TEST_DIR"); output != "" {
		dir = output
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	g := &game{owner: o, mainFont: f, live: &pod.Live{Font: f}, barFonts: fonts, cheat: &pod.Invincible{},
		barText: testBarText, scale: 3, chinese: true,
		speed:     speedControl{base: 50, set: 1, pct: 50},
		speedPath: filepath.Join(dir, "pod-speed.txt"), barSizePath: filepath.Join(dir, "pod-barfont.txt")}
	for _, scale := range []int{2, 3} {
		g.scale = scale
		for size := barSmall; size <= barLarge; size++ {
			g.bar.sizes = true
			menu := barFontMenu(g.barText, g.barSize, g.toolbarItems())
			g.clickBarFont((menu[0].col-1)*8+1, barH*(int(size)+1)+1)
			if g.bar.sizes || g.barSize != size || readBarSize(g.barSizePath) != size {
				t.Fatalf("字級點擊與讀回不一致：%d/%d", scale, size)
			}
			v, px := g.barGlyphs()
			if v.W != px || v.H != px || px != barGlyphPx(scale, size) {
				t.Fatalf("原生字模不符：%d/%d %+v", scale, size, v)
			}
			if (px == 12 || px == 22) && !strings.HasSuffix(v.Name, "-bar"+map[int]string{12: "12", 22: "22"}[px]+".golemfnt") {
				t.Fatalf("未使用原生工具列字模：%s", v.Name)
			}
		}
	}
	// 透過正式點擊處理函式保存設定。未登記Delay時維持1x，不能偷開加速。
	for i, m := range speedChoices {
		g.bar.speeds = true
		menu := speedMenu(g.speed.set, g.speed.actual(), g.toolbarItems())
		g.clickSpeed((menu[0].col-1)*8+1, barH*(i+2)+1)
		if g.bar.speeds || g.speed.set != m || readSpeed(g.speedPath) != m || o.ClockPercent() != 50 {
			t.Fatalf("速度點擊、讀回或安全閘門不符：%d", m)
		}
	}
	before, _ := os.ReadFile(g.speedPath)
	menu := speedMenu(g.speed.set, g.speed.actual(), g.toolbarItems())
	g.clickSpeed((menu[0].col-1)*8+1, barH+1)
	after, _ := os.ReadFile(g.speedPath)
	if string(before) != string(after) {
		t.Fatal("實際倍率資訊列不應改設定")
	}
	if err := os.WriteFile(g.barSizePath, []byte("99\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if readBarSize(g.barSizePath) != barMedium {
		t.Fatal("損壞偏好未回到中")
	}
	g.setBarSize(barLarge)
}
