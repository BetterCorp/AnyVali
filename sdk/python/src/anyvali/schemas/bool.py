"""Bool schema."""

from __future__ import annotations

from typing import Any

from ..issue_codes import INVALID_TYPE
from .base import BaseSchema, ValidationContext, _anyvali_type_name


class BoolSchema(BaseSchema[bool]):
    """Schema for boolean values."""

    def __init__(self) -> None:
        super().__init__()

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        if not isinstance(input, bool):
            received = _anyvali_type_name(input)
            ctx.add_issue(INVALID_TYPE, f"Expected boolean, received {received}", expected="boolean", received=received)
            return None
        return input

    def _to_node(self) -> dict[str, Any]:
        return self._add_common_node_fields({"kind": "bool"})
