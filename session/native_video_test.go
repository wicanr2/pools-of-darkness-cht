package session

import (
	"github.com/wicanr2/dosgolem/internal/machine"
	"testing"
)

type commandObserver struct {
	recordingObserver
	commands int
	fail     bool
}

func (o *commandObserver) VideoCommand(machine.VideoCommand) {
	o.commands++
	if o.fail {
		panic("command observer failure")
	}
}
func TestNativeVideoObserverGuardAndCleanup(t *testing.T) {
	if _, e := New(Config{NativeVideoCommands: true, Observer: &recordingObserver{}}); e == nil {
		t.Fatal("missing required command observer accepted")
	}
	for _, fail := range []bool{false, true} {
		obs := &commandObserver{fail: fail}
		o := observedOwner(t, obs)
		o.nativeVideoCommands = true
		o.machine.SetNativeVideo(true)
		o.machine.SetVideoMode(0x13)
		tick, e := runTurn(t, o, 20)
		if obs.commands == 0 {
			t.Fatal("no command callback")
		}
		if fail && e == nil && tick.RawError == nil {
			t.Fatal("command panic escaped fault gate")
		}
		if !fail && (e != nil || tick.RawError != nil) {
			t.Fatal(e, tick.RawError)
		}
		n := obs.commands
		o.machine.Write8(0xa0000, 99)
		if obs.commands != n {
			t.Fatal("callback retained after turn")
		}
		o.Close()
	}
}
