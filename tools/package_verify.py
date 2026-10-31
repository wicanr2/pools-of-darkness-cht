#!/usr/bin/env python3
"""驗證實際候選封包、正常啟動與存讀檔；通過後才複製到dist-all。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil
import subprocess
import tarfile
import tempfile
import zipfile

import numpy as np
from PIL import Image


def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(p):return json.loads(p.read_text())
def tree(p):return {str(f.relative_to(p)):sha(f) for f in sorted(p.rglob('*')) if f.is_file()}


def state(log):
    a=re.search(r'frames=(\d+) steps=(\d+) memory_sha256=(\w+) indexed_sha256=(\w+)',log)
    b=re.search(r'cpu_sha256=(\w+) palette_sha256=(\w+)',log)
    assert a and b
    return dict(frames=int(a[1]),steps=int(a[2]),memory=a[3],indexed=a[4],cpu=b[1],palette=b[2])


def background_transition(capture):
    log=(capture/'run.log').read_text()
    events=re.findall(r'閃現 第 (\d+) 格原樣呈現、第 (\d+) 格被換 列 (\d+) 欄 ([0-9 ]+)',log)
    assert len(events)==1 and events[0][:3]==('2804','2805','24')
    assert [int(x) for x in events[0][3].split()]==list(range(23))
    # F10檔名比已呈現畫格多1：2805及2806對應量尺2804及2805。
    before,after='pod-f002805','pod-f002806'
    idx=(capture/'sources'/(before+'.idx')).read_bytes();pal=(capture/'sources'/(before+'.pal')).read_bytes()
    assert idx==(capture/'sources'/(after+'.idx')).read_bytes() and pal==(capture/'sources'/(after+'.pal')).read_bytes()
    area=np.frombuffer(idx,np.uint8).reshape(200,320)[192:200,:184]
    assert set(area.tobytes())<=set(range(160,176)) and len(set(area.tobytes()))>1
    # 原版這裡只有紅色背景，尚無之後使用的粉紅13、綠10及白15字形。
    rgb=np.frombuffer(pal,np.uint8).reshape(256,3)
    assert all(rgb[c,0]>0 and rgb[c,1]==rgb[c,2]==0 for c in set(area.tobytes()))
    a=np.array(Image.open(capture/'shots'/(before+'.png')).convert('RGBA'))
    b=np.array(Image.open(capture/'shots'/(after+'.png')).convert('RGBA'))
    delta=np.any(a!=b,axis=2);outside=delta.copy();outside[576:600,:552]=False
    assert a.shape==b.shape==(600,960,4) and delta.any() and not outside.any()
    later=(capture/'sources/pod-f002810.idx').read_bytes()[192*320:]
    assert {10,13,15}<=set(later)
    return dict(classification='early translated title appearing over unchanged red credits background; original text not yet drawn',
        raw_count=23,classified_events=23,unexpected_events=0,shown=2804,replaced=2805,row=24,columns=list(range(23)),
        source_indexed_sha256=sha(capture/'sources'/(before+'.idx')),palette_sha256=sha(capture/'sources'/(before+'.pal')),
        before_png_sha256=sha(capture/'shots'/(before+'.png')),after_png_sha256=sha(capture/'shots'/(after+'.png')),
        evidence='same full indices and palette; only title rectangle changes; source uses160..175 red texture without10/13/15 text; native-size visual review')


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    for n in ('root','work','original','receipt'):ap.add_argument('--'+n,type=Path,required=True)
    ap.add_argument('--emit',type=Path);a=ap.parse_args();w=a.work;r=a.root;pack=w/'packages';smoke=w/'smoke';meta=read(pack/'packages.json')
    assert meta['version']=='v.0.1.0-20261101' and meta['engine_commit']=='4a8ef9f4ff18c82d8c98595faf19fb1ee803c609'
    assert len(meta['files'])==9
    originals={p.name:sha(p) for p in a.original.iterdir() if p.is_file()};assert len(originals)==47
    summaries=[]
    for row in meta['files']:
        archive=pack/row['name'];assert sha(archive)==row['sha256'] and archive.stat().st_size==row['size']
        with tempfile.TemporaryDirectory(dir=w,prefix='package-inspect-') as tmp:
            tmp=Path(tmp)
            if archive.suffix=='.AppImage':
                subprocess.run([str(archive),'--appimage-extract'],cwd=tmp,stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,check=True,timeout=60)
                top=tmp/'squashfs-root'
            else:
                with zipfile.ZipFile(archive) as z:
                    assert z.testzip() is None and all(i.flag_bits&0x800 for i in z.infolist() if not i.filename.isascii())
                    z.extractall(tmp)
                    # zipfile不恢復mode，架構檢查仍讀真實內容，執行權限另由ZIP metadata核對。
                    if row['platform']=='macos':assert any(i.filename.endswith('/Contents/MacOS/pod-play') and ((i.external_attr>>16)&0o111) for i in z.infolist())
                top=next(p for p in tmp.iterdir() if p.is_dir())
            m=read(top/'PACKAGE-MANIFEST.json');actual=tree(top);actual.pop('PACKAGE-MANIFEST.json')
            assert actual==m['files'] and sha(top/'PACKAGE-MANIFEST.json')==row['payload_manifest_sha256']
            assert m['version']==row['version']==meta['version'] and m['engine_commit']==meta['engine_commit']
            data=top/'usr/bin' if row['platform']=='linux' else top/'PODCHT.app/Contents/Resources' if row['platform']=='macos' else top
            assert (data/'README.txt').read_bytes().startswith(b'\xef\xbb\xbf')
            assert len(list((data/'font').glob('*.golemfnt')))==6
            if row['kind']=='full-local':assert tree(data/'original')==originals and not(data/'original/SAVE').exists()
            else:
                assert not(data/'original').exists() and not(set(actual.values())&set(originals.values()))
                assert not any(p.suffix.lower() in ('.dax','.bma','.acf','.idx') for p in top.rglob('*') if p.is_file())
                with tarfile.open(data/'source/dosgolem.tar.gz') as t:
                    assert not any(Path(x.name).suffix.lower() in ('.exe','.com','.dax','.bma','.acf','.rom') for x in t.getmembers() if x.isfile())
            if row['edition']=='hd':assert len(tree(data/'hd'))==1525
            else:assert not(data/'hd').exists()
            summaries.append(dict(name=row['name'],kind=row['kind'],platform=row['platform'],edition=row['edition'],sha256=row['sha256'],size=row['size'],manifest_sha256=row['payload_manifest_sha256'],content_verified=True))
    cases={}
    for edition in ('base','hd'):
        p=smoke/('linux-'+edition);log=(p/'run.log').read_text();s=state(log)
        assert s['frames']==3800 and s['steps']==365750000
        demo=re.search(r'中止 (\d+) 次、中止後返回 (\d+) 次',log);assert demo and list(demo.groups())==['1','1'] and '執行中 false' in log
        assert read(p/'restart-pass.json')['game_files_unchanged']
        assert len(list((p/'shots').glob('*.png')))==4
        assert meta['version'] in (p/'version.txt').read_text() and meta['engine_commit'] in (p/'version.txt').read_text()
        game=read(p/'first-import.json');assert set(originals)<=set(game) and all(game[n]==h for n,h in originals.items()) and not any(n.startswith('SAVE/') for n in game)
        draw=read(p/'draw-report.json');assert draw['FallbackBytes']==draw['Mismatches']==draw['LegacyRGBFrames']==0 and draw['HostBytes']==draw['NativeBytes']
        cases[edition]=dict(state=s,imported_files=len(game),demo_abort=1,demo_return=1,restart_without_original=True,game_files=game,
            log_sha256=sha(p/'run.log'),frame_sha256=sha(p/'frame.png'),source_sha256=sha(p/'source.idx'),palette_sha256=sha(p/'source.pal'),
            draw_sha256=sha(p/'draw-report.json'),native_bytes=draw['NativeBytes'],fallback_bytes=0,mismatches=0,legacy_rgb_frames=0)
    assert cases['base']['state']==cases['hd']['state'] and cases['base']['game_files']==cases['hd']['game_files']
    assert (smoke/'linux-base/source.idx').read_bytes()==(smoke/'linux-hd/source.idx').read_bytes()
    assert (smoke/'linux-base/source.pal').read_bytes()==(smoke/'linux-hd/source.pal').read_bytes()
    assert (smoke/'linux-base/original.rgba').read_bytes()==(smoke/'linux-hd/original.rgba').read_bytes()
    assert 'pod-play: 閃現 0 次' in (smoke/'linux-hd/run.log').read_text()
    raw=int(re.search(r'閃現 (\d+) 次',(smoke/'linux-base/run.log').read_text())[1]);assert raw==23
    events=re.findall(r'閃現 第 (\d+) 格原樣呈現、第 (\d+) 格被換 列 (\d+) 欄 ([0-9 ]+)',(smoke/'linux-base/run.log').read_text())
    assert len(events)==1 and events[0][:3]==('2804','2805','24') and [int(x) for x in events[0][3].split()]==list(range(23))
    capture=w/'test/base-flash-packaged'
    flash=background_transition(capture);assert raw==flash['classified_events']
    assert 'pod-play: 閃現 23 次' in (capture/'run.log').read_text()
    missing=read(smoke/'linux-missing/missing-pass.json');assert missing['missing_original_rejected'] and missing['version_hint']=='1.10'
    save=read(smoke/'linux-save/save-load-pass.json');assert save['save_hashes_preserved'] and 'savgamB.pty' in save['files']
    assert state((smoke/'linux-save/save.log').read_text())['frames']==4700 and state((smoke/'linux-save/load.log').read_text())['frames']==3250
    win=read(smoke/'windows/windows-pass.json');assert win['status']=='PASS' and meta['version'] in (smoke/'windows/version.log').read_text()
    assert state((smoke/'windows/play.log').read_text())['frames']==1250
    proof=dict(date='2026-11-01',issue=69,status='PASS within candidate scope; project incomplete',version=meta['version'],engine_commit=meta['engine_commit'],
        files=summaries,linux=dict(cases=cases,same_original_state=True,same_guest_files=True,hd_flash_count=0,base_flash=flash,save_load=save,missing_original=missing),
        windows=win,macos=dict(architectures=['x86_64','arm64'],dependencies_checked=True,execution='not-run; no macOS hardware',signed=False,notarized=False),
        assets=dict(source_git_tree=(w/'main-tree.txt').read_text().strip(),receipt_sha256=sha(w/'assets/assets-receipt.json'),theme_files=1525,fonts=6),
        build_logs={p.name:sha(p) for p in sorted((w/'test').glob('*.log'))},
        build_info={str(p.relative_to(w/'bin')):sha(p) for p in sorted((w/'bin').rglob('*.txt'))},
        limits=['candidate only; no tag or public Release','HD patch not delivered','new music approval, complete multilingual data, invincibility and full animation HD remain pending',
                'Windows tested under Wine only; macOS structure only; no physical Windows/macOS acceptance','base raw flash23 classified by unchanged red background and early translated title; metric unchanged'],
        tools=dict(go='go1.24.13',linux_image='dosgolem-go-ebiten:1.24-r1',mac_image='cok-osxcross:go1.24.13-sdk15.5-r1',
            squashfs_image='kgb-cht-pkg:1.99.1-r2',wine_image='yuan-wine-verify:noble-r2',
            appimage_runtime_sha256='2fca8b443c92510f1483a883f60061ad09b46b978b2631c807cd873a47ec260d'),
        tool_sha256={n:sha(r/'tools'/n) for n in ('package.sh','package_build.sh','package_payload.py','package_appimage.py','package_smoke_linux.sh','package_smoke_windows.py','package_verify.py')})
    a.receipt.write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n')
    if a.emit:
        assert a.emit.name==meta['version'] and a.emit.parent.name=='dist-all' and not a.emit.exists()
        a.emit.mkdir(parents=True);(a.emit/'smoke').mkdir()
        for kind in ('patch','full-local'):(a.emit/kind).mkdir()
        for row in summaries:shutil.copy2(pack/row['name'],a.emit/row['kind']/row['name'])
        shutil.copyfile(a.receipt,a.emit/'smoke/package-runtime.json')
        for case in ('linux-base','linux-hd','linux-save','linux-missing','windows'):
            dest=a.emit/'smoke'/case;dest.mkdir();src=smoke/case
            for pattern in ('*.log','*.json','*.png','shots/*.png','save-shots/*.png','load-shots/*.png'):
                for f in src.glob(pattern):
                    target=dest/f.relative_to(src);target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(f,target)
        flash_dest=a.emit/'smoke/base-flash-source';flash_dest.mkdir()
        for pattern in ('*.log','shots/*.png','sources/*.idx','sources/*.pal'):
            for f in capture.glob(pattern):
                target=flash_dest/f.relative_to(capture);target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(f,target)
        (a.emit/'SHA256SUMS.json').write_text(json.dumps(dict(version=meta['version'],engine_commit=meta['engine_commit'],candidate=True,
            files=[dict(path=x['kind']+'/'+x['name'],size=x['size'],sha256=x['sha256']) for x in summaries],
            validation=dict(path='smoke/package-runtime.json',sha256=sha(a.receipt)),rights='full-local never upload; patch basic only; no public Release'),ensure_ascii=False,indent=2)+'\n')
        (a.emit/'README.md').write_text('# 黑暗之池可試玩候選 '+meta['version']+'\n\n這是本機候選，原案尚未全部完成。full-local含原版資料，請勿上傳或散布；patch只有基本中文化包，HD patch待完成。\n\nLinux與Windows Wine啟動、demo中止與Linux存讀檔已抽驗；macOS只有universal結構與依賴檢查，真機未驗。各包README.txt記錄操作與存檔位置，SHA256SUMS.json及smoke/package-runtime.json提供完整驗收範圍。\n')
        assert all(sha(a.emit/x['kind']/x['name'])==x['sha256'] for x in summaries)
    print(json.dumps(dict(status=proof['status'],packages=9,linux_demo_abort=True,save_B_cold_load=True,windows_wine=True,macos='structure-only',base_raw_flash=23,unexpected_flash=0),indent=2))


if __name__=='__main__':main()
