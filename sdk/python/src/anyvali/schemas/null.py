"""Null schema."""

from __future__ import annotations

from typing import Any

from ..issue_codes import INVALID_TYPE
from .base import BaseSchema, ValidationContext, _anyvali_type_name


class NullSchema(BaseSchema[None]):
    """Schema that only accepts None/null."""

    def __init__(self) -> None:
        super().__init__()

    def _accepts_none(self) -> bool:
        return True

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        if input is not None:
            received = _anyvali_type_name(input)
            ctx.add_issue(INVALID_TYPE, f"Expected null, received {received}", expected="null", received=received)
            return None
        return None

    def _to_node(self) -> dict[str, Any]:
        return self._add_common_node_fields({"kind": "null"})
