"""把 tools/inventory.py 的 inventory.json 轉成逐檔雜湊表（docs/re/001-input-files.tsv）。

用法（Docker 內）：python3 -I tools/inventory_tsv.py <inventory.json> <輸出 tsv>
欄位：group、path、size、sha256。group 為 game、save、manual、walkthrough。
"""
import json, sys

inv = json.load(open(sys.argv[1]))
rows = []
for f in inv["game"]["files"]:
    rows.append(("save" if f["path"].startswith("SAVE/") else "game", f["path"], f["size"], f["sha256"]))
for f in inv["manual"]["files"]:
    rows.append(("manual", f["path"], f["size"], f["sha256"]))
for f in inv["walkthrough"]["files"]:
    rows.append(("walkthrough", f["path"], f["size"], f["sha256"]))
with open(sys.argv[2], "w", encoding="utf-8") as o:
    o.write("group\tpath\tsize\tsha256\n")
    for r in rows:
        o.write("\t".join(map(str, r)) + "\n")
print(len(rows), "rows")
