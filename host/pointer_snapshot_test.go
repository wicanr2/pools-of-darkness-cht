package host

import "testing"

func TestPointerTwoButtonsAndCapture(t *testing.T) {
	l := MouseLayout{Epoch: 1, Scale: 2, ChromeHeight: 36, Canvas: Canvas{Width: 320, Height: 200}, FrameWidth: 640, FrameHeight: 436}
	s := PointerSnapshot{}
	run := func(c PointerCapture) []MouseAction {
		t.Helper()
		p, e := PlanPointerCapture(s, l, c)
		if e != nil {
			t.Fatal(e)
		}
		s = p.Next
		return p.Actions
	}
	c := PointerCapture{X: 320, Y: 236, Buttons: 3, Target: MouseTargetCanvas, Focused: true}
	a := run(c)
	if len(a) != 3 || a[0].X != 160 || a[0].Y != 100 || a[1].Button != 0 || a[2].Button != 1 {
		t.Fatalf("both %v", a)
	}
	if a = run(c); len(a) != 0 {
		t.Fatalf("duplicate %v", a)
	}
	c.X = 322
	if a = run(c); len(a) != 1 || a[0].Kind != MouseActionMove {
		t.Fatalf("move %v", a)
	}
	c.Target = MouseTargetHost
	if a = run(c); len(a) != 2 || a[0].Kind != MouseActionRelease || s.Accepted != 0 {
		t.Fatalf("leave %v", a)
	}
	c.Target = MouseTargetCanvas
	if a = run(c); len(a) != 0 {
		t.Fatalf("drag back %v", a)
	}
	c.Buttons = 0
	run(c)
	c.Buttons = 2
	run(c)
	c.Focused = false
	if a = run(c); len(a) != 1 || a[0].Button != 1 || a[0].Kind != MouseActionRelease {
		t.Fatalf("focus %v", a)
	}
	c.Focused = true
	if a = run(c); len(a) != 0 {
		t.Fatalf("focus hold %v", a)
	}
	c.Buttons = 0
	run(c)
	c.Buttons = 3
	l.PanelOpen = true
	if a = run(c); len(a) != 0 {
		t.Fatalf("panel %v", a)
	}
	l.PanelOpen = false
	if a = run(c); len(a) != 0 {
		t.Fatalf("panel drag %v", a)
	}
	c.Buttons = 4
	if _, e := PlanPointerCapture(s, l, c); e == nil {
		t.Fatal("invalid buttons accepted")
	}
}
