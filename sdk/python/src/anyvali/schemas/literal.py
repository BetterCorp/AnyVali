"""Literal schema."""

from __future__ import annotations

from typing import Any

from ..issue_codes import INVALID_LITERAL
from .base import BaseSchema, ValidationContext


class LiteralSchema(BaseSchema):
    """Schema that accepts only a specific literal value."""

    def __init__(self, value: Any) -> None:
        super().__init__()
        self._value = value

    def _accepts_none(self) -> bool:
        return self._value is None

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        # Strict type + value check (no bool/int coercion)
        if type(input) is not type(self._value) or input != self._value:
            ctx.add_issue(
                INVALID_LITERAL,
                f"Expected literal {self._value!r}, received {input!r}",
                expected=self._value,
                received=input,
            )
            return None
        return input

    def _to_node(self) -> dict[str, Any]:
        return self._add_common_node_fields({"kind": "literal", "value": self._value})
