"""Export schemas to AnyVali interchange format."""

from __future__ import annotations

import json
from typing import Any

from ..schemas.base import BaseSchema
from ..types import AnyValiDocument, ExportMode


def export_schema(
    schema: BaseSchema,
    *,
    mode: ExportMode = "portable",
    definitions: dict[str, BaseSchema] | None = None,
    extensions: dict[str, Any] | None = None,
) -> dict[str, Any]:
    """Export a schema to an AnyVali document dict."""
    root_node = schema._to_node()

    defs: dict[str, Any] = {}
    if definitions:
        for name, defn_schema in definitions.items():
            defs[name] = defn_schema._to_node()

    ext: dict[str, Any] = {}
    if mode == "extended" and extensions:
        ext = dict(extensions)

    doc = AnyValiDocument(
        root=root_node,
        definitions=defs,
        extensions=ext,
    )
    return doc.to_dict()


def export_schema_json(
    schema: BaseSchema,
    *,
    mode: ExportMode = "portable",
    definitions: dict[str, BaseSchema] | None = None,
    extensions: dict[str, Any] | None = None,
    indent: int | None = 2,
) -> str:
    """Export a schema to a JSON string."""
    doc = export_schema(
        schema, mode=mode, definitions=definitions, extensions=extensions
    )
    return json.dumps(doc, indent=indent, ensure_ascii=False)
