"""Export schemas to AnyVali interchange format."""

from __future__ import annotations

import copy
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

    def add_definition(name: str, node: Any) -> None:
        if name in defs and json.dumps(defs[name], sort_keys=True) != json.dumps(node, sort_keys=True):
            raise ValueError(f"Conflicting definition: {name}")
        defs[name] = node

    pending = [schema, *(definitions or {}).values()]
    seen: set[int] = set()
    while pending:
        current = pending.pop()
        if id(current) in seen:
            continue
        seen.add(id(current))
        for name, node in current._imported_definitions.items():
            add_definition(name, node)
        pending.extend(current._children())
    if definitions:
        for name, defn_schema in definitions.items():
            add_definition(name, defn_schema._to_node())
    defs = copy.deepcopy(defs)

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
