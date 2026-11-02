package main

import (
	"flag"
	"fmt"
	"github.com/wicanr2/dosgolem/host"
)

// Receipt-only flags: real Ebitengine input can accompany a bounded script.
var receiptHostInput = flag.Bool("receipt-host-input", false, "有界收據亦讀取真實視窗輸入")
var pointerAudit = flag.Bool("pointer-audit", false, "記錄指標快照及已提交DOS呼叫數")

func (g *game) auditPointer(c *host.PointerCapture, calls uint64) {
	if !*pointerAudit || c == nil || g.lastPointer != nil && *g.lastPointer == *c {
		return
	}
	copy := *c
	g.lastPointer = &copy
	fmt.Printf("pointer frame=%d x=%d y=%d buttons=%d target=%d focused=%t dos_calls=%d\n", g.frame, c.X, c.Y, c.Buttons, c.Target, c.Focused, calls)
}
