#!/usr/bin/env python3
"""固定OpenCC tw2s加詞彙表，僅轉顯示欄；在Go工具映像Python3.11離線執行。"""
import argparse
import csv
import hashlib
import io
import json
from pathlib import Path
import re
import sys
import tempfile
import zipfile

# 目標、來源與可轉欄；未列欄及列序逐字保留。
RECIPES = [
    (stem+'.zh-CN.tsv',stem+'.zh-TW.tsv',{'translation'})
    for stem in ('host-text','ui-text','ecl-text','name','poster','cover','title')
] + [
    ('journal.zh-CN.tsv','journal.zh-TW.tsv',{'title','body'}),
    ('walkthrough.zh-CN.tsv','walkthrough.zh-TW.tsv',{'title','body','source'}),
    ('map-guide.zh-CN.tsv','map-guide.zh-TW.tsv',{'label','summary'}),
    ('map-titles.zh-CN.tsv','map-geo.tsv',{'title'}),
    ('composed-templates.zh-CN.tsv','composed-templates.tsv',{'translation'}),
    ('menu-masters.zh-CN.tsv','menu-masters.tsv',{'translation'}),
    ('translit-zh-CN/translit-table.tsv','translit-table.tsv',{'zh','zh_female','zh_initial'}),
    ('translit-zh-CN/translit-names.tsv','translit-names.tsv',{'chinese'}),
    ('translit-zh-CN/translit-chars.zh-TW.tsv','translit-chars.zh-TW.tsv',{'translation'}),
    ('translit-zh-CN/translit-arpabet.tsv','translit-arpabet.tsv',set()),
]

def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def rows(p): return list(csv.reader(io.StringIO(p.read_text()),delimiter='\t',quoting=csv.QUOTE_NONE))
def ascii_only(s): return ''.join(c for c in s if ord(c)<128)

def generate(root,out):
    if sys.version_info[:2]!=(3,11): raise ValueError('固定OpenCC wheel需要容器Python3.11')
    vendor=root/'tools/vendor/opencc';lock=json.loads((vendor/'manifest.json').read_text())
    for f in lock['files']:
        if sha(vendor/f['file'])!=f['sha256']: raise ValueError('OpenCC來源SHA失配')
    wheel=next(vendor/f['file'] for f in lock['files'] if f['file'].endswith('.whl'))
    terms=rows(root/'text/zh-CN-terms.tsv')[1:]
    if len({r[0] for r in terms})!=len(terms): raise ValueError('詞彙表重複')
    with tempfile.TemporaryDirectory() as tmp:
        with zipfile.ZipFile(wheel) as z:z.extractall(tmp)
        sys.path.insert(0,tmp)
        from opencc import OpenCC
        converter=OpenCC('tw2s.json')
        def convert(value):
            # 以最長詞先分段，避免特殊詞被OpenCC拆開；詞彙輸出不再轉換。
            pattern='|'.join(re.escape(r[0]) for r in sorted(terms,key=lambda r:-len(r[0])))
            result=[];pos=0;special={r[0]:r[1] for r in terms}
            for match in re.finditer(pattern,value):
                result.extend((converter.convert(value[pos:match.start()]),special[match[0]]));pos=match.end()
            result.append(converter.convert(value[pos:]));return ''.join(result)
        outputs=[]
        for target,source,columns in RECIPES:
            source_path=root/'text'/source;table=rows(source_path);header=table[0]
            if not columns<=set(header):raise ValueError('欄位不存在：'+source)
            selected={i for i,h in enumerate(header) if h in columns};data=[header]
            for row in table[1:]:
                if len(row)!=len(header):raise ValueError('TSV欄數不符：'+source)
                new=[convert(value) if i in selected else value for i,value in enumerate(row)]
                for i,(old,value) in enumerate(zip(row,new)):
                    if i not in selected and old!=value:raise ValueError('改到非顯示欄')
                    if ascii_only(old)!=ascii_only(value):raise ValueError('ASCII或控制資料改變：'+source)
                    if re.findall(r'\{[^}]*\}|%[ds]',old)!=re.findall(r'\{[^}]*\}|%[ds]',value):raise ValueError('placeholder改變')
                data.append(new)
            if target=='host-text.zh-CN.tsv':
                extras=rows(root/'text/host-extra.zh-CN.tsv')
                if extras[0]!=header:raise ValueError('前端補充欄位不同')
                for extra in extras[1:]:
                    if extra[0] not in ('lang.name','journal.ref','journal.refdot'):raise ValueError('未知前端補充key')
                    if extra[0]=='lang.name':
                        old=next(row for row in data[1:] if row[0]==extra[0]);old[1]=extra[1]
                    elif any(row[0]==extra[0] for row in data[1:]):raise ValueError('補充key已存在')
                    else:data.append(extra)
            if target=='map-titles.zh-CN.tsv':
                # 地圖顯示表要求map、title為前兩欄；保留來源雜湊為第三欄。
                order=[header.index('map'),header.index('title'),header.index('payload_sha256')]
                data=[[row[i] for i in order] for row in data]
            text='\n'.join('\t'.join(row) for row in data)+'\n';dest=out/target
            dest.parent.mkdir(parents=True,exist_ok=True);dest.write_text(text)
            outputs.append(dict(path=target,source=source,rows=len(data)-1,fields=sorted(columns),source_sha256=sha(source_path),sha256=sha(dest)))
    return dict(version='OpenCC1.1.9',config='tw2s.json',terms_sha256=sha(root/'text/zh-CN-terms.tsv'),extra_sha256=sha(root/'text/host-extra.zh-CN.tsv'),outputs=outputs)

def main():
    ap=argparse.ArgumentParser(description=__doc__);ap.add_argument('--root',type=Path,required=True);ap.add_argument('--out',type=Path,required=True);ap.add_argument('--check',action='store_true');a=ap.parse_args()
    with tempfile.TemporaryDirectory() as tmp:
        stage=Path(tmp);proof=generate(a.root,stage)
        for row in proof['outputs']:
            p=a.out/row['path']
            if a.check:
                if not p.is_file() or sha(p)!=row['sha256']:raise ValueError('簡中資料需重建：'+row['path'])
            else:p.parent.mkdir(parents=True,exist_ok=True);p.write_bytes((stage/row['path']).read_bytes())
    print(json.dumps(proof,ensure_ascii=False,indent=2))

if __name__=='__main__':main()
