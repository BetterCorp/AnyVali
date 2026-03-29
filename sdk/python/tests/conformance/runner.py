"""Corpus loader for AnyVali conformance tests.

Reads JSON test files from spec/corpus/ and yields test cases
for parametrized pytest execution.
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any


# Resolve the corpus directory relative to this file
_CORPUS_DIR = Path(__file__).resolve().parent.parent.parent.parent.parent / "spec" / "corpus"


def _find_test_files() -> list[Path]:
    """Find all JSON test files in the corpus directory."""
    if not _CORPUS_DIR.exists():
        return []
    return sorted(_CORPUS_DIR.rglob("*.json"))


def load_corpus() -> list[dict[str, Any]]:
    """Load all test cases from the corpus.

    Each JSON file is expected to contain either:
    - A single test case dict with keys: schema, input, expected
    - A list of test case dicts

    Each test case has:
    - "description": human-readable description
    - "schema": AnyVali document dict
    - "input": the input value to parse
    - "expected": dict with:
        - "success": bool
        - "data": expected parsed output (if success)
        - "issues": list of expected issues (if failure), each with at least "code"
    """
    cases: list[dict[str, Any]] = []
    for path in _find_test_files():
        try:
            with open(path, "r", encoding="utf-8") as f:
                data = json.load(f)
        except (json.JSONDecodeError, OSError):
            continue

        rel_path = path.relative_to(_CORPUS_DIR)

        if isinstance(data, list):
            for i, case in enumerate(data):
                case.setdefault("description", f"{rel_path}[{i}]")
                case["_source"] = str(rel_path)
                cases.append(case)
        elif isinstance(data, dict):
            if "cases" in data:
                for i, case in enumerate(data["cases"]):
                    case.setdefault("description", f"{rel_path}[{i}]")
                    case["_source"] = str(rel_path)
                    cases.append(case)
            else:
                data.setdefault("description", str(rel_path))
                data["_source"] = str(rel_path)
                cases.append(data)

    return cases


def corpus_ids(cases: list[dict[str, Any]]) -> list[str]:
    """Generate test IDs from test case descriptions."""
    return [case.get("description", f"case-{i}") for i, case in enumerate(cases)]
