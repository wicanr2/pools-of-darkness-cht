#!/usr/bin/env python3
"""從Git素材與自備原版，重建已驗主題和字型；不讀舊workplace。

日韓候選的字集與素材登錄入口為 tools/jako_runtime_manifest.py，預設只檢查。
"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile
import struct

sys.path.insert(0, str(Path(__file__).resolve().parent))
from dax_vga import block_vga, sprite_vga
from ecl_text_catalog import dax_blocks
from intro_bma import decode as decode_bma, states as bma_states


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def inside(root, name):
    p = root / name
    if not name or Path(name).is_absolute() or root.resolve() not in p.resolve().parents:
        raise ValueError(f'路徑超出根目錄：{name}')
    return p


def rounded_sources(root, row):
    region = 'JP' if row['path'].startswith('pod-ja') else 'TW'
    weight = 'B' if row['cell'] in (22, 24) else 'M'
    source = root/'font/source/rounded'/f'GenSenRounded2{region}-{weight}.otf'
    fallback = root/'font/source/unifont'/('unifont_jp-17.0.05.hex' if region == 'JP' else 'unifont_t-17.0.05.hex')
    return source, fallback


def rebuild(root, original, output):
    metadata = root / 'review/hd/runtime/manifest.json'
    lock = json.loads(metadata.read_text())
    if output.exists():
        raise ValueError(f'輸出已存在，請選新的空路徑：{output}')
    if original.resolve() == output.resolve() or original.resolve() in output.resolve().parents:
        raise ValueError('輸出不可放在原版資料內')
    files = lock['theme']
    names = [row['path'] for row in files]
    if len(names) != len(set(names)):
        raise ValueError('重建清單有重複路徑')
    blocks, sources, planned = {}, {}, {}
    movies = {}
    # 在任何輸出寫入前驗證全部輸入。
    for name, expected in lock['source_files'].items():
        if sha(inside(root, name)) != expected:
            raise ValueError(f'字型來源雜湊不符：{name}')
    for name, expected in lock['original_archives'].items():
        path = inside(original, name)
        if sha(path) != expected:
            raise ValueError(f'原版雜湊不符：{name}')
        if name.endswith('.DAX'):
            blocks[name] = dict(dax_blocks(path.read_bytes()))
        elif name in ('INTRO.BMA', 'FINAL.BMA'):
            movie = {}; movies[name] = movie
            for frame, before, after in bma_states(decode_bma(path.read_bytes())):
                points = [p for span in frame.spans for p in range(span.y*320+span.x,span.y*320+span.x+len(span.pixels)) if p<53760]
                movie[(frame.block,frame.index)] = dict(before=before[:53760],after=after[:53760],writes=struct.pack('<'+'H'*len(points),*points))
        elif name not in ('INTRO.ACF', 'FINAL.ACF'):
            raise ValueError('未知的原版重建輸入')
    for row in files:
        name = row['path']
        inside(output, name)
        if row['kind'] == 'repo-file':
            p = inside(root, row['source'])
            if sha(p) != row['sha256']:
                raise ValueError(f'素材雜湊不符：{row["source"]}')
            planned[name] = p
            continue
        if row['kind'] == 'original-intro':
            archive = row.get('archive') or 'INTRO.BMA'
            if archive not in ('INTRO.BMA', 'FINAL.BMA'):
                raise ValueError('未知BMA來源')
            source = movies[archive][(row['block'],row['frame'])]
            if row['decoder'] == 'bma-index' and row['side'] in ('before','after'):
                data = source[row['side']]
            elif row['decoder'] == 'bma-writes':
                data = source['writes']
            else:
                raise ValueError('未知的BMA重建來源')
            if hashlib.sha256(data).hexdigest() != row['sha256']:
                raise ValueError(f'BMA來源重生失配：{name}')
            planned[name] = data
            continue
        if row['kind'] != 'original-index':
            raise ValueError('未知的重建來源')
        key = (row['archive'], row['block'], row['decoder'])
        if key not in sources:
            data = blocks[key[0]][key[1]]
            if key[2] == 'wall-8x8':
                if len(data) < 434 or (len(data)-434) % 64:
                    raise ValueError('牆面來源長度不合')
                frames = [data[n:n+64] for n in range(434, len(data), 64)]
            elif key[2] in ('sprite-vga', 'block-vga'):
                parser = sprite_vga if key[2] == 'sprite-vga' else block_vga
                # 只重生索引，未定義RGB不參與輸出，不以全零色盤產生圖片。
                frames = parser(data, bytes(768))[2]
            else:
                raise ValueError('未知的索引解碼器')
            sources[key] = frames
        data = sources[key][row['frame']]
        if hashlib.sha256(data).hexdigest() != row['sha256']:
            raise ValueError(f'索引重生失配：{name}')
        planned[name] = data
    output.parent.mkdir(parents=True, exist_ok=True)
    # 候選完整產生後才發布此輪輸出，不碰正式主題。
    with tempfile.TemporaryDirectory(prefix='.pod-rebuild-', dir=output.parent) as tmp:
        stage = Path(tmp)
        theme, fonts = stage/'theme', stage/'font'
        theme.mkdir(); fonts.mkdir()
        for name, data in planned.items():
            p = inside(theme, name); p.parent.mkdir(parents=True, exist_ok=True)
            if isinstance(data, Path):
                shutil.copyfile(data, p)
            else:
                p.write_bytes(data)
        for row in lock['fonts']:
            p = fonts / row['path']
            if not re.fullmatch(r'pod-(?:[a-z0-9-]+|zh-CN(?:-3x|-bar12|-bar22)?)\.golemfnt', row['path']):
                raise ValueError('字型檔名不合')
            if row['kind'] == 'rounded':
                source, fallback = rounded_sources(root, row)
                args = [str(root/'tools/rounded_font.py'), '--otf', str(source), '--cell', str(row['cell']),
                        '--charset-file', str(inside(root,row['charset'])), '--fallback', str(fallback)]
            elif row['kind'] == 'courier':
                args = [str(root/'tools/en_font.py'), '--ttf', str(root/'font/source/courier/CourierPrime-Bold.ttf'), '--cell', str(row['cell'])]
            else:
                raise ValueError('未知字型來源')
            subprocess.run([sys.executable, '-I', *args, '--out', str(p)], check=True)
            if sha(p) != row['sha256']:
                raise ValueError(f'字型重生失配：{p.name}')
        actual = {str(p.relative_to(theme)):sha(p) for p in theme.rglob('*') if p.is_file()}
        if actual != {row['path']:row['sha256'] for row in files}:
            raise ValueError('主題重建集合不同')
        proof = dict(status='PASS', manifest_sha256=sha(metadata), theme_files=len(actual),
                     theme_tree_sha256=hashlib.sha256(json.dumps(actual,sort_keys=True).encode()).hexdigest(),
                     fonts={p.name:sha(p) for p in sorted(fonts.iterdir())},
                     original_archives=lock['original_archives'], engine=lock['engine'],
                     input_scope='Git files and explicitly supplied original DAX/BMA/ACF; no old workplace',
                     rights='private review inputs; original indexes reconstructed locally')
        (stage/'assets-receipt.json').write_text(json.dumps(proof,ensure_ascii=False,indent=2)+'\n')
        stage.rename(output)
    return proof


def main():
    ap=argparse.ArgumentParser(description=__doc__)
    ap.add_argument('--root',type=Path,required=True)
    ap.add_argument('--original',type=Path,required=True)
    ap.add_argument('--out',type=Path,required=True)
    a=ap.parse_args()
    try:
        print(json.dumps(rebuild(a.root,a.original,a.out),ensure_ascii=False,indent=2))
    except (ValueError,KeyError,IndexError,OSError) as e:
        ap.exit(1,str(e)+'\n')


if __name__=='__main__':
    main()
