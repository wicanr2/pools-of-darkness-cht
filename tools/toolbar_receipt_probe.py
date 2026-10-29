#!/usr/bin/env python3
"""只在隔離副本附加前端讀值及點擊契約測試，正式cmd/pod-play保持原樣。"""
import shutil
from pathlib import Path

src = Path('/module/cmd/pod-play')
dst = Path('/module/workplace/toolbar-receipt-cmd')
if dst.exists():
    shutil.rmtree(dst)
dst.mkdir()
for p in src.glob('*.go'):
    shutil.copy2(p, dst / p.name)
main = dst / 'main.go'
s = main.read_text()
needle = 'func (g *game) screenshot() {'
assert s.count(needle) == 1
s = s.replace(needle, needle + '\n\tg.observeToolbar("shot")', 1)
needle = '\tebiten.SetWindowTitle(windowTitle)'
assert s.count(needle) == 1
s = s.replace(needle, '\tg.observeToolbar("startup")\n' + needle, 1)
main.write_text(s)
(dst / 'toolbar_observe.go').write_text('''package main
import("encoding/json";"os")
func(g *game)observeToolbar(event string){
 path:=os.Getenv("POD_TOOLBAR_OBSERVE");if path==""{return}
 font,px:=g.barGlyphs()
 d,err:=g.owner.Digest();if err!=nil{panic(err)}
 f,err:=os.OpenFile(path,os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);if err!=nil{panic(err)};defer f.Close()
 err=json.NewEncoder(f).Encode(map[string]any{"event":event,"frame":g.frame,"scale":g.scale,"size":g.barSize,"font_name":font.Name,"font_width":font.W,"font_height":font.H,"glyph_px":px,"language":g.langs[g.lang].Code,"speed":g.speed.set,"actual":g.speed.actual(),"clock":g.owner.ClockPercent(),"budget":g.budget,"delay_ready":g.speed.ready,"steps":d.Steps})
 if err!=nil{panic(err)}
}
''')
shutil.copy2('/tools/toolbar_acceptance_test.go', dst / 'toolbar_acceptance_test.go')
