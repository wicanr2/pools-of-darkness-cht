#!/usr/bin/env python3
"""建立僅加入PCM／OPL記錄的前端捕捉副本；不修改正式引擎或原版程式。"""
import argparse
import difflib
import hashlib
import json
import re
from pathlib import Path
import subprocess


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--engine',type=Path,required=True);ap.add_argument('--out',type=Path,required=True)
    ap.add_argument('--version');ap.add_argument('--commit');a=ap.parse_args()
    if a.version:assert re.fullmatch(r'v\.\d+\.\d+\.\d+-\d{8}',a.version) and a.commit and re.fullmatch('[0-9a-f]{40}',a.commit)
    assert Path('/.dockerenv').exists() and a.out.is_dir()
    p=a.engine/'cmd/pod-play/main.go';original=p.read_text();text=original
    marker='\t"crypto/sha256"\n';assert text.count(marker)==1
    text=text.replace(marker,marker+'\tpromoBinary "encoding/binary"\n\tpromoMath "math"\n')
    marker='\t\tg.samples = g.mix.Render(g.samples[:0], tick.MachineStepsBefore, tick.MachineStepsAfter)\n'
    assert text.count(marker)==1;text=text.replace(marker,marker+'\t\tpromoPCM(g.samples)\n')
    marker='func (o observer) OPLWrite(w machine.OPLWrite)           { o.mix.OPLWrite(w) }'
    assert text.count(marker)==1
    text=text.replace(marker,'func (o observer) OPLWrite(w machine.OPLWrite) { o.mix.OPLWrite(w); promoOPL(w) }')
    text+='''
// 只讀已由正式Nuked OPL混音器產生的樣本，不取代音源或修改遊戲時間。
var promoPCMFile, promoOPLFile *os.File
func promoOutput(file **os.File, variable string) *os.File {
    path := os.Getenv(variable)
    if path == "" { return nil }
    if *file == nil {
        var err error
        *file, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
        if err != nil { fail(err) }
    }
    return *file
}
func promoPCM(samples []float32) {
    f := promoOutput(&promoPCMFile, "POD_PROMO_PCM")
    if f == nil { return }
    raw := make([]byte, 0, len(samples)*2)
    for _, v := range samples {
        if v < -1 || v > 1 || promoMath.IsNaN(float64(v)) { fail(fmt.Errorf("invalid promo PCM")) }
        raw = promoBinary.LittleEndian.AppendUint16(raw, uint16(int16(promoMath.Round(float64(v)*32767))))
    }
    if _, err := f.Write(raw); err != nil { fail(err) }
}
func promoOPL(w machine.OPLWrite) {
    f := promoOutput(&promoOPLFile, "POD_PROMO_OPL")
    if f == nil { return }
    b, err := json.Marshal(w)
    if err != nil { fail(err) }
    if _, err = f.Write(append(b, 10)); err != nil { fail(err) }
}
'''
    target=a.out/'main-promo.go';target.write_text(text)
    overlay=a.out/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(p):str(target)}})+'\n')
    (a.out/'capture.diff').write_text(''.join(difflib.unified_diff(original.splitlines(True),text.splitlines(True),fromfile='production main.go',tofile='capture main.go')))
    binary=a.out/'pod-promo-capture'
    command=['go','build','-trimpath','-buildvcs=false','-overlay',str(overlay)]
    if a.version:command+=['-ldflags',f'-s -w -X main.version={a.version} -X main.dosgolemCommit={a.commit}']
    subprocess.run([*command,'-o',str(binary),'./cmd/pod-play'],cwd=a.engine,check=True,timeout=180)
    info=dict(source_sha256=sha(p),capture_source_sha256=sha(target),binary_sha256=sha(binary),
        scope='same CONTROL frontend with read-only PCM/OPL logging; no queue or guest-state changes',
        pcm='48000 Hz, stereo signed16 little-endian; existing Nuked OPL mixer before added music',
        go_version=subprocess.check_output(['go','env','GOVERSION'],text=True).strip(),version=a.version,engine_commit=a.commit)
    (a.out/'capture-build.json').write_text(json.dumps(info,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(info,ensure_ascii=False))


if __name__=='__main__':main()
