from __future__ import annotations

import shutil
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
OUTPUT_DIR = ROOT / ".docs-build"

SOURCE_FILES = [
    Path("README.md"),
    Path("CONTRIBUTING.md"),
    Path("LICENSE"),
    Path("docs/overview.md"),
    Path("docs/numeric-semantics.md"),
    Path("docs/portability-guide.md"),
    Path("docs/sdk-authors-guide.md"),
    Path("spec/spec.md"),
    Path("spec/json-format.md"),
    Path("spec/corpus/README.md"),
]


def rewrite_home_links(text: str) -> str:
    return (
        text.replace("(docs/overview.md)", "(docs/overview.md)")
        .replace("(docs/numeric-semantics.md)", "(docs/numeric-semantics.md)")
        .replace("(docs/portability-guide.md)", "(docs/portability-guide.md)")
        .replace("(docs/sdk-authors-guide.md)", "(docs/sdk-authors-guide.md)")
        .replace("(spec/spec.md)", "(spec/spec.md)")
        .replace("(spec/json-format.md)", "(spec/json-format.md)")
        .replace("(spec/corpus/README.md)", "(spec/corpus/README.md)")
    )


def destination_for(source: Path) -> Path:
    if source == Path("README.md"):
        return OUTPUT_DIR / "index.md"
    return OUTPUT_DIR / source


def main() -> None:
    shutil.rmtree(OUTPUT_DIR, ignore_errors=True)
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    for relative_source in SOURCE_FILES:
        source_path = ROOT / relative_source
        content = source_path.read_text(encoding="utf-8")
        if relative_source == Path("README.md"):
            content = rewrite_home_links(content)

        destination_path = destination_for(relative_source)
        destination_path.parent.mkdir(parents=True, exist_ok=True)
        destination_path.write_text(content, encoding="utf-8")


if __name__ == "__main__":
    main()
