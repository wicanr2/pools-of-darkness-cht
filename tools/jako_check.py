#!/usr/bin/env python3
"""獨立驗證日韓候選資料；不執行 GUI、字型或翻譯生成器。"""

import argparse
from collections import Counter
import copy
import csv
import hashlib
import io
import json
from pathlib import Path
import re
import sys


LANGUAGES = ("ja", "ko")
COUNTS = {
    "ecl-text": 4634, "ui-text": 980, "name": 1404,
    "composed-templates": 192, "menu-masters": 150, "host-text": 49,
    "title": 18, "cover": 3, "poster": 6, "journal": 99,
    "walkthrough": 78, "map-titles": 32, "map-guide": 7,
}
DISPLAY_FIELDS = {
    stem: ("title", "body") if stem in ("journal", "walkthrough")
    else ("title",) if stem == "map-titles"
    else ("label", "summary") if stem == "map-guide"
    else ("translation",) for stem in COUNTS
}
MACHINE_SOURCE = "machine-translation-76"
PLACEHOLDERS = re.compile(r"\{[^{}]+\}|%[ds]")
CONTROL = re.compile(r"[\x00-\x1f\x7f\ufeff\u200b]")
SCRIPTS = {
    "ja": re.compile(r"[\u1100-\u11ff\u3130-\u318f\uac00-\ud7af]"),
    "ko": re.compile(r"[\u3040-\u30ff\u3400-\u9fff\uf900-\ufaff]"),
}
SIGNED = re.compile(r"(?<![0-9])[+-]\d+")
REVIEW_FIELDS = {
    "lang", "stem", "index", "field", "kind", "source_sha256",
    "target_sha256", "verdict", "reason",
}
ACCEPTED_VERDICTS = {"equivalent-number", "layout-title-copy"}


def digest(value):
    """欄位 SHA 使用精確 UTF-8；字面反斜線 n 不轉成換行。"""
    return hashlib.sha256(value.encode("utf-8")).hexdigest()


def file_sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def warning_key(item):
    return tuple(item[name] for name in ("lang", "stem", "index", "field", "kind"))


