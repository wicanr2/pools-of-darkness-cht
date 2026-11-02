package host

import "fmt"

// PointerCapture is a single frontend sample. Coordinates use MouseLayout's
// output pixels; Buttons contains left (1) and right (2). Spec 437.
type PointerCapture struct {
	X, Y    int
	Buttons uint8
	Target  MouseTarget
	Focused bool
}

type PointerSnapshot struct {
	Held, Accepted uint8
	X, Y           int
	HasPosition    bool
}

type PointerPlan struct {
	Next    PointerSnapshot
	Actions []MouseAction
}

// PlanPointerCapture is pure. Remember hardware holds even outside the canvas
// so dragging back never creates a new press. Releases always reach DOS.
func PlanPointerCapture(s PointerSnapshot, layout MouseLayout, c PointerCapture) (PointerPlan, error) {
	p := PointerPlan{Next: s}
	if !layout.valid() || c.Buttons&^3 != 0 || s.Accepted&^s.Held != 0 || s.Held&^3 != 0 {
		return p, fmt.Errorf("host: invalid pointer snapshot")
	}
	if c.Target != MouseTargetCanvas && c.Target != MouseTargetHost && c.Target != MouseTargetOutside {
		return p, fmt.Errorf("host: invalid pointer target")
	}
	accept := c.Focused && !layout.PanelOpen && c.Target == MouseTargetCanvas && layout.canvasContains(c.X, c.Y)
	if accept {
		x, y := c.X/int(layout.Scale), (c.Y-layout.ChromeHeight)/int(layout.Scale)
		if !s.HasPosition || s.X != x || s.Y != y {
			p.Actions = append(p.Actions, MouseAction{Kind: MouseActionMove, X: x, Y: y})
			p.Next.X, p.Next.Y, p.Next.HasPosition = x, y, true
		}
	}
	for button := 0; button < 2; button++ {
		bit := uint8(1 << button)
		if s.Accepted&bit != 0 && (!accept || c.Buttons&bit == 0) {
			p.Actions = append(p.Actions, MouseAction{Kind: MouseActionRelease, Button: button})
			p.Next.Accepted &^= bit
		}
		if accept && c.Buttons&bit != 0 && s.Held&bit == 0 {
			p.Actions = append(p.Actions, MouseAction{Kind: MouseActionPress, Button: button})
			p.Next.Accepted |= bit
		}
	}
	p.Next.Held = c.Buttons
	return p, nil
}
