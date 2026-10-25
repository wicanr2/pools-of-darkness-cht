#!/usr/bin/env bash
# BIGPIC只讀來源觀測，沿既有DEMO兩次空白鍵，不注入狀態或座標。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAG="${1:-b25-source}"
case "$TAG" in *[!A-Za-z0-9_-]*|'') echo '收據名稱不合法' >&2; exit 2;; esac
for d in workplace/dosgolem workplace/probe/g9-hd workplace/input/dos/pod workplace/probe/g2-receipt/scratch workplace/hd/inventory tools; do test -d "$ROOT/$d"; done
for d in workplace/dosgolem/workplace/gocache workplace/dosgolem/workplace/gomodcache; do test -d "$ROOT/$d"; done
test -f "$ROOT/tools/bigpic_source_probe.go"
test ! -e "$ROOT/workplace/probe/g9-hd/$TAG" || { echo '請保留舊收據並改名' >&2; exit 2; }
timeout 1100s docker run --rm --network none --label pod.task=bigpic-source \
  --user "$(id -u):$(id -g)" --memory 2g --cpus 2 --pids-limit 128 \
  --log-opt max-size=10m --log-opt max-file=3 -e HOME=/tmp -e GOMAXPROCS=2 -e TAG="$TAG" \
  -e GOCACHE=/src/workplace/gocache -e GOMODCACHE=/src/workplace/gomodcache -e GOPROXY=off -e GOSUMDB=off \
  -v "$ROOT/workplace/dosgolem:/src" -v "$ROOT/tools:/tools:ro" \
  -v "$ROOT/workplace/input/dos/pod:/orig:ro" -v "$ROOT/workplace/probe/g2-receipt/scratch:/seed:ro" \
  -v "$ROOT/workplace/hd/inventory:/inventory:ro" -v "$ROOT/workplace/probe/g9-hd:/out" \
  -w /src dosgolem-go-ebiten:1.24-r1 sh -eu -c '
    test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
    for path in workplace/bigpicprobe workplace/bigpicprobe/main.go workplace/bin workplace/bin/bigpicprobe; do
      if test -e "$path"; then test "$(stat -c %u:%g "$path")" = "$(id -u):$(id -g)"; fi
    done
    mkdir -p workplace/bigpicprobe
    cp /tools/bigpic_source_probe.go workplace/bigpicprobe/main.go
    timeout 120 go build -o workplace/bin/bigpicprobe ./workplace/bigpicprobe
    python3 -I -c '\''import hashlib,json,os;from pathlib import Path
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
r=dict(scope="source-only normal cold DEMO; spaces900/1800; no state injection",probe_sha256=sha(Path("/src/workplace/bin/bigpicprobe")),source_sha256=sha(Path("/tools/bigpic_source_probe.go")),catalog_sha256=sha(Path("/inventory/inventory.json")),original={n:sha(Path("/orig")/n)for n in ("ADLIB.COM","CONTROL.EXE","GAME.EXE","GAME.OVR","VGA.DRV","BIGPIC1.DAX")},seed={n:sha(Path("/seed")/n)for n in ("pool4.cfg","snd.com")},frames=26000,budget=96250)
Path("/out",os.environ["TAG"]+".inputs.json").write_text(json.dumps(r,ensure_ascii=False,indent=2)+"\n")'\''
    timeout 900 workplace/bin/bigpicprobe -orig /orig -seed /seed -inventory /inventory -out "/out/$TAG" -frames 26000 > "/out/$TAG.log" 2>&1
    find "/out/$TAG" -user root -o -type d -name "*.md"
  '
