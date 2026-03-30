"""Union schema."""

from __future__ import annotations

from typing import Any

from ..issue_codes import INVALID_UNION
from .base import BaseSchema, ValidationContext


class UnionSchema(BaseSchema[Any]):
    """Schema that accepts any of the given schemas (first match wins)."""

    def __init__(self, schemas: list[BaseSchema]) -> None:
        super().__init__()
        self._schemas = list(schemas)

    def _accepts_none(self) -> bool:
        return any(s._accepts_none() for s in self._schemas)

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        for schema in self._schemas:
            # Try each schema independently
            trial_ctx = ValidationContext(
                path=list(ctx.path),
                issues=[],
                definitions=ctx.definitions,
            )
            result = schema._run_pipeline(input, trial_ctx)
            if not trial_ctx.issues:
                return result

        ctx.add_issue(
            INVALID_UNION,
            "Input does not match any schema in the union",
            received=type(input).__name__,
        )
        return None

    def _to_node(self) -> dict[str, Any]:
        return self._add_common_node_fields({
            "kind": "union",
            "variants": [s._to_node() for s in self._schemas],
        })
