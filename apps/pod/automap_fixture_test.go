package pod

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestAutoMapPodMemory 讀正常按鍵產生的真實快照，避免合成測試跟著錯誤位址一起通過。
// 不提交原版記憶體；缺檔明確 skip。重生入口見主專案 tools/g8-map-evidence.sh。
func TestAutoMapPodMemory(t *testing.T) {
	dir := os.Getenv("DOSGOLEM_POD_MAP_PROBE_DIR")
	text := os.Getenv("DOSGOLEM_POD_TEXT_DIR")
	if dir == "" || text == "" {
		t.Skip("需要本機 G8 原版記憶體與 text/，原版資料不隨 repo 散布")
	}
	a, err := LoadAutoMapDir(text)
	if err != nil || a == nil {
		t.Fatalf("載入地圖資料：%v", err)
	}
	for _, tc := range []struct {
		name      string
		x, y, dir int
	}{
		{"start", 11, 2, 2}, {"left", 11, 2, 1}, {"east1", 12, 2, 1},
		{"east2", 13, 2, 1}, {"south", 13, 2, 2}, {"west", 13, 2, 3},
		{"north", 13, 2, 0}, {"gate", 0, 4, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join(dir, tc.name+".mem"))
			if os.IsNotExist(err) {
				t.Skip("缺原版快照，先重跑 tools/g8-map-evidence.sh")
			}
			if err != nil || len(b) != 0x100000 {
				t.Fatalf("記憶體：%v，長度 %d", err, len(b))
			}
			before := bytes.Clone(b)
			a.Sample(fakeMem(b), defaultDS)
			if a.Key != "1/10" || a.X != tc.x || a.Y != tc.y || a.Dir != tc.dir || !a.loaded {
				t.Fatalf("地圖=%q (%d,%d) 方向=%d loaded=%v", a.Key, a.X, a.Y, a.Dir, a.loaded)
			}
			if !a.seen(tc.x, tc.y) {
				t.Fatal("未記錄走過的格子")
			}
			if !bytes.Equal(b, before) {
				t.Fatal("讀取地圖修改了原版記憶體")
			}
		})
	}
}

func TestAutoMapIgnoresInitializingBuffer(t *testing.T) {
	m := make(fakeMem, 0x100000)
	m.put16(linear(defaultDS, 0)+geoPtrOff+2, 0x5a7f)
	m[linear(defaultDS, 0)+partyXOff] = 7
	m[linear(defaultDS, 0)+partyXOff+1] = 13
	visits := 0
	a := &AutoMap{Guide: &MapGuide{ByHash: map[string]string{}}, OnVisit: func(string, int, int) { visits++ }}
	a.Sample(m, defaultDS)
	if visits != 0 || a.loaded || a.Key != "" {
		t.Fatal("把初始化緩衝區記成有效地圖")
	}
}
