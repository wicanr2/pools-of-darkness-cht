package pod

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestCPICVGASourceIDs(t *testing.T) {
	for _, n := range []int{1, 108, 112, 122, 125, 126, 127, 129, 236, 240, 250, 253, 254, 255} {
		if vga, err := vgaActorReference(fmt.Sprintf("cpic1-%03d", n), []byte{8, 144}); !vga || err != nil {
			t.Fatalf("原版存在的ID %d被拒絕：%v", n, err)
		}
	}
	for _, n := range []int{-1, 0, 109, 110, 111, 123, 124, 128, 237, 238, 239, 251, 252, 256, 381} {
		if _, err := vgaActorReference(fmt.Sprintf("cpic1-%03d", n), []byte{8, 144}); err == nil {
			t.Fatalf("未知ID %d未被拒絕", n)
		}
	}
	for _, key := range []string{"125", "cpic1-0125", "cpic1-+125", "CPIC1-125"} {
		if _, err := vgaActorReference(key, []byte{8, 144}); err == nil {
			t.Fatalf("非正規key %q未被拒絕", key)
		}
	}
	if _, err := vgaActorReference("cpic1-125", []byte{8, 143}); err == nil {
		t.Fatal("不合法CPIC色號未被拒絕")
	}
	if vga, err := vgaActorReference("legacy-ega", []byte{1, 8, 15}); vga || err != nil {
		t.Fatalf("原EGA路徑改變：%v", err)
	}
}

func TestCPICVGAOriginalDirectory(t *testing.T) {
	path := os.Getenv("DOSGOLEM_POD_CPIC_DIRECTORY")
	if path == "" {
		t.Skip("未提供原版CPIC目錄證據")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var proof struct {
		OriginalDirectory struct {
			SHA256 string `json:"sha256"`
			Blocks []int  `json:"blocks"`
		} `json:"original_directory"`
	}
	if err = json.Unmarshal(data, &proof); err != nil {
		t.Fatal(err)
	}
	if proof.OriginalDirectory.SHA256 != "4e9a15fb4ba3574fb154a9fbda660a5be9e75e38b2c96ee66a3e16558a55bbec" || len(proof.OriginalDirectory.Blocks) != 244 {
		t.Fatal("不是本作244筆原始目錄")
	}
	want := map[int]bool{}
	for _, n := range proof.OriginalDirectory.Blocks {
		if want[n] {
			t.Fatal("原始目錄有重複ID")
		}
		want[n] = true
	}
	for n := 0; n <= 255; n++ {
		vga, err := vgaActorReference(fmt.Sprintf("cpic1-%03d", n), []byte{8, 144})
		if (vga && err == nil) != want[n] {
			t.Fatalf("ID %d與原始目錄不符", n)
		}
	}
}
