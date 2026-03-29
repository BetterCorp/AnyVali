"""Optional schema wrapper."""

from __future__ import annotations

from typing import Any

from .base import BaseSchema, ValidationContext, _SENTINEL


class OptionalSchema(BaseSchema):
    """Wraps a schema to make it optional (accepts absent/undefined values)."""

    def __init__(self, inner: BaseSchema) -> None:
        super().__init__()
        self._inner = inner
        # Inherit defaults from inner
        if inner._has_default:
            self._has_default = True
            self._default_value = inner._default_value

    def _accepts_none(self) -> bool:
        return self._inner._accepts_none()

    def _run_pipeline(self, input: Any, ctx: ValidationContext) -> Any:
        # If absent (sentinel), return None without error
        if input is _SENTINEL:
            if self._has_default:
                import copy
                return copy.deepcopy(self._default_value)
            return None
        return self._inner._run_pipeline(input, ctx)

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        return self._inner._validate(input, ctx)

    def _to_node(self) -> dict[str, Any]:
        return self._add_common_node_fields({
            "kind": "optional",
            "schema": self._inner._to_node(),
        })
