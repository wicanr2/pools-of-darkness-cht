#!/usr/bin/env python3
"""獨立解開三平台HD patch，驗實際匯入、正常路徑與權利邊界。"""
import argparse
import hashlib
import json
from pathlib import Path
import plistlib
import re
import shutil
import struct
import subprocess
import tarfile
import tempfile

import numpy as np
from PIL import Image


def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p): return json.loads(p.read_text())
def tree(p): return {str(f.relative_to(p)):sha(f) for f in sorted(p.rglob('*')) if f.is_file()}
def rgba(p): return np.array(Image.open(p).convert('RGBA'))
def state(log):
    a=re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',log)
    b=re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)',log)
    assert a and b
    return list(a.groups())+list(b.groups())


def archives(r,w):
    import zipfile
    meta=read(w/'packages/packages.json');lock=read(w/'main/review/hd/runtime/manifest.json')
    assert meta['version']=='v.0.1.1-20261101' and len(meta['files'])==3
    assert meta['engine_commit']==lock['engine']['commit']
    original=tree(w/'smoke-source/original')
    assert len(original)==47
    assert tree(r/'workplace/input/dos/pod')==read(r/'docs/re/031-intro-hd-runtime.json')['original_inputs']
    font={x['path']:x['sha256'] for x in lock['fonts']}
    sources={x['source']:x['sha256'] for x in lock['theme'] if x['kind']=='repo-file'}
    assert len(lock['theme'])==1866 and len(font)==10
    assert sum(x['kind']!='repo-file' for x in lock['theme'])==983
    summaries=[]
    for row in meta['files']:
        arc=w/'packages'/row['name'];assert sha(arc)==row['sha256'] and arc.stat().st_size==row['size']
        with tempfile.TemporaryDirectory(dir=w,prefix='inspect-hd-') as td:
            tmp=Path(td)
            if row['platform']=='linux':
                subprocess.run([str(arc),'--appimage-extract'],cwd=tmp,stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,check=True,timeout=90)
                top=tmp/'squashfs-root';data=top/'usr/bin';exe=data/'pod-play'
                b=exe.read_bytes();assert b[:4]==b'\x7fELF' and struct.unpack_from('<H',b,18)[0]==62
            else:
                with zipfile.ZipFile(arc) as z:
                    assert z.testzip() is None
                    assert all(i.flag_bits&0x800 for i in z.infolist() if not i.filename.isascii())
                    if row['platform']=='macos':assert any(i.filename.endswith('/Contents/MacOS/pod-play') and (i.external_attr>>16)&0o111 for i in z.infolist())
                    z.extractall(tmp)
                top=next(tmp.iterdir());data=top if row['platform']=='windows' else top/'PODCHT.app/Contents/Resources'
                exe=top/'PODCHT.exe' if row['platform']=='windows' else top/'PODCHT.app/Contents/MacOS/pod-play'
                b=exe.read_bytes()
                if row['platform']=='windows':
                    off=struct.unpack_from('<I',b,0x3c)[0]
                    assert b[:2]==b'MZ' and b[off:off+4]==b'PE\0\0' and struct.unpack_from('<H',b,off+4)[0]==0x8664
                else:
                    magic,n=struct.unpack_from('>II',b);assert magic in (0xcafebabe,0xcafebabf) and n==2
                    stride=20 if magic==0xcafebabe else 32
                    assert {struct.unpack_from('>I',b,8+i*stride)[0] for i in range(n)}=={0x1000007,0x100000c}
                    plist=plistlib.loads((top/'PODCHT.app/Contents/Info.plist').read_bytes())
                    assert plist['CFBundleVersion']==meta['version'] and plist['CFBundleExecutable']=='pod-play'
            m=read(top/'PACKAGE-MANIFEST.json');actual=tree(top);actual.pop('PACKAGE-MANIFEST.json')
            assert actual==m['files'] and sha(top/'PACKAGE-MANIFEST.json')==row['payload_manifest_sha256']
            for key in ('version','engine_commit','kind','edition','platform'):assert m[key]==row[key]
            assert m['kind']=='patch' and m['edition']=='hd' and m['original_files']==0
            assert m['fonts']==10 and m['theme_files']==1866
            assert not (data/'original').exists() and not (data/'hd').exists()
            assert not (set(actual.values())&set(original.values()))
            assert not any(Path(p).suffix.lower() in ('.dax','.bma','.acf','.idx','.ops','.rom') for p in actual)
            assert (data/'README.txt').read_bytes().startswith(b'\xef\xbb\xbf')
            assert tree(data/'font')==font
            assert sha(data/'hd-source/manifest.json')==sha(w/'main/review/hd/runtime/manifest.json')
            expected=dict(sources);expected['manifest.json']=sha(w/'main/review/hd/runtime/manifest.json')
            assert tree(data/'hd-source')==expected
            # 程式原始碼巢狀封包也逐筆掃描，不把外層副檔名當權利證明。
            with tarfile.open(data/'source/dosgolem.tar.gz') as t:
                for member in t.getmembers():
                    assert not member.issym() and not member.islnk()
                    if member.isfile():
                        assert Path(member.name).suffix.lower() not in ('.exe','.com','.dax','.bma','.acf','.idx','.ops','.rom')
                        assert hashlib.sha256(t.extractfile(member).read()).hexdigest() not in original.values()
            for f in ('LICENSE','THIRD-PARTY.md','licenses/COPYING.LGPL','licenses/OFL-GenSenRounded.txt','licenses/Go-LICENSE','source/nukedopl/COPYING.LGPL'):
                assert (data/f).is_file(),f
            assert (data/'text/translit-zh-CN').is_dir()
            for x in lock['theme']:
                if x['path'].startswith('posters/'):
                    assert x['kind']=='repo-file' and 'original-reference' not in x.get('rights','')
            summaries.append({k:row[k] for k in ('name','size','sha256','payload_manifest_sha256','platform')})
    libs=(w/'bin/macos/libraries.txt').read_text()
    assert 'architecture x86_64' in libs and 'architecture arm64' in libs
    assert all(line.strip().startswith(('/usr/lib/','/System/Library/')) for line in libs.splitlines() if line.startswith('\t'))
    old=read(r/'docs/re/026-package-runtime.json')
    for x in old['files']:
        assert sha(r/'dist-all/v.0.1.0-20261101'/x['kind']/x['name'])==x['sha256']
    assert tree(w/'smoke-source/original')==original
    return dict(files=summaries,source_commit=(w/'main-commit.txt').read_text().strip(),engine_commit=meta['engine_commit'],
        theme_files=1866,fonts=10,reconstructed_original_files=983,original_files_in_patch=0,
        originals=original,old_nine_packages_unchanged=True,source_archives_scanned=True)


