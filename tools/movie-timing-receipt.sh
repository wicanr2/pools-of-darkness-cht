#!/usr/bin/env bash
# 正常session回合、50%時脈、無輸入；保留全黑DAC的字幕色號轉換。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TAG="${1:-intro-timing}"
case "$TAG" in *[!A-Za-z0-9_-]*|'') echo '收據名稱不合法' >&2; exit 2;; esac
for d in workplace/dosgolem/workplace/bin workplace/dosgolem/workplace/gocache workplace/dosgolem/workplace/gomodcache workplace/input/dos/pod workplace/probe/g2-receipt/scratch; do test -d "$ROOT/$d"; done
for f in pool4.cfg snd.com; do test -f "$ROOT/workplace/probe/g2-receipt/scratch/$f"; done
test -f "$ROOT/tools/movie_timing_probe.go"
OUT="$ROOT/workplace/probe/$TAG"
test ! -e "$OUT" || { echo '保留舊收據，請另定名稱' >&2; exit 2; }
mkdir -p "$OUT" "$ROOT/workplace/dosgolem/workplace/movieprobe"
timeout 600s docker run --rm --label pod.task=movie-timing --network none --memory 2g --cpus 2 --pids-limit 128 \
 --user "$(id -u):$(id -g)" --log-opt max-size=10m --log-opt max-file=3 \
 -v "$ROOT/workplace/dosgolem:/go-src" -v "$ROOT/tools:/tools:ro" \
 -v "$ROOT/workplace/input/dos/pod:/orig:ro" -v "$ROOT/workplace/probe/g2-receipt/scratch:/seed:ro" -v "$OUT:/out" \
 -e GOMAXPROCS=2 -e GOCACHE=/go-src/workplace/gocache -e GOMODCACHE=/go-src/workplace/gomodcache \
 dosgolem-go-ebiten:1.24-r1 sh -eu -c '
 test "$(stat -c %u:%g /out)" = "$(id -u):$(id -g)"
 cp /tools/movie_timing_probe.go /go-src/workplace/movieprobe/main.go
 cd /go-src
 go build -o workplace/bin/movieprobe ./workplace/movieprobe
 python3 -I -c '\''import hashlib,json;from pathlib import Path
sha=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
record={"original":{str(p.relative_to("/orig")):sha(p) for p in sorted(Path("/orig").rglob("*")) if p.is_file()},"seed":{n:sha(Path("/seed")/n) for n in ("pool4.cfg","snd.com")},"probe_sha256":sha(Path("/go-src/workplace/bin/movieprobe")),"source_sha256":sha(Path("/tools/movie_timing_probe.go"))}
Path("/out/inputs.json").write_text(json.dumps(record,indent=2)+"\n")'\''
 workplace/bin/movieprobe -orig /orig -seed /seed -out /out -frames 5100 > /out/run.log 2>&1
 find /out -user root -o -type d -name "*.md"
 '
