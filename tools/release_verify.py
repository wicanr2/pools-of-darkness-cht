#!/usr/bin/env python3
"""獨立解開十二個實際封包，核對內容、平台、來源與執行收據。"""
import argparse
import hashlib
import json
from pathlib import Path
import plistlib
import re
import struct
import subprocess
import tarfile
import tempfile
import zipfile
import numpy as np
from PIL import Image


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):return json.loads(p.read_text())
def tree(p):return {str(q.relative_to(p)):sha(q) for q in sorted(p.rglob('*')) if q.is_file()}
def state(log):
    a=re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',log)
    b=re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)',log);assert a and b
    return list(a.groups())+list(b.groups())


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for n in ('root','work','receipt'):ap.add_argument('--'+n,type=Path,required=True)
    ap.add_argument('--content-only',action='store_true')
    ap.add_argument('--version',default='v.0.1.2-20261102')
    a=ap.parse_args();r,w=a.root,a.work;meta=read(w/'packages/packages.json');lock=read(w/'main/review/hd/runtime/manifest.json');source=read(w/'release-source.json')
    assert re.fullmatch(r'v\.\d+\.\d+\.\d+-\d{8}',a.version)
    assert meta['version']==source['version']==a.version and len(meta['files'])==12
    expected={x['path']:x['sha256'] for x in lock['theme']};fonts={x['path']:x['sha256'] for x in lock['fonts']}
    assert len(expected)==2034 and len(fonts)==18 and meta['engine_commit']==lock['engine']['commit']==source['engine_commit']
    original={p.name:sha(p) for p in (r/'workplace/input/dos/pod').iterdir() if p.is_file()};assert len(original)==47
    records=[]
    for row in meta['files']:
        arc=w/'packages'/row['name'];assert sha(arc)==row['sha256'] and arc.stat().st_size==row['size']
        with tempfile.TemporaryDirectory(dir=w,prefix='verify-') as tmp:
            tmp=Path(tmp)
            if row['platform']=='linux':
                subprocess.run([str(arc),'--appimage-extract'],cwd=tmp,stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,timeout=90,check=True)
                top=tmp/'squashfs-root';data=top/'usr/bin';exe=data/'pod-play'
                b=exe.read_bytes();assert b[:4]==b'\x7fELF' and struct.unpack_from('<H',b,18)[0]==62
            else:
                with zipfile.ZipFile(arc) as z:
                    assert z.testzip() is None and all(i.flag_bits&0x800 for i in z.infolist() if not i.filename.isascii())
                    assert all(not Path(i.filename).is_absolute() and '..' not in Path(i.filename).parts for i in z.infolist())
                    if row['platform']=='macos':assert any(i.filename.endswith('/Contents/MacOS/pod-play') and (i.external_attr>>16)&0o111 for i in z.infolist())
                    z.extractall(tmp)
                top=next(tmp.iterdir());data=top if row['platform']=='windows' else top/'PODCHT.app/Contents/Resources'
                exe=top/'PODCHT.exe' if row['platform']=='windows' else top/'PODCHT.app/Contents/MacOS/pod-play';b=exe.read_bytes()
                if row['platform']=='windows':
                    off=struct.unpack_from('<I',b,0x3c)[0];assert b[:2]==b'MZ' and b[off:off+4]==b'PE\0\0' and struct.unpack_from('<H',b,off+4)[0]==0x8664
                else:
                    magic,n=struct.unpack_from('>II',b);assert magic in (0xcafebabe,0xcafebabf) and n==2
                    stride=20 if magic==0xcafebabe else 32;assert {struct.unpack_from('>I',b,8+i*stride)[0] for i in range(2)}=={0x1000007,0x100000c}
                    pl=plistlib.loads((top/'PODCHT.app/Contents/Info.plist').read_bytes());assert pl['CFBundleVersion']==meta['version']
            assert meta['version'].encode() in b and source['engine_commit'].encode() in b
            m=read(top/'PACKAGE-MANIFEST.json');actual=tree(top);actual.pop('PACKAGE-MANIFEST.json')
            assert actual==m['files'] and sha(top/'PACKAGE-MANIFEST.json')==row['payload_manifest_sha256']
            for k in ('version','platform','kind','edition','engine_commit'):assert m[k]==row[k]
            if not meta['version'].startswith('v.0.'):
                assert m['candidate'] is False
                readme=(data/'README.txt').read_text(encoding='utf-8-sig')
                assert '正式版' in readme and '候選' not in readme and '試玩' not in readme
            assert tree(data/'font')==fonts and m['fonts']==18
            assert (data/'README.txt').read_bytes().startswith(b'\xef\xbb\xbf')
            for n in ('ja','ko','zh-CN'):
                assert (data/f'text/ui-text.{n}.tsv').is_file(),n
            assert (data/'text/host-text.en.tsv').is_file()
            for f in ('LICENSE','THIRD-PARTY.md','licenses/COPYING.LGPL','licenses/OFL-GenSenRounded.txt','licenses/Go-LICENSE','source/nukedopl/COPYING.LGPL'):
                assert (data/f).is_file(),f
            if row['kind']=='full-local':
                assert m['original_files']==47 and tree(data/'original')==original and not(data/'original/SAVE').exists()
            else:
                assert m['original_files']==0 and not(data/'original').exists() and not(set(actual.values())&set(original.values()))
                assert not any(Path(p).suffix.lower() in ('.dax','.bma','.acf','.idx','.ops','.rom') for p in actual)
            if row['edition']=='hd':
                assert m['theme_files']==2034
                if row['kind']=='full-local':assert tree(data/'hd')==expected
                else:
                    src={x['source']:x['sha256'] for x in lock['theme'] if x['kind']=='repo-file'}
                    src['manifest.json']=sha(w/'main/review/hd/runtime/manifest.json')
                    assert tree(data/'hd-source')==src and not(data/'hd').exists()
            else:assert m['theme_files']==0 and not(data/'hd').exists() and not(data/'hd-source').exists()
            with tarfile.open(data/'source/dosgolem.tar.gz') as t:
                for member in t.getmembers():
                    assert not member.issym() and not member.islnk()
                    if member.isfile():
                        assert Path(member.name).suffix.lower() not in ('.exe','.com','.dax','.bma','.acf','.idx','.ops','.rom')
                        assert hashlib.sha256(t.extractfile(member).read()).hexdigest() not in original.values()
            records.append(dict(row,content_verified=True,embedded_binary_sha256=sha(exe)))
    libs=(w/'bin/macos/libraries.txt').read_text();assert 'architecture x86_64' in libs and 'architecture arm64' in libs
    assert all(s.strip().startswith(('/usr/lib/','/System/Library/')) for s in libs.splitlines() if s.startswith('\t'))
    if a.content_only:
        result=dict(status='PASS archive content only',version=meta['version'],source=source,files=records,
                    theme_files=len(expected),fonts=len(fonts),original_files_in_patch=0,source_archives_scanned=True,
                    original_game_files=original,macos='universal structure only; native unverified')
        a.receipt.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print('PASS 12 actual archives; runtime separately recorded');return
    smoke=w/'smoke';cases={n:read(smoke/n/'pass.json') for n in ('linux-intro','linux-base','linux-full','linux-demo','linux-save','linux-missing','windows')}
    assert all(c['status']=='PASS' and c['version']==meta['version'] for c in cases.values())
    assert cases['linux-intro']['first']['cache']==cases['windows']['cache']==expected
    # 五語海報與F12中央的獨立像素比對，來源取實際封包截圖。
    p=smoke/'linux-intro';samples={1000:('zh-TW',3),1010:('en',3),1020:('ja',3),1030:('ko',3),1040:('zh-CN',3),1050:('zh-TW',3),1070:('zh-TW',2),1090:('en',2),1110:('en',3),1130:('zh-TW',3)}
    for n,(lang,scale) in samples.items():
        im=np.array(Image.open(p/f'window/pod-f{n:06d}.png').convert('RGBA'));assert im.shape==(200*scale,480*scale,4)
        for side,area in [('left',im[:,:80*scale]),('right',im[:,-80*scale:])]:
            base=Image.open(w/f'assets/theme/posters/poster-{side}-{scale}x.png').convert('RGBA')
            text=Image.open(w/f'assets/theme/posters/poster-text-{side}.{lang}-{scale}x.png').convert('RGBA')
            assert np.array_equal(area,np.array(Image.alpha_composite(base,text)))
        if n in (1090,1110):
            idx=np.frombuffer((p/f'sources/pod-f{n:06d}.idx').read_bytes(),np.uint8).reshape(200,320)
            pal=np.frombuffer((p/f'sources/pod-f{n:06d}.pal').read_bytes(),np.uint8).reshape(256,3)
            assert np.array_equal(im[:,80*scale:400*scale,:3],pal[idx].repeat(scale,0).repeat(scale,1))
    baseline=read(w/'smoke-baseline-old-now/pass.json');assert baseline['status']=='PASS' and baseline['version']=='v.0.1.1-20261101'
    comparisons={}
    for label,file in [('save','save.log'),('load','load.log')]:
        current=state((smoke/'linux-save'/file).read_text())
        control=state((w/'smoke-baseline-old-now'/file).read_text())
        historical=state((r/'dist-all/v.0.1.1-20261101/smoke/linux-save'/file).read_text())
        assert current==control,label
        comparisons[label]=dict(current=current,old_package_replayed_now=control,historical=historical,historical_equal=current==historical)
    save_new=read(smoke/'linux-save/saved.json')['game'];save_old=read(w/'smoke-baseline-old-now/saved.json')['game']
    save_new={k:v for k,v in save_new.items() if k.startswith('SAVE/')};save_old={k:v for k,v in save_old.items() if k.startswith('SAVE/')}
    historical_save=read(r/'docs/re/033-hd-patch-runtime.json')['runtime']['linux_save']['save_hashes']
    assert save_new==save_old
    # 本次原版輸入含78份歷史SAVE，首次匯入會保留；033的輸入只有47個遊戲檔。
    # 目前新舊包須嚴格同84份原始路徑／SHA。歷史A/B十份另保留DOS大小寫對應。
    assert len(save_new)==84 and len(historical_save)==10
    save_names={}
    for name,want in historical_save.items():
        names=[name] if name in save_new else [k for k in save_new if k.casefold()==name.casefold()]
        assert len(names)==1 and save_new[names[0]]==want,(name,names)
        save_names[name]=names[0]
    capture=(w/'capture/intro/run.log').read_text();old=read(r/'docs/re/031-intro-hd-runtime.json')
    assert state(capture)==['5100',str(old['steps']),old['memory_sha256'],old['indexed_sha256'],old['cpu_sha256'],old['palette_sha256']]
    oldpackages=[]
    for name in ('026-package-runtime.json','033-hd-patch-runtime.json'):
        old=read(r/'docs/re'/name)
        for row in old['files']:
            found=list((r/'dist-all'/old['version']).rglob(row['name']));assert len(found)==1 and sha(found[0])==row['sha256'] and found[0].stat().st_size==row['size'];oldpackages.append(row['sha256'])
    assert len(oldpackages)==12 and tree(r/'workplace/input/dos/pod')==read(r/'docs/re/035-final-runtime.json')['originals_unchanged']
    result=dict(date='2026-11-02',issue=78,status='PASS candidate package scope',version=meta['version'],source=source,files=records,
        theme_files=2034,fonts=18,reconstructed_original_files=1091,original_files_in_patch=0,source_archives_scanned=True,
        runtime=cases,poster_samples=samples,F12_original_pixels=True,normal_save_load_state_unchanged=True,
        old_package_replay=baseline,save_state_comparisons=comparisons,save_files_equal_historical=True,historical_save_name_mapping=save_names,
        existing_save_files_preserved_in_test_input=True,save_files_in_same_input_replays=len(save_new),
        old_twelve_packages_unchanged=oldpackages,originals=original,assets=read(w/'assets/assets-receipt.json'),
        audio_capture_state_equals_original_baseline=True,capture_build=read(w/'capture/capture-build.json'),
        logs={str(f.relative_to(smoke)):sha(f) for f in smoke.glob('*/*.log')},
        limits=['Windows Wine only; native Windows/macOS unverified','macOS universal architecture and dependencies only; unsigned','New music/listening approval remains separate','No full campaign required by user; no completion claim'])
    a.receipt.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print('PASS 12 actual packages, first/restart, five locales, F12, saves, Wine; old12 unchanged')


if __name__=='__main__':main()
