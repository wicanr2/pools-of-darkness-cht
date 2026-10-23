package dos

import (
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
)

// 這一份釘 `docs/spec/431`：疊底程式沒配過記憶體就常駐時，RootTSRShrink 讓下一支從 PSP＋保留量接著放。
// 預設行為（197 §5.1 的 max）由 TestInt27RootThenQueuedProgramLandsAboveResident 釘住，這裡不重複。

// rootInt27Shrink 同 rootInt27，但打開 RootTSRShrink，並把 rootFree 設成頂層 .COM 的名義段頂。
func rootInt27Shrink(t *testing.T, freeSeg uint16) *DOS {
	t.Helper()
	m, d := newTest(t)
	d.RootTSRShrink = true
	d.rootFree = 0x1100 // LoadCOM：PSPSeg＋1000h
	d.freeSeg = freeSeg
	writeChild(t, d, "NEXT.COM", queuedNext())
	d.Enqueue("NEXT.COM", "")
	m.CPU.R[cpu.DX] = 0x8000 // int 27h：保留 8000h bytes ＝ 800h 段
	call(m, d, 0x27, 0)
	return d
}

// TestRootTSRShrinkReclaimsTheGap：沒配過記憶體（freeSeg＝rootFree），常駐 800h 段，佇列程式落在 0901h。
func TestRootTSRShrinkReclaimsTheGap(t *testing.T) {
	d := rootInt27Shrink(t, 0x1100)
	if d.curPSP != 0x0901 {
		t.Errorf("佇列程式 PSP=%04X，預期 0901（0100＋0800＋1）", d.curPSP)
	}
}

// TestRootTSRShrinkKeepsAllocations：常駐前配過記憶體（freeSeg 已往上推），照 max，不蓋掉那些塊。
func TestRootTSRShrinkKeepsAllocations(t *testing.T) {
	d := rootInt27Shrink(t, 0x1200)
	if d.curPSP != 0x1201 {
		t.Errorf("佇列程式 PSP=%04X，預期 1201（freeSeg 1200 之上）", d.curPSP)
	}
}
