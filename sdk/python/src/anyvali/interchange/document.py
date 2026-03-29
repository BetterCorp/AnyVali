"""Document helpers for AnyVali interchange format."""

from __future__ import annotations

import json
from typing import Any

from ..types import AnyValiDocument


def create_document(
    root: dict[str, Any],
    *,
    definitions: dict[str, Any] | None = None,
    extensions: dict[str, Any] | None = None,
) -> AnyValiDocument:
    """Create a new AnyValiDocument."""
    return AnyValiDocument(
        root=root,
        definitions=definitions or {},
        extensions=extensions or {},
    )


def document_to_json(doc: AnyValiDocument, *, indent: int | None = 2) -> str:
    """Serialize an AnyValiDocument to JSON."""
    return json.dumps(doc.to_dict(), indent=indent, ensure_ascii=False)


def document_from_json(source: str) -> AnyValiDocument:
    """Deserialize a JSON string to an AnyValiDocument."""
    data = json.loads(source)
    return AnyValiDocument.from_dict(data)
