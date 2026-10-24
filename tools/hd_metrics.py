"""HD 評量公式，沿用銀色匕首之謎 hd_compare.py；來源 SHA-256 b6fce5f8260761e9ebe78138f31e66369fcd3a8d0fb04b34eacf4eb463587f4b。
本作以 VGA RGB 來源驗證，無 EGA 色號假設。見 docs/re/014-vga-hd.md。"""
import struct
import zlib
from pathlib import Path
import numpy as np

def read_png(p):
    """以 PNG IHDR 保留原色彩型態，Pillow 解碼成 RGBA；避免逐點 Python 解濾波。"""
    from PIL import Image
    d = Path(p).read_bytes()
    if d[:8] != b'\x89PNG\r\n\x1a\n' or d[12:16] != b'IHDR':
        raise ValueError(f'{p}：不是 PNG')
    w, h, bd, ct, _, _, il = struct.unpack('>IIBBBBB', d[16:29])
    if bd != 8 or il or ct not in (0, 2, 3, 4, 6):
        raise ValueError(f'{p}：不支援 bit depth {bd}、color type {ct} 或交錯 {il}')
    with Image.open(p) as im:
        px = np.array(im.convert('RGBA'), dtype=np.uint8)
    if px.shape != (h, w, 4):
        raise ValueError(f'{p}：PNG 尺寸不合')
    return ct, px


def write_png(p, rgb):
    h, w, _ = rgb.shape
    raw = b''.join(b'\x00' + rgb[y].tobytes() for y in range(h))

    def chunk(t, d):
        return struct.pack('>I', len(d)) + t + d + struct.pack('>I', zlib.crc32(t + d) & 0xffffffff)
    Path(p).write_bytes(b'\x89PNG\r\n\x1a\n' + chunk(b'IHDR', struct.pack('>IIBBBBB', w, h, 8, 2, 0, 0, 0))
                        + chunk(b'IDAT', zlib.compress(raw, 6)) + chunk(b'IEND', b''))


def edges(lum):
    gx = np.abs(np.diff(lum, axis=1))[:-1, :]
    gy = np.abs(np.diff(lum, axis=0))[:, :-1]
    g = gx + gy
    return g > np.percentile(g, 85)


def align_score(orig_rgb, hd_rgb):
    """HD 以區塊平均縮回原版尺寸，比較強邊緣重合率；另試 ±2 像素平移，回 (原位分數, 最佳分數, 最佳平移)。"""
    h, w, _ = orig_rgb.shape
    H, W, _ = hd_rgb.shape
    s = H // h
    small = hd_rgb[:h * s, :w * s, :3].reshape(h, s, w, s, 3).mean(axis=(1, 3))
    lum = lambda a: a[..., 0] * .299 + a[..., 1] * .587 + a[..., 2] * .114  # noqa: E731
    eo, eh = edges(lum(orig_rgb.astype(float))), edges(lum(small))
    best = (0, (0, 0))
    base = 0
    for dy in range(-2, 3):
        for dx in range(-2, 3):
            sh = np.roll(np.roll(eh, dy, 0), dx, 1)
            sc = (sh & eo).sum() / max(1, eo.sum())
            if (dy, dx) == (0, 0):
                base = sc
            if sc > best[0]:
                best = (sc, (dx, dy))
    return base, best[0], best[1]


def hue_score(orig_rgb, hd_rgb):
    """色相對照：HD 以區塊平均縮回 1 倍，在原版與 HD 都有彩度的像素比色相差（度）。回 (中位數, 超過 30 度的比例)。"""
    h, w, _ = orig_rgb.shape
    s = hd_rgb.shape[0] // h
    small = hd_rgb[:h * s, :w * s, :3].reshape(h, s, w, s, 3).mean(axis=(1, 3))

    def hs(a):
        a = a.astype(float) / 255
        mx, mn = a.max(axis=2), a.min(axis=2)
        d = mx - mn
        sat = np.where(mx > 0, d / np.maximum(mx, 1e-9), 0)
        r, g, b = a[..., 0], a[..., 1], a[..., 2]
        dd = np.maximum(d, 1e-9)
        hue = np.where(mx == r, (g - b) / dd % 6, np.where(mx == g, (b - r) / dd + 2, (r - g) / dd + 4)) * 60
        return hue, sat
    ho, so = hs(orig_rgb)
    hh, sh = hs(small)
    m = (so > 0.3) & (sh > 0.15)
    if not m.any():
        return 0.0, 0.0
    dh = np.abs(ho - hh)[m]
    dh = np.minimum(dh, 360 - dh)
    return float(np.median(dh)), float((dh > 30).mean())


def lowpass(rgb):
    """規格 010 §20.2：對稱 [1,2,1]/4 分離卷積，抑制一像素網點。"""
    rgb = rgb.astype(float)
    for axis in (0, 1):
        padded = np.pad(rgb, [(1, 1) if k == axis else (0, 0) for k in range(3)], mode='edge')
        if axis == 0:
            rgb = (padded[:-2] + 2 * padded[1:-1] + padded[2:]) / 4
        else:
            rgb = (padded[:, :-2] + 2 * padded[:, 1:-1] + padded[:, 2:]) / 4
    return rgb


def structural_score(orig_rgb, hd_rgb):
    """原始邊緣指標另存；雙側去網點後比較同位置輪廓，不搜尋最佳平移。"""
    h, w = orig_rgb.shape[:2]
    sy, sx = hd_rgb.shape[0] // h, hd_rgb.shape[1] // w
    small = hd_rgb[..., :3].reshape(h, sy, w, sx, 3).mean(axis=(1, 3))
    filtered = np.repeat(np.repeat(lowpass(small), sy, 0), sx, 1)
    return float(align_score(lowpass(orig_rgb), filtered)[0])