def runtime(r,w):
    smoke=w/'smoke';p=smoke/'linux-intro';lock=read(w/'main/review/hd/runtime/manifest.json')
    expected={x['path']:x['sha256'] for x in lock['theme']}
    a=read(p/'first.json');b=read(p/'restart.json')
    assert a==b and a['cache']==expected and len(a['cache'])==1866
    assert state((p/'run.log').read_text())==state((p/'restart.log').read_text())
    assert state((p/'run.log').read_text())[0]=='5100'
    assert tree(p/'shots')==tree(p/'restart-shots') and tree(p/'window')==tree(p/'restart-window')
    assert tree(p/'sources')==tree(p/'restart-sources')
    assert len(tree(p/'window'))==22
    assert np.array_equal(rgba(p/'window/pod-f001979.png'),rgba(r/'review/hd/b47-hd-posters/approved-window.png'))
    # F12沿既有英文原版契約，也將兩側文字切為英文。
    samples={700:('zh-TW',3),702:('en',3),704:('en',2),706:('zh-CN',2),708:('en',2),710:('en',3),712:('zh-CN',3),1979:('zh-TW',3)}
    for frame,(lang,scale) in samples.items():
        im=rgba(p/f'window/pod-f{frame:06d}.png');side=80*scale
        assert im.shape==(200*scale,480*scale,4)
        for name,area in [('left',im[:,:side]),('right',im[:,-side:])]:
            preview=rgba(r/f'review/hd/b47-hd-posters/poster-full-{name}.{lang}-{scale}x.png')
            assert np.array_equal(area,preview),(frame,name)
    # F12中央是該格原版色盤／索引的最近鄰整數放大。
    for frame,scale in ((708,2),(710,3)):
        idx=np.frombuffer((p/f'sources/pod-f{frame:06d}.idx').read_bytes(),np.uint8).reshape(200,320)
        pal=np.frombuffer((p/f'sources/pod-f{frame:06d}.pal').read_bytes(),np.uint8).reshape(256,3)
        original=np.repeat(np.repeat(pal[idx],scale,axis=0),scale,axis=1)
        assert np.array_equal(rgba(p/f'window/pod-f{frame:06d}.png')[:,80*scale:400*scale,:3],original)
    log=(p/'run.log').read_text();raw=int(re.search(r'閃現 (\d+) 次',log)[1]);counts={'explicit_switch':0,'black_transition':0}
    for first,last,row,cols in re.findall(r'閃現 第 (\d+) 格原樣呈現、第 (\d+) 格被換 列 (\d+) 欄 ([0-9 ]+)',log):
        if (first,last)==('706','711'):counts['explicit_switch']+=len(cols.split())
        elif (first,last)==('3767','3769'):counts['black_transition']+=len(cols.split())
        else:raise AssertionError((first,last,row,cols))
    assert raw==sum(counts.values())==599 and counts==dict(explicit_switch=559,black_transition=40)
    black=rgba(p/'shots/pod-f003768.png');assert not black[:,:,:3].any()
    idx=np.frombuffer((p/'sources/pod-f003768.idx').read_bytes(),np.uint8)
    pal=np.frombuffer((p/'sources/pod-f003768.pal').read_bytes(),np.uint8).reshape(256,3)
    assert not pal[idx].any()
    demo=(smoke/'linux-demo/run.log').read_text()
    assert state(demo)[0]=='3800' and 'pod-play: 閃現 0 次' in demo
    assert re.search(r'中止 1 次、中止後返回 1 次',demo) and '執行中 false' in demo
    assert read(smoke/'linux-demo/first.json')['cache']==expected
    save=read(smoke/'linux-save/saved.json');load=read(smoke/'linux-save/loaded.json')
    saved={k:v for k,v in save['game'].items() if k.startswith('SAVE/')}
    loaded={k:v for k,v in load['game'].items() if k.startswith('SAVE/')}
    assert saved==loaded and 'SAVE/savgamB.pty' in saved
    assert save['cache']==load['cache']==expected and save['mtime']==load['mtime']
    assert state((smoke/'linux-save/save.log').read_text())[0]=='4700'
    assert state((smoke/'linux-save/load.log').read_text())[0]=='3250'
    missing=read(smoke/'linux-missing/pass.json');assert missing['missing_original_rejected']
    corrupt=read(smoke/'linux-corrupt/pass.json');assert all(corrupt.values())
    before=read(smoke/'linux-corrupt/before.json');after=read(smoke/'linux-corrupt/restored.json')
    assert before['cache']==after['cache']==expected and before['game']==after['game']
    win=read(smoke/'windows/windows-pass.json');assert win['status']=='PASS' and win['cache']==expected
    assert win['cache_reused_without_write'] and win['package_unchanged']
    assert state((smoke/'windows/play.log').read_text())[0]=='1250'
    rebuild=read(smoke/'rebuild/receipt.json');assert rebuild['status']=='PASS'
    return dict(linux_intro=dict(first_restart_same=True,cache_reused_without_write=True,frames=5100,shots=22,
                approved_B_window_pixel_exact=True,poster_locales_and_scales=True,F12_original_pixels=True,
                raw_flash=raw,classified=counts,unexpected_visible_flash=0,state=state(log)),
        linux_demo=dict(frames=3800,abort=1,returned=1,flash=0),linux_save=dict(saved_B=True,cold_loaded_B=True,save_hashes=saved),
        missing=missing,corrupt=corrupt,windows=dict(scope=win['scope'],frames=1250,cache_reused=True,files=1866),
        macos=dict(architectures=['x86_64','arm64'],dependencies_checked=True,native_tested=False,signed=False,notarized=False),
        clean_rebuild=rebuild,logs={str(f.relative_to(smoke)):sha(f) for f in smoke.glob('*/*.log')})


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for n in ('root','work','receipt'):ap.add_argument('--'+n,type=Path,required=True)
    ap.add_argument('--packages-only',action='store_true');ap.add_argument('--emit',type=Path)
    a=ap.parse_args();proof=archives(a.root,a.work)
    if not a.packages_only:proof['runtime']=runtime(a.root,a.work)
    proof.update(date='2026-11-01',issue=75,version='v.0.1.1-20261101',status='PASS within HD patch candidate scope',
        limits=['original project incomplete: Japanese/Korean, music/listening approval, FINAL HD and promo pending',
                'Windows Wine only; macOS structure/dependencies only; no native acceptance',
                'no tag or public Release; source originals required separately'],
        tools={f:sha(a.root/'tools'/f) for f in ('hd-patch.sh','hd_patch_smoke_linux.sh','hd_patch_smoke_windows.py','hd_patch_verify.py')})
    a.receipt.write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n')
    if a.emit:
        assert not a.packages_only and a.emit.name==proof['version'] and a.emit.parent.name=='dist-all' and not a.emit.exists()
        (a.emit/'patch').mkdir(parents=True);(a.emit/'smoke').mkdir()
        for row in proof['files']:
            shutil.copy2(a.work/'packages'/row['name'],a.emit/'patch'/row['name'])
            assert sha(a.emit/'patch'/row['name'])==row['sha256']
        shutil.copy2(a.receipt,a.emit/'smoke/hd-patch-runtime.json')
        for case in ('linux-intro','linux-demo','linux-save','linux-missing','linux-corrupt','windows','rebuild'):
            dest=a.emit/'smoke'/case;src=a.work/'smoke'/case;dest.mkdir()
            for pattern in ('*.log','*.png','pass.json','receipt.json','window/*.png','save-shots/*.png','load-shots/*.png','shots/*.png'):
                for f in src.glob(pattern):
                    target=dest/f.relative_to(src);target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(f,target)
        (a.emit/'SHA256SUMS.json').write_text(json.dumps(dict(version=proof['version'],engine_commit=proof['engine_commit'],candidate=True,
            files=[dict(path='patch/'+x['name'],size=x['size'],sha256=x['sha256']) for x in proof['files']],
            validation=dict(path='smoke/hd-patch-runtime.json',sha256=sha(a.receipt))),ensure_ascii=False,indent=2)+'\n')
    print(json.dumps({k:proof[k] for k in ('status','version','engine_commit')},ensure_ascii=False))


if __name__=='__main__':main()