class Check:
    def __init__(self, root, work, text, review, merge):
        self.root, self.work, self.text = root, work, text
        self.review_path, self.merge_path = review, merge
        self.errors, self.warnings, self.tables, self.inputs = [], [], [], {}
        self.review_matches = []
        self.empty = []
        self.display_annotations = []
        self.display_fixes = []
        self.negative_controls = []

    def error(self, code, **detail):
        self.errors.append(dict(code=code, **detail))

    def read_json(self, path):
        self.inputs[str(path)] = file_sha(path)
        return json.loads(path.read_text(encoding="utf-8"))

    def read_table(self, path):
        # QUOTE_NONE 必須明列；正文引號不是 TSV 的跳脫語法。
        raw = path.read_bytes()
        self.inputs[str(path)] = hashlib.sha256(raw).hexdigest()
        contents = raw.decode("utf-8")
        if b"\r" in raw:
            self.error("raw-carriage-return", path=str(path))
        reader = csv.reader(io.StringIO(contents), delimiter="\t", quoting=csv.QUOTE_NONE)
        records = list(reader)
        if not records:
            self.error("empty-table", path=str(path))
            return [], []
        header = records[0]
        if len(header) != len(set(header)):
            self.error("duplicate-header", path=str(path))
        rows = []
        for line, values in enumerate(records[1:], 2):
            if len(values) != len(header):
                self.error("column-count", path=str(path), line=line,
                           expected=len(header), actual=len(values))
                continue
            rows.append(dict(zip(header, values)))
        return header, rows

    def original_table(self, stem):
        if stem in ("ecl-text", "ui-text", "name"):
            name = stem + ".zh-TW.tsv"
        elif stem in ("composed-templates", "menu-masters"):
            name = stem + ".tsv"
        else:
            name = stem + ".en.tsv"
        return self.root / "text" / name

    def source_evidence(self, plans):
        """另查原版定位台帳，不以合併器的成功狀態代替來源驗證。"""
        for stem in COUNTS:
            table = plans["ja"][stem]
            header, rows = self.read_table(self.original_table(stem))
            if header != table["header"] or rows != [r["row"] for r in table["rows"]]:
                self.error("plan-baseline-mismatch", stem=stem)
        for stem in ("ecl-text", "ui-text", "name"):
            _, events = self.read_table(self.root / "text" / (stem + "-events.tsv"))
            by_key = {}
            for event in events:
                key = event["event_key"]
                if key in by_key:
                    self.error("duplicate-original-event", stem=stem, key=key)
                by_key[key] = event
            for index, item in enumerate(plans["ja"][stem]["rows"]):
                key = item["row"]["key"]
                source = item["changes"]["translation"]["english"]
                event = by_key.get(key)
                try:
                    length = len(source.encode("ascii"))
                except UnicodeEncodeError:
                    self.error("original-not-ascii", stem=stem, index=index)
                    continue
                if not event or event["original_sha256"] != digest(source) or int(event["original_length"]) != length:
                    self.error("original-event-mismatch", stem=stem, index=index, key=key)
        for index, item in enumerate(plans["ja"]["menu-masters"]["rows"]):
            row, source = item["row"], item["changes"]["translation"]["english"]
            if row["original_sha256"] != digest(source) or int(row["original_length"]) != len(source.encode("ascii")):
                self.error("menu-original-mismatch", index=index)

    def check_cell(self, lang, stem, index, item, field, target):
        loc = dict(lang=lang, stem=stem, index=index, field=field)
        source = item["changes"][field]["english"]
        # 組句譯文的數字參數沿正式繁中協定；原文 {n} 是捕獲語法。
        contract = item["row"][field] if stem == "composed-templates" else source
        if Counter(PLACEHOLDERS.findall(contract)) != Counter(PLACEHOLDERS.findall(target)):
            self.error("placeholder", **loc)
        if CONTROL.search(target):
            self.error("display-control", **loc)
        if SCRIPTS[lang].search(target):
            self.error("cross-language-script", **loc)
        if not target.strip():
            reason = "empty-English-source" if not source.strip() else "article-omission" if stem == "title" and item["row"]["key"] == "title.article" else None
            if reason:
                self.empty.append(dict(**loc, reason=reason))
            else:
                self.error("empty-display", **loc)
        elif not source.strip():
            self.error("unexpected-display-for-empty-source", **loc)
        # [MAP] 是參考書的可翻譯圖示說明；journal.go 原樣排正文，
        # 沒有以此字串控制資料查找。<> 同樣不是 placeholder 語法。
        if "[MAP]" in source:
            self.display_annotations.append(dict(**loc, source=source, target=target,
                                                 interpretation="reference-map-label; display only"))
        for marker in (r"\r", r"\t"):
            if source.count(marker) != target.count(marker):
                self.error("display-marker", marker=marker, **loc)
        if Counter(SIGNED.findall(source)) != Counter(SIGNED.findall(target)):
            self.error("signed-number", **loc)
        if stem in ("ecl-text", "journal", "walkthrough"):
            expected, actual = Counter(re.findall(r"\d+", source)), Counter(re.findall(r"\d+", target))
            if expected != actual:
                self.warnings.append(dict(**loc, kind="numbers", expected=dict(expected), actual=dict(actual),
                                          source_sha256=digest(source), target_sha256=digest(target)))
            expected, actual = source.count(r"\n"), target.count(r"\n")
            if expected != actual:
                self.warnings.append(dict(**loc, kind="paragraphs", expected=expected, actual=actual,
                                          source_sha256=digest(source), target_sha256=digest(target)))
        elif source.count(r"\n") != target.count(r"\n"):
            self.error("unreviewed-paragraph-change", **loc)

    def bind_reviews(self, review, merge):
        """逐條檢查定位與兩側 SHA；數量相同或名字相近均不能放行。"""
        if not isinstance(review, list):
            self.error("review-not-array")
            return
        by_key = {}
        for index, entry in enumerate(review):
            if not isinstance(entry, dict) or not REVIEW_FIELDS <= set(entry):
                self.error("review-schema", index=index)
                continue
            key = warning_key(entry)
            if key in by_key:
                self.error("duplicate-review", identity=list(key))
            by_key[key] = entry
            for field in ("source_sha256", "target_sha256"):
                if not re.fullmatch(r"[0-9a-f]{64}", entry[field]):
                    self.error("review-hash-format", index=index, field=field)
            if entry["verdict"] not in ACCEPTED_VERDICTS or not entry["reason"].strip():
                self.error("unresolved-review", identity=list(key))
        current = {warning_key(w): w for w in self.warnings}
        if len(current) != len(self.warnings):
            self.error("duplicate-computed-warning")
        for key, warning in current.items():
            entry = by_key.get(key)
            if not entry:
                self.error("missing-review", identity=list(key))
                continue
            matched = True
            for field in ("source_sha256", "target_sha256"):
                if warning[field] != entry[field]:
                    self.error("review-" + field.replace("_", "-"), identity=list(key),
                               reviewed=entry[field], actual=warning[field])
                    matched = False
            self.review_matches.append(dict(**warning, verdict=entry["verdict"], reason=entry["reason"], matched=matched))
        for key in by_key.keys() - current.keys():
            self.error("unused-review", identity=list(key))
        recorded = {warning_key(w): w for w in merge["warnings"]}
        for key in current.keys() | recorded.keys():
            old, new = recorded.get(key), current.get(key)
            if not old or not new or any(old[f] != new[f] for f in ("expected", "actual")):
                self.error("merge-warning-mismatch", identity=list(key))

    def check_display_fixes(self, plans, candidates, merge):
        """正式後校對只改顯示欄；不要求譯文仍等於前作 reused。"""
        path = self.root / "text" / "jako-term-fixes.tsv"
        _, fixes = self.read_table(path)
        if merge.get("term_fixes_sha256") != file_sha(path) or merge.get("term_fixes_applied") != len(fixes):
            self.error("merge-display-fixes-sha-or-count")
        seen = set()
        for fix in fixes:
            lang, stem, index, field = fix["lang"], fix["stem"], int(fix["index"]), fix["field"]
            loc = dict(lang=lang, stem=stem, index=index, field=field)
            identity = (lang, stem, index, field)
            if identity in seen:
                self.error("duplicate-display-fix", **loc)
            seen.add(identity)
            if field not in DISPLAY_FIELDS.get(stem, ()):
                self.error("non-display-fix", **loc)
                continue
            if digest(fix["before"]) != fix["before_sha256"]:
                self.error("display-fix-before-sha", **loc)
            target = candidates[(lang, stem)][index][field]
            matched = target == fix["after"]
            if not matched:
                self.error("display-fix-target", **loc)
            self.display_fixes.append(dict(**loc, english_term=fix["english_term"],
                                           source_sha256=digest(plans[lang][stem]["rows"][index]["changes"][field]["english"]),
                                           before_sha256=fix["before_sha256"], target_sha256=digest(target),
                                           after_sha256=digest(fix["after"]), matched=matched,
                                           reason=fix["reason"]))

    def run(self):
        plan_path = self.work / "plan.json"
        plan = self.read_json(plan_path)
        review = self.read_json(self.review_path)
        merge = self.read_json(self.merge_path)
        self.review, self.merge = review, merge
        if merge.get("status") != "MERGED; validation pending":
            self.error("merge-not-complete", status=merge.get("status"))
        if merge.get("plan_sha256") != file_sha(plan_path):
            self.error("merge-plan-sha")
        plans = {}
        for lang in LANGUAGES:
            tables = plan.get(lang, [])
            plans[lang] = {t["stem"]: t for t in tables}
            if len(tables) != 13 or set(plans[lang]) != set(COUNTS):
                self.error("plan-table-set", lang=lang)
                return
            for stem, table in plans[lang].items():
                if len(table["rows"]) != COUNTS[stem]:
                    self.error("plan-row-count", lang=lang, stem=stem, actual=len(table["rows"]), expected=COUNTS[stem])
        for stem in COUNTS:
            left, right = plans["ja"][stem], plans["ko"][stem]
            if left["header"] != right["header"] or [x["row"] for x in left["rows"]] != [x["row"] for x in right["rows"]]:
                self.error("language-baseline-mismatch", stem=stem)
            if [dict((f, v["english"]) for f, v in x["changes"].items()) for x in left["rows"]] != [dict((f, v["english"]) for f, v in x["changes"].items()) for x in right["rows"]]:
                self.error("language-English-source-mismatch", stem=stem)
        self.source_evidence(plans)
        candidates = {}
        for lang in LANGUAGES:
            for stem, expected_count in COUNTS.items():
                table = plans[lang][stem]
                path = self.text / (stem + "." + lang + ".tsv")
                header, rows = self.read_table(path)
                candidates[(lang, stem)] = rows
                if header != table["header"]:
                    self.error("header-mismatch", lang=lang, stem=stem)
                if len(rows) != expected_count:
                    self.error("row-count", lang=lang, stem=stem, expected=expected_count, actual=len(rows))
                entry = next((e for e in merge["counts"][lang] if e["path"] == path.name), None)
                if not entry or entry["sha256"] != file_sha(path) or entry["rows"] != len(rows):
                    self.error("merge-table-sha-or-count", lang=lang, stem=stem)
                for index, (item, row) in enumerate(zip(table["rows"], rows)):
                    if set(item["changes"]) != set(DISPLAY_FIELDS[stem]):
                        self.error("display-field-set", lang=lang, stem=stem, index=index)
                    for field in header:
                        if field not in item["changes"]:
                            expected = MACHINE_SOURCE if field == "source" and stem in ("ecl-text", "ui-text", "name") else item["row"][field]
                            if row[field] != expected:
                                self.error("original-field-changed", lang=lang, stem=stem, index=index, field=field,
                                           expected=expected, actual=row[field])
                    for field in DISPLAY_FIELDS[stem]:
                        self.check_cell(lang, stem, index, item, field, row[field])
                self.tables.append(dict(lang=lang, stem=stem, rows=len(rows), expected_rows=expected_count,
                                        fields=header, sha256=file_sha(path), path=str(path)))
        self.check_display_fixes(plans, candidates, merge)
        self.bind_reviews(review, merge)

    def self_test(self):
        """純記憶體負對照：不修改候選、審查檔或其他專案檔案。"""
        for name, field, expected in (
            ("stale-target-sha", "target_sha256", "review-target-sha256"),
            ("stale-source-sha", "source_sha256", "review-source-sha256"),
            ("missing-individual-review", None, "missing-review"),
        ):
            probe = Check(self.root, self.work, self.text, self.review_path, self.merge_path)
            probe.warnings = copy.deepcopy(self.warnings)
            review = copy.deepcopy(self.review)
            if not review:
                self.error("self-test-no-review-fixture")
                return
            if field:
                review[0][field] = "0" * 64
            else:
                review = review[1:]
            probe.bind_reviews(review, self.merge)
            codes = [e["code"] for e in probe.errors]
            passed = expected in codes
            self.negative_controls.append(dict(name=name, expected=expected, observed=codes, passed=passed))
            if not passed:
                self.error("self-test-failed", name=name)
        fixtures = (
            ("missing-placeholder", "ja", "Value {n} %d", "値 {n}", "placeholder"),
            ("raw-control", "ja", "caption", "表示\n", "display-control"),
            ("Korean-in-Japanese", "ja", "Magic", "마법", "cross-language-script"),
            ("CJK-in-Korean", "ko", "Magic", "魔法", "cross-language-script"),
            ("changed-signed-value", "ja", "+3 weapons", "武器 +4", "signed-number"),
        )
        for name, lang, source, target, expected in fixtures:
            probe = Check(self.root, self.work, self.text, self.review_path, self.merge_path)
            item = {"row": {"key": "test"}, "changes": {"translation": {"english": source}}}
            probe.check_cell(lang, "host-text", 0, item, "translation", target)
            codes = [e["code"] for e in probe.errors]
            passed = expected in codes
            self.negative_controls.append(dict(name=name, expected=expected, observed=codes, passed=passed))
            if not passed:
                self.error("self-test-failed", name=name)
        for name, source, target in (
            ("signed-value-next-to-CJK", "+3 weapons or better to hit", "命中には+3以上の武器が必要"),
            ("literal-angle-brackets", " <No Spell Effects>", "<呪文効果なし>"),
            ("translated-map-label", "[MAP]", "[地図]"),
        ):
            probe = Check(self.root, self.work, self.text, self.review_path, self.merge_path)
            item = {"row": {"key": "test"}, "changes": {"translation": {"english": source}}}
            probe.check_cell("ja", "host-text", 0, item, "translation", target)
            codes = [e["code"] for e in probe.errors]
            passed = not codes
            self.negative_controls.append(dict(name=name, expected="accepted-display-text", observed=codes, passed=passed))
            if not passed:
                self.error("self-test-failed", name=name)

    def report(self):
        totals = {lang: sum(t["rows"] for t in self.tables if t["lang"] == lang) for lang in LANGUAGES}
        for lang, total in totals.items():
            if total != 7652:
                self.error("language-total", lang=lang, expected=7652, actual=total)
        return dict(status="PASS" if not self.errors else "FAIL", scope="獨立資料驗證；不含GUI、字型或正常玩家路徑",
                    counts=dict(tables=len(self.tables), per_language=totals, expected_per_language=7652,
                                reviewed_warnings=len(self.review_matches)),
                    python_version=sys.version.split()[0],
                    inputs=self.inputs, tables=self.tables, intentional_empty=self.empty,
                    display_annotations=self.display_annotations,
                    display_fixes=self.display_fixes,
                    negative_controls=self.negative_controls,
                    warning_reviews=self.review_matches, errors=self.errors,
                    hash_basis="檔案SHA以原始bytes計算；欄位SHA以精確UTF-8字串計算，保留字面反斜線n。")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ("root", "work", "text", "out"):
        parser.add_argument("--" + name, type=Path, required=True)
    parser.add_argument("--numeric-review", type=Path)
    parser.add_argument("--merge", type=Path)
    parser.add_argument("--self-test", action="store_true", help="另跑記憶體負對照，不改候選")
    args = parser.parse_args()
    review = args.numeric_review or args.work / "numeric-review.json"
    merge = args.merge or args.work / "candidate" / "merge.json"
    if not args.out.parent.is_dir():
        parser.error("--out的父目錄必須已存在")
    check = Check(args.root, args.work, args.text, review, merge)
    try:
        check.run()
        if args.self_test:
            check.self_test()
    except (OSError, ValueError, KeyError, TypeError, IndexError, UnicodeError, csv.Error) as exc:
        check.error("invalid-input", exception=type(exc).__name__, detail=str(exc))
    result = check.report()
    protected = {p.resolve() for p in (review, merge, args.work / "plan.json")}
    if args.out.resolve() in protected or args.out.resolve().is_relative_to(args.text.resolve()):
        parser.error("--out不得覆寫輸入或位於候選text內")
    result["tool_sha256"] = file_sha(Path(__file__))
    args.out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"status": result["status"], "counts": result["counts"], "errors": len(result["errors"])}, ensure_ascii=False))
    return 0 if result["status"] == "PASS" else 1


if __name__ == "__main__":
    sys.exit(main())
