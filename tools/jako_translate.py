#!/usr/bin/env python3
# 研究用試譯，輸出已淘汰；不產生正式日韓資料。
"""固定NLLB模型的離線草稿。保護符號逐項核對；不寫正式text。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import sys
import time

sys.path.insert(0,'/opt/pod-mt')
import ctranslate2
import sentencepiece

PROTECTED=re.compile(r'\{[^{}]+\}|%[ds]|\\n|<[^<>]+>|[+-]?\d+(?:[./]\d+)*')


def parts(s):
    # 每句及段落分別送入模型，避免長文件超過來源訓練長度。
    return re.split(r'(\\n|(?<=[.!?])\s+)',s)


def model_input(s):
    # 原版以全大寫儲存劇情。只在送模副本改句首大小寫，避免被當成縮寫。
    # 原始bytes、雜湊、查找鍵與正式存檔保持原樣。
    return s.capitalize() if s.upper()==s else s


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for key in ('plan','model','out'):ap.add_argument('--'+key,type=Path,required=True)
    ap.add_argument('--lang',choices=('ja','ko'),required=True);a=ap.parse_args()
    lock=json.loads((a.model/'manifest.json').read_text())
    for f in lock['files']:
        assert hashlib.sha256((a.model/f['file']).read_bytes()).hexdigest()==f['sha256'],f['file']
    sp=sentencepiece.SentencePieceProcessor(model_file=str(a.model/'sentencepiece.bpe.model'))
    translator=ctranslate2.Translator(str(a.model),device='cpu',compute_type='int8',intra_threads=4,inter_threads=1)
    target={'ja':'jpn_Jpan','ko':'kor_Hang'}[a.lang]
    plans=json.loads(a.plan.read_text())[a.lang];cache={};needed=[];seen=set()
    for table in plans:
        for row in table['rows']:
            for cell in row['changes'].values():
                if cell['reused'] is not None:continue
                for s in parts(cell['english']):
                    if not re.search('[A-Za-z]',s) or s in seen:continue
                    # 確保長句仍在固定token上限內，按原文詞界分段。
                    assert len(sp.encode(s))<=480,('source too long',len(sp.encode(s)))
                    seen.add(s);needed.append(s)
    start=time.monotonic();a.out.mkdir(parents=True,exist_ok=True)
    path=a.out/(a.lang+'-segments.json')
    if path.exists():cache=json.loads(path.read_text())
    pending=[s for s in needed if s not in cache]
    for offset in range(0,len(pending),24):
        batch=pending[offset:offset+24]
        encoded=[['eng_Latn']+sp.encode(model_input(s),out_type=str)+['</s>'] for s in batch]
        result=translator.translate_batch(encoded,target_prefix=[[target]]*len(batch),beam_size=4,max_decoding_length=512)
        for s,x in zip(batch,result):
            text=sp.decode(x.hypotheses[0][1:]).strip()
            assert text and '\t' not in text and '\n' not in text
            cache[s]=text
        temp=path.with_suffix('.tmp');temp.write_text(json.dumps(cache,ensure_ascii=False,indent=2)+'\n');temp.replace(path)
        print(a.lang,len(cache),len(needed),'seconds',round(time.monotonic()-start,1),flush=True)
    output=[];failed=[]
    for table in plans:
        rows=[]
        for item in table['rows']:
            row=dict(item['row'])
            for column,cell in item['changes'].items():
                source=cell['english'];value=cell['reused']
                if value is None:value=''.join(cache.get(s,s) for s in parts(source))
                before=sorted(PROTECTED.findall(source));after=sorted(PROTECTED.findall(value))
                if before!=after and cell['reused'] is None:failed.append(dict(stem=table['stem'],key=next(iter(row.values())),column=column,source=source,translation=value,expected=before,actual=after))
                row[column]=value
            rows.append(row)
        output.append(dict(stem=table['stem'],header=table['header'],rows=rows))
    (a.out/(a.lang+'-draft.json')).write_text(json.dumps(output,ensure_ascii=False,indent=2)+'\n')
    (a.out/(a.lang+'-review.json')).write_text(json.dumps(dict(status='DRAFT; review required',normalization='all-caps source copy to sentence case; original keys unchanged',segments=len(cache),seconds=time.monotonic()-start,model=lock,protected_mismatches=failed),ensure_ascii=False,indent=2)+'\n')
    print(a.lang,'DRAFT complete; protected mismatches',len(failed),flush=True)


if __name__=='__main__':main()
