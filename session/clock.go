package session

import "errors"

// SetClockPercent 在回合之間改機器時脈比例（dosgolem `docs/spec/429` §2.3）。
func (o *Owner) SetClockPercent(p int) error {
	if o == nil || o.machine == nil {
		return errors.New("session: Owner 尚未建立機器")
	}
	return o.machine.SetClockPercentLive(p)
}

// SetTPDelay 登記 Turbo Pascal CRT 的 DelayLoop 與 DelayCnt 線性位址（`docs/spec/429` §2.2）；loop 為 0 時取消。
func (o *Owner) SetTPDelay(loop, count uint32) error {
	if o == nil || o.machine == nil {
		return errors.New("session: Owner 尚未建立機器")
	}
	o.machine.SetTPDelay(loop, count)
	return nil
}

// ClockPercent 回目前的時脈比例。
func (o *Owner) ClockPercent() int {
	if o == nil || o.machine == nil {
		return 100
	}
	return o.machine.ClockPercent()
}
