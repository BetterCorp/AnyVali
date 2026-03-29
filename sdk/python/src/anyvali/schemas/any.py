"""Any schema."""

from __future__ import annotations

from typing import Any

from .base import BaseSchema, ValidationContext


class AnySchema(BaseSchema):
    """Schema that accepts any value."""

    def __init__(self) -> None:
        super().__init__()

    def _accepts_none(self) -> bool:
        return True

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        return input

    def _to_node(self) -> dict[str, Any]:
        return self._add_common_node_fields({"kind": "any"})
