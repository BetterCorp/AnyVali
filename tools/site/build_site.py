from __future__ import annotations

import json
import shutil
from pathlib import Path

from PIL import Image


ROOT = Path(__file__).resolve().parents[2]
SOURCE_DIR = ROOT / "site"
OUTPUT_DIR = ROOT / "dist-site"
LOGO_SOURCE = ROOT / "logo.png"


def build_favicons(destination: Path) -> None:
    with Image.open(LOGO_SOURCE) as image:
        image = image.convert("RGBA")

        sizes = {
            "favicon-16x16.png": (16, 16),
            "favicon-32x32.png": (32, 32),
            "apple-touch-icon.png": (180, 180),
            "android-chrome-192x192.png": (192, 192),
            "android-chrome-512x512.png": (512, 512),
        }

        for filename, size in sizes.items():
            resized = image.resize(size, Image.Resampling.LANCZOS)
            resized.save(destination / filename)

        image.resize((64, 64), Image.Resampling.LANCZOS).save(
            destination / "favicon.ico",
            sizes=[(16, 16), (32, 32), (48, 48), (64, 64)],
        )

    manifest = {
        "name": "AnyVali",
        "short_name": "AnyVali",
        "icons": [
            {
                "src": "/android-chrome-192x192.png",
                "sizes": "192x192",
                "type": "image/png",
            },
            {
                "src": "/android-chrome-512x512.png",
                "sizes": "512x512",
                "type": "image/png",
            },
        ],
        "theme_color": "#0d1220",
        "background_color": "#0d1220",
        "display": "standalone",
    }
    (destination / "site.webmanifest").write_text(
        json.dumps(manifest, indent=2),
        encoding="utf-8",
    )


def main() -> None:
    shutil.rmtree(OUTPUT_DIR, ignore_errors=True)
    shutil.copytree(SOURCE_DIR, OUTPUT_DIR)
    build_favicons(OUTPUT_DIR)


if __name__ == "__main__":
    main()
