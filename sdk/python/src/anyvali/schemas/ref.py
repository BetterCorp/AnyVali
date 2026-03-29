"""Ref schema for definition references."""

from __future__ import annotations

from typing import Any

from ..issue_codes import INVALID_TYPE
from .base import BaseSchema, ValidationContext


class RefSchema(BaseSchema):
    """Schema that references a named definition."""

    def __init__(self, ref: str) -> None:
        super().__init__()
        self._ref = ref
        self._resolved: BaseSchema | None = None
        # Lazy resolution via definitions dict
        self._definitions: dict[str, BaseSchema] | None = None

    def resolve(self, schema: BaseSchema) -> None:
        """Resolve this reference to a concrete schema."""
        self._resolved = schema

    def set_definitions(self, definitions: dict[str, BaseSchema]) -> None:
        """Set the definitions dict for lazy resolution."""
        self._definitions = definitions

    def _get_resolved(self) -> BaseSchema | None:
        """Get the resolved schema, trying lazy resolution if needed."""
        if self._resolved is not None:
            return self._resolved
        if self._definitions is not None:
            ref_name = self._ref
            if ref_name.startswith("#/definitions/"):
                ref_name = ref_name[len("#/definitions/"):]
            if ref_name in self._definitions:
                return self._definitions[ref_name]
        return None

    def _accepts_none(self) -> bool:
        resolved = self._get_resolved()
        if resolved is not None:
            return resolved._accepts_none()
        return False

    def _run_pipeline(self, input: Any, ctx: ValidationContext) -> Any:
        """Override to delegate the full pipeline to the resolved schema."""
        resolved = self._get_resolved()
        if resolved is not None:
            return resolved._run_pipeline(input, ctx)
        # Try context definitions
        ref_name = self._ref
        if ref_name.startswith("#/definitions/"):
            ref_name = ref_name[len("#/definitions/"):]
        if ref_name in ctx.definitions:
            return ctx.definitions[ref_name]._run_pipeline(input, ctx)
        ctx.add_issue(INVALID_TYPE, f"Unresolved reference: {self._ref}", expected=self._ref)
        return None

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        resolved = self._get_resolved()
        if resolved is not None:
            return resolved._validate(input, ctx)
        ctx.add_issue(INVALID_TYPE, f"Unresolved reference: {self._ref}", expected=self._ref)
        return None

    def _to_node(self) -> dict[str, Any]:
        return self._add_common_node_fields({"kind": "ref", "ref": self._ref})
