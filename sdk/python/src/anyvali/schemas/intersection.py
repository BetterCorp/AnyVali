"""Intersection schema."""

from __future__ import annotations

from typing import Any

from ..issue_codes import INVALID_TYPE
from .base import BaseSchema, ValidationContext


class IntersectionSchema(BaseSchema[Any]):
    """Schema that requires input to match all given schemas."""

    def __init__(self, schemas: list[BaseSchema]) -> None:
        super().__init__()
        self._schemas = list(schemas)

    def _accepts_none(self) -> bool:
        return all(s._accepts_none() for s in self._schemas)

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        # Run all schemas against the original input, collecting all issues.
        # For object schemas, merge the results.
        results: list[Any] = []
        has_failure = False
        for schema in self._schemas:
            trial_ctx = ValidationContext(
                path=list(ctx.path),
                issues=[],
                definitions=ctx.definitions,
            )
            result = schema._run_pipeline(input, trial_ctx)
            results.append(result)
            if trial_ctx.issues:
                has_failure = True
                ctx.issues.extend(trial_ctx.issues)

        if has_failure:
            return None

        # Merge dict results
        if all(isinstance(r, dict) for r in results):
            merged: dict[str, Any] = {}
            for r in results:
                merged.update(r)
            return merged

        # For non-dict, return the last result
        return results[-1] if results else input

    def _to_node(self) -> dict[str, Any]:
        return self._add_common_node_fields({
            "kind": "intersection",
            "allOf": [s._to_node() for s in self._schemas],
        })
