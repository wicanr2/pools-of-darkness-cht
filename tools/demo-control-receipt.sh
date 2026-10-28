#!/usr/bin/env bash
# 規格018：正常冷開機，選單不逾時、對話／戰鬥中斷、重播及返回後正常進入。
# 用法：tools/demo-control-receipt.sh <新收據名稱> [已建置binary名稱]
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAG="${1:?請指定新的收據名稱}"
BINARY="${2:-pod-play}"
for value in "$TAG" "$BINARY"; do
  case "$value" in *[!A-Za-z0-9_.-]*|'') echo '名稱只接受英數、點、底線與連字號' >&2; exit 2;; esac
done
for path in workplace/dosgolem/workplace/bin workplace/input/dos/pod workplace/font workplace/hd/theme text tools; do
  test -d "$ROOT/$path"
done
test -f "$ROOT/workplace/dosgolem/workplace/bin/$BINARY"
OUT="$ROOT/workplace/probe/$TAG"
test ! -e "$OUT" || { echo '保留既有收據，請指定新名稱' >&2; exit 2; }
mkdir -p "$OUT"
timeout 900s docker run --rm --init --network none --label pod.task=demo-control-receipt \
  --user "$(id -u):$(id -g)" --memory 2g --cpus 2 --pids-limit 256 \
  --log-opt max-size=10m --log-opt max-file=3 -e HOME=/tmp -e POD_NO_DIALOG=1 -e GOMAXPROCS=2 -e BINARY="$BINARY" \
  -v "$ROOT/workplace/dosgolem/workplace/bin:/app:ro" -v "$ROOT/workplace/input/dos/pod:/orig:ro" \
  -v "$ROOT/workplace/font:/font:ro" -v "$ROOT/text:/text:ro" -v "$ROOT/workplace/hd/theme:/theme:ro" \
  -v "$ROOT/tools:/tools:ro" -v "$ROOT/docs/spec:/spec:ro" -v "$OUT:/out" dosgolem-go-ebiten:1.24-r1 sh -eu -c '
    test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
    go version -m "/app/$BINARY" > /out/build-info.txt
    python3 -I -c '\''import hashlib,json;from pathlib import Path
def tree(root):
 return {str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest()for p in sorted(root.rglob("*"))if p.is_file()}
roots={n:tree(Path("/"+n))for n in("orig","text","font","theme")}
Path("/out/inputs.json").write_text(json.dumps(roots,sort_keys=True,indent=2)+"\n")'\''
    for name in menu story combat-repeat; do
      mkdir "/out/$name"
      case "$name" in
        menu) frames=6500; script="900: ,2050:shot,6400:shot" ;;
        story) frames=3800; script="900: ,1800:d,2050:shot,2300:x,2500:shot,3000:shot,3200:p,3700:shot" ;;
        combat-repeat) frames=10900; script="900: ,1800:d,6900:shot,7100:Esc,7900:shot,8400:shot,8500:d,8900:shot,9100:Enter,10000:shot,10300:p,10800:shot" ;;
      esac
      xvfb-run -a "/app/$BINARY" -original /orig -save "/out/$name/save" -text-dir /text \
        -font /font/pod-rounded.golemfnt -hd /theme -scale 3 -audio=false -frames "$frames" -script "$script" \
        -shot-dir "/out/$name/shots" -shot "/out/$name/final.rgba" > "/out/$name/run.log" 2>&1
    done
    python3 -I /tools/demo_control_check.py --out /out --binary "/app/$BINARY" --spec /spec/018-demo-control.md --receipt /out/summary.json
    find /out -user root -o -type d -name "*.md"
  '
