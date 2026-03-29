"""Never schema."""

from __future__ import annotations

from typing import Any

from ..issue_codes import INVALID_TYPE
from .base import BaseSchema, ValidationContext


class NeverSchema(BaseSchema):
    """Schema that rejects all values."""

    def __init__(self) -> None:
        super().__init__()

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        ctx.add_issue(INVALID_TYPE, "No value is allowed", expected="never", received=type(input).__name__)
        return None

    def _to_node(self) -> dict[str, Any]:
        return self._add_common_node_fields({"kind": "never"})
