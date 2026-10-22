// Package tpcrt 收 Turbo Pascal 7 CRT 單元的執行期知識（dosgolem `docs/spec/429` §2.4）。
// 位址由簽章在執行期找出，不寫死任何一支程式的位址。
package tpcrt

// delaySig 是 TP7 CRT.Delay 的本體；-1 為 DelayCnt 的資料段位移（每支程式不同）。
//
//	mov bx,sp / mov dx,ss:[bx+4] / or dx,dx / jz done / mov ax,40h / mov es,ax / xor di,di / mov al,es:[di]
//	perMs: mov cx,[DelayCnt] / call DelayLoop / dec dx / jnz perMs / done: retf 2
//	DelayLoop: mov bx,4 …
var delaySig = []int{
	0x8B, 0xDC, 0x36, 0x8B, 0x57, 0x04, 0x0B, 0xD2, 0x74, 0x14, 0xB8, 0x40, 0x00, 0x8E, 0xC0, 0x33,
	0xFF, 0x26, 0x8A, 0x05, 0x8B, 0x0E, -1, -1, 0xE8, 0x06, 0x00, 0x4A, 0x75, 0xF6, 0xCA, 0x02,
	0x00, 0xBB, 0x04, 0x00, 0x4B, 0x75, 0xFD, 0x26, 0x3A, 0x05, 0xE1, 0xF5, 0xC3,
}

// delaySigTP5 是較舊版 CRT.Delay（Secret of the Silver Blades 1990 實測）：ES:DI＝0000:0000，
// 每毫秒 `mov bx,[DelayCnt] / mov cx,bx / call DelayLoop`，DelayLoop 為 `cmp al,es:[di] / jne done / loop`。
//
//	mov bx,sp / mov dx,ss:[bx+4] / or dx,dx / jz done / xor di,di / mov es,di / mov al,es:[di] / mov bx,[DelayCnt]
//	perMs: mov cx,bx / call DelayLoop / dec dx / jnz perMs / done: retf 2
//	DelayLoop: cmp al,es:[di] / jne $+4 / loop DelayLoop / ret
var delaySigTP5 = []int{
	0x8B, 0xDC, 0x36, 0x8B, 0x57, 0x04, 0x0B, 0xD2, 0x74, 0x13, 0x33, 0xFF, 0x8E, 0xC7, 0x26, 0x8A,
	0x05, 0x8B, 0x1E, -1, -1, 0x8B, 0xCB, 0xE8, 0x06, 0x00, 0x4A, 0x75, 0xF8, 0xCA, 0x02, 0x00,
	0x26, 0x3A, 0x05, 0x75, 0x02, 0xE2, 0xF9, 0xC3,
}

type delayForm struct {
	sig                 []int
	countOperand, loopAt uint32
}

var delayForms = []delayForm{
	{delaySig, 0x16, 0x21},    // TP7：mov cx,[DelayCnt] 的位移欄、DelayLoop
	{delaySigTP5, 0x13, 0x20}, // TP5：mov bx,[DelayCnt] 的位移欄、DelayLoop
}

const searchEnd = 0xA0000

// FindDelay 在常規記憶體找 CRT.Delay（TP7 或 TP5 寫法），回 DelayLoop 與 DelayCnt（ds×16＋位移）的線性位址。
// 兩種寫法合計必須恰好一處命中；否則回 ok=false，不猜位址。
func FindDelay(read func(addr uint32) uint8, ds uint16) (loop, count uint32, ok bool) {
	var hit uint32
	var form delayForm
	hits := 0
	for _, f := range delayForms {
		for a := uint32(0); a+uint32(len(f.sig)) <= searchEnd; a++ {
			match := true
			for i, b := range f.sig {
				if b >= 0 && read(a+uint32(i)) != uint8(b) {
					match = false
					break
				}
			}
			if match {
				hit, form = a, f
				hits++
			}
		}
	}
	if hits != 1 {
		return 0, 0, false
	}
	off := uint32(read(hit+form.countOperand)) | uint32(read(hit+form.countOperand+1))<<8
	return hit + form.loopAt, uint32(ds)<<4 + off, true
}
