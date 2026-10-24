#!/usr/bin/env python3
"""Check source-only subtitle transitions; retain hashes, not original art."""
import argparse
import hashlib
import json
from pathlib import Path

CUES = {
    "a3fcd07235724640d430c73838507b31a9d63fcb9adeed935872fce43ff2786f": "intro.001",
    "2c56a35a0aab3a9cd886167bfd028cabebf267cb736b3cb4aa0867d76feeac47": "intro.002",
    "513dabb86ca006cc28f6b5295769620e5b726e52d48fd60fd29266a6a495c073": "intro.003",
    "560a5b065c8e8e6bc23c2dee218cd998a21f5cf6539d224a1a8030f137568796": "intro.004",
    "12c51da1e763a7c154c9d1fda9e81850bc7eabdec8cbd1c92d48275f0d63014d": "intro.005",
    "03581843fa3b3f9b2f97040d301eaeb848d69096ac80f3f7386cb099b5231c1e": "intro.006",
    "760787555c1643f5ccd990367d6fc69f2d642a10a3aac75075cc5fba709da591": "intro.007",
    "2523926e02c3c538342953d63bbebe149a2494173f09441e4b80d8af146e6162": "intro.008",
}


def sha(data):
    return hashlib.sha256(data).hexdigest()


def check(root):
    j = json.loads((root / "timing.json").read_text())
    assert j["Budget"] == 96250 and j["Frames"] == 5100
    items = {}
    events = j["Events"]
    for i, event in enumerate(events):
        p = root / (event["File"] + ".idx")
        idx, pal = p.read_bytes(), p.with_suffix(".pal").read_bytes()
        assert len(idx) == 64000 and len(pal) == 768
        crop = idx[168 * 320:]
        digest = sha(crop)
        key = CUES.get(digest)
        if not key:
            continue
        bg = 176 if key == "intro.004" else 32
        assert set(crop) == {bg, 7}
        assert pal[bg*3:bg*3+3] == bytes(3)
        fg = list(pal[21:24])
        assert fg[0] == fg[1] == fg[2] and fg[0] <= 170
        last = events[i+1]["Update"] - 1 if i+1 < len(events) else j["Frames"]
        entry = items.setdefault(key, {"key": key, "crop_indexed_sha256": digest, "background_index": bg, "states": []})
        entry["states"].append({"first_update": event["Update"], "last_update": last,
                                "first_step_before_instruction": event["Steps"], "foreground": fg,
                                "indexed_sha256": sha(idx), "palette_sha256": sha(pal)})
    assert len(items) == 8
    for item in items.values():
        visible = [v for v in item["states"] if v["foreground"][0] > 0]
        assert visible
        item["first_visible_update"] = visible[0]["first_update"]
        item["last_visible_update"] = visible[-1]["last_update"]
    digest = {k: bytes(v).hex() if isinstance(v, list) else v for k,v in j["Digest"].items()}
    return {"date": "2026-10-25", "status": "source-timing-confirmed-not-runtime",
            "scope": "normal SND.COM/CONTROL chain, no key, 5100 session updates; same clock and update budget as pod-play",
            "canvas": [320,200], "crop": [0,168,320,32], "clock_percent": 50,
            "frames": j["Frames"], "budget": j["Budget"], "timing_sha256": sha((root/"timing.json").read_bytes()),
            "inputs": json.loads((root/"inputs.json").read_text()), "digest": digest,
            "items": list(items.values()),
            "limits": ["Only observed host-presented frames, not each CPU video write", "No FINAL or movie HD claim", "Grayscale fade between observed DAC values is a rendering contract, not an exact timing oracle"]}


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--capture", type=Path, required=True)
    parser.add_argument("--proof", type=Path, required=True)
    args = parser.parse_args()
    proof = check(args.capture)
    args.proof.write_text(json.dumps(proof,ensure_ascii=False,indent=2)+"\n")
    print("八組字幕來源、淡出與清除通過；正式接入另驗。")
