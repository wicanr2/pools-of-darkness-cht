package session

import (
	"github.com/wicanr2/dosgolem/host"
	"github.com/wicanr2/dosgolem/internal/dos"
	"testing"
)

func TestOwnerPointerSnapshotsCommitAndPanelCleanup(t *testing.T) {
	o := startSyntheticOwner(t, []byte{0x90, 0xeb, 0xfd})
	defer o.Close()
	o.pointerMode = true
	deliver := func(c *host.PointerCapture, events []host.PanelEvent) InputReceipt {
		t.Helper()
		v := ownerView(t, o)
		r, e := o.Deliver(CapturedUpdate{StartedPanel: v.Panel, Layout: v.Layout, SourceGeneration: v.SourceGeneration, Pointer: c, PanelEvents: events})
		if e != nil {
			t.Fatal(e)
		}
		if _, e = o.Advance(1); e != nil {
			t.Fatal(e)
		}
		return r
	}
	c := host.PointerCapture{X: 320, Y: 236, Buttons: 3, Target: host.MouseTargetCanvas, Focused: true}
	if r := deliver(&c, nil); r.DOSCallsCommitted != 3 || o.dos.Mouse.Buttons != 3 || o.dos.Mouse.X != 160 || o.dos.Mouse.Y != 100 {
		t.Fatalf("initial %+v mouse=%+v", r, o.dos.Mouse)
	}
	if r := deliver(&c, nil); r.DOSCallsCommitted != 0 {
		t.Fatal("duplicate", r)
	}
	deliver(nil, []host.PanelEvent{{Kind: host.PanelEventOpen}})
	if o.dos.Mouse.Buttons != 0 || o.pointer.Accepted != 0 {
		t.Fatal("panel held button")
	}
}

func TestOwnerPointerSnapshotRejectsWholeBatch(t *testing.T) {
	o := startSyntheticOwner(t, []byte{0x90})
	defer o.Close()
	o.pointerMode = true
	v := ownerView(t, o)
	c := host.PointerCapture{X: 320, Y: 236, Buttons: 4, Target: host.MouseTargetCanvas, Focused: true}
	_, e := o.Deliver(CapturedUpdate{StartedPanel: v.Panel, Layout: v.Layout, SourceGeneration: v.SourceGeneration, Pointer: &c, BIOSKeys: []dos.Key{{ASCII: 'X'}}})
	if e == nil || o.dos.KeysPending() != 0 || o.dos.Mouse.Buttons != 0 || o.machine.Steps != 0 {
		t.Fatal("partial invalid commit", e)
	}
}
