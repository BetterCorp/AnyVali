"""Tuple schema."""

from __future__ import annotations

from typing import Any

from ..issue_codes import INVALID_TYPE, TOO_LARGE, TOO_SMALL
from .base import BaseSchema, ValidationContext, _anyvali_type_name


class TupleSchema(BaseSchema[list[Any]]):
    """Schema for fixed-length tuples with positional element schemas."""

    def __init__(self, items: list[BaseSchema]) -> None:
        super().__init__()
        self._items = list(items)

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        if not isinstance(input, (list, tuple)):
            received = _anyvali_type_name(input)
            ctx.add_issue(INVALID_TYPE, f"Expected tuple, received {received}", expected="tuple", received=received)
            return None

        if len(input) != len(self._items):
            code = TOO_SMALL if len(input) < len(self._items) else TOO_LARGE
            ctx.add_issue(code, f"Expected exactly {len(self._items)} items, received {len(input)}", expected=len(self._items), received=len(input))
            return None

        result = []
        for i, (schema, value) in enumerate(zip(self._items, input)):
            child_ctx = ctx.child(i)
            parsed = schema._run_pipeline(value, child_ctx)
            result.append(parsed)

        return result

    def _to_node(self) -> dict[str, Any]:
        return self._add_common_node_fields({
            "kind": "tuple",
            "elements": [s._to_node() for s in self._items],
        })
