#!/usr/bin/env bash
# 沿既有G2常駐驅動及CONTROL入口，不跳片頭；只匯出原始畫格與色盤。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAG="${1:-intro-source}"
case "$TAG" in *[!A-Za-z0-9_-]*|'') echo '收據名稱不合法' >&2; exit 2;; esac
for d in workplace/dosgolem/workplace/bin workplace/input/dos/pod workplace/probe/g2-receipt/scratch; do test -d "$ROOT/$d"; done
for f in probe; do test -x "$ROOT/workplace/dosgolem/workplace/bin/$f"; done
for f in pool4.cfg snd.com; do test -f "$ROOT/workplace/probe/g2-receipt/scratch/$f"; done
OUT="$ROOT/workplace/probe/$TAG"
test ! -e "$OUT" || { echo '保留舊收據，請另定名稱' >&2; exit 2; }
mkdir -p "$OUT"
timeout 960s docker run --rm --label pod.task=movie-source --network none --memory 2g --cpus 2 --pids-limit 128 \
  --user "$(id -u):$(id -g)" --log-opt max-size=10m --log-opt max-file=3 \
  -v "$ROOT/workplace/dosgolem/workplace/bin:/app:ro" -v "$ROOT/workplace/input/dos/pod:/orig:ro" \
  -v "$ROOT/workplace/probe/g2-receipt/scratch:/seed:ro" -v "$OUT:/out" -e HOME=/tmp -e GOMAXPROCS=2 \
  dosgolem-go-ebiten:1.24-r1 python3 -I -c '
import hashlib,json,shutil,subprocess
from pathlib import Path
out=Path("/out");assert out.stat().st_uid==1000
(out/"scratch").mkdir();(out/"shots").mkdir()
for name in ("pool4.cfg","snd.com"):shutil.copy2(Path("/seed")/name,out/"scratch"/name)
shots=",".join(f"{step}:/out/shots/f{step:010d}.raw" for step in range(20000000,1200000000,20000000))
args=["/app/probe","-root","/orig","-exe","/orig/ADLIB.COM","-program-path",r"C:\POD\SND.COM",
      "-queue","CONTROL.EXE","-root-tsr-shrink","-scratch-dir","/out/scratch","-steps","1200000000",
      "-shots",shots,"-save-state","1199999000:/out/end.state","-log-calls"]
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
inputs={name:sha(Path("/orig")/name) for name in ("CONTROL.EXE","VACE.EXE","VGA.DRV","INTRO.ACF","INTRO.BMA","P1.SMS","ADLIB.COM")}
record=dict(scope="既有G2正常冷啟動，常駐snd.com與CONTROL鏈；不送跳過鍵；只做來源盤點",inputs=inputs,
            probe_sha256=sha(Path("/app/probe")),seed={n:sha(Path("/seed")/n) for n in ("pool4.cfg","snd.com")},command=args)
(out/"inputs.json").write_text(json.dumps(record,ensure_ascii=False,indent=2)+"\n")
with (out/"run.log").open("w") as log:subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,check=True,timeout=900)
print("原始片頭來源擷取結束",len(list((out/"shots").glob("*.raw"))),"格；不代表字幕或HD已接入")
'
