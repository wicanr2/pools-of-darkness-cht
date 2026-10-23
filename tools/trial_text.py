"""G3 覆繪驗證用的佔位譯文：從前作 ui-text.zh-TW.tsv 取 key 與本作 catalog 相同的條目（docs/re/004）。

用法（Docker 內）：python3 -I tools/trial_text.py <本作 ui-text-events.tsv> <輸出 tsv> <標籤>=<前作 tsv> ...
依參數順序取第一個有譯文的前作；source 欄記 trial-<標籤>。輸出只放 workplace/text-src/trial/，不是本作譯文。
"""
import sys

events, out = sys.argv[1], sys.argv[2]
keys = {l.split("\t")[0] for l in open(events, encoding="utf-8").read().split("\n")[1:] if l}
got = {}
for spec in sys.argv[3:]:
    tag, path = spec.split("=", 1)
    for line in open(path, encoding="utf-8").read().split("\n")[1:]:
        if not line:
            continue
        k, t, *_ = line.split("\t")
        if k in keys and k not in got:
            got[k] = (t, f"trial-{tag}")
with open(out, "w", encoding="utf-8") as f:
    f.write("key\ttranslation\tsource\n")
    for k, (t, s) in sorted(got.items()):
        f.write(f"{k}\t{t}\t{s}\n")
print(len(got), "trial entries", out)
