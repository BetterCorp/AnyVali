"""Enum schema."""

from __future__ import annotations

from typing import Any

from ..issue_codes import INVALID_TYPE
from .base import BaseSchema, ValidationContext


class EnumSchema(BaseSchema[Any]):
    """Schema that accepts one of a fixed set of values."""

    def __init__(self, values: list[Any]) -> None:
        super().__init__()
        self._values = list(values)

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        if input not in self._values:
            values_str = ",".join(str(v) for v in self._values)
            ctx.add_issue(
                INVALID_TYPE,
                f"Expected one of enum({values_str}), received {input!r}",
                expected=f"enum({values_str})",
                received=str(input),
            )
            return None
        # Also check type strictness -- "1" should not match 1
        for v in self._values:
            if v == input and type(v) is type(input):
                return input
        # Type mismatch
        values_str = ",".join(str(v) for v in self._values)
        ctx.add_issue(
            INVALID_TYPE,
            f"Expected one of enum({values_str}), received {input!r}",
            expected=f"enum({values_str})",
            received=str(input),
        )
        return None

    def _to_node(self) -> dict[str, Any]:
        return self._add_common_node_fields({"kind": "enum", "values": list(self._values)})
