package pod

import "github.com/wicanr2/dosgolem/xlate"

func (l *Live) composeDirect(scale int, missing func(rune)) []byte {
	if len(l.indexed) != 64000 || scale <= 0 {
		return nil
	}
	indexed := l.indexed
	if len(l.Direct.committed) == 64000 {
		indexed = l.Direct.committed
	}
	source := xlate.NewIndexedSource(indexed, l.pal, 320, 200, scale)
	if source == nil {
		return nil
	}
	out := make([]byte, 320*200*scale*scale*4)
	l.HDLayers.Draw(out, indexed, scale)
	if l.hdLatch != nil && scale == hdHoldScale {
		l.hdLatch.ApplyDeferred(source, l.palSig(), out)
	}
	l.localized.art(out, scale)
	l.bigpics.art(out, scale)
	l.glyphText.draw(out, scale, l.pal)
	l.Layer.Draw(out, scale, missing)
	if scale == hdHoldScale {
		if l.localized != nil && l.localized.hold != nil {
			l.localized.hold.ApplyDeferred(source, out, l.localized.region)
		}
		if l.bigpics != nil {
			l.bigpics.hold.ApplyDeferred(source, out, l.bigpics.region)
		}
		if l.hdHold != nil {
			l.hdHold.ApplyDeferred(source, out, l.hdRegion)
		}
		if l.battleHold != nil {
			l.battleHold.applyDeferred(indexed, out, l.palSig(), l.HDLayers.pendingBattleCells(), source)
		}
	}
	l.drawIntro(out, scale, missing)
	count := source.FillOriginal(out)
	if count < 0 {
		panic("non-opaque draw in deferred RGB path")
	}
	l.Direct.stats.RGBFrames++
	l.Direct.stats.OriginalRGBPixels += uint64(count)
	l.Direct.stats.SelectedRGBPixels += uint64(len(out)/4 - count)
	if l.Flash != nil && l.flashFrame != 0 {
		base := l.composeLegacy(scale, false, nil)
		l.Flash.Observe(l.flashFrame, indexed, base, out, scale)
		l.flashFrame = 0
	}
	if l.Journal.Open() {
		l.Journal.Draw(out, scale, l.Font, missing)
	}
	if l.Guide.Open() {
		l.Guide.Draw(out, scale, l.Font, missing)
	}
	if l.Map.Open() {
		l.Map.Draw(out, scale, l.Font, missing)
	}
	return out
}
