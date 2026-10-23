#!/usr/bin/env bash
# 珍157 掃描的 PaddleOCR 草稿（docs/re/005）。沿用黑暗之后的 dqk-ocr:ppocr5-r1 映像與模型快取（唯讀掛載），
# tools/ocr_batch.py、ocr_sample.py 複製自銀色匕首之謎。輸出未校對，只寫 workplace/。
# 用法：tools/ocr-batch.sh <輸入目錄（workplace 內）> <輸出目錄（workplace 內）>
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CACHE="$ROOT/../dark_queen_of_krynn/workplace/walkthrough/paddle-home"
INPUT="$(cd "$1" && pwd)"
OUT="$(mkdir -p "$2" && cd "$2" && pwd)"
test -d "$CACHE/.paddlex/official_models/PP-OCRv5_mobile_det"
test -d "$CACHE/.paddlex/official_models/PP-OCRv5_mobile_rec"
IMAGE_ID="$(docker image inspect dqk-ocr:ppocr5-r1 --format '{{.Id}}')"
timeout "${POD_OCR_TIMEOUT:-3600s}" docker run --rm --init --network none \
  --memory 2g --cpus 2 --pids-limit 128 --user "$(id -u):$(id -g)" \
  --log-opt max-size=10m --log-opt max-file=3 \
  -v "$INPUT:/input:ro" -v "$CACHE:/cache:ro" -v "$ROOT/tools:/tools:ro" \
  -v "$OUT:/out" -e HOME=/cache -e OMP_NUM_THREADS=1 -e DQK_IMAGE_ID="$IMAGE_ID" \
  "$IMAGE_ID" python3 /tools/ocr_batch.py /input /out
