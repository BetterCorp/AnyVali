"""Nullable schema wrapper."""

from __future__ import annotations

from typing import Any, Generic, TypeVar

from .base import BaseSchema, ValidationContext

T = TypeVar("T")


class NullableSchema(BaseSchema[T | None], Generic[T]):
    """Wraps a schema to allow None/null values."""

    def __init__(self, inner: BaseSchema[T]) -> None:
        super().__init__()
        self._inner = inner

    def _accepts_none(self) -> bool:
        return True

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        if input is None:
            return None
        return self._inner._validate(input, ctx)

    def _run_pipeline(self, input: Any, ctx: ValidationContext) -> Any:
        if input is None:
            return None
        return self._inner._run_pipeline(input, ctx)

    def _to_node(self) -> dict[str, Any]:
        return self._add_common_node_fields({
            "kind": "nullable",
            "schema": self._inner._to_node(),
        })
