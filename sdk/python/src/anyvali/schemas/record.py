"""Record schema (dict with uniform value type)."""

from __future__ import annotations

from typing import Any

from ..issue_codes import INVALID_TYPE
from .base import BaseSchema, ValidationContext, _anyvali_type_name


class RecordSchema(BaseSchema[dict[str, Any]]):
    """Schema for record/dict with string keys and uniform value schema."""

    def __init__(self, value_schema: BaseSchema) -> None:
        super().__init__()
        self._value_schema = value_schema

    def _validate(self, input: Any, ctx: ValidationContext) -> Any:
        if not isinstance(input, dict):
            received = _anyvali_type_name(input)
            ctx.add_issue(INVALID_TYPE, f"Expected object, received {received}", expected="object", received=received)
            return None

        result: dict[str, Any] = {}
        for key, value in input.items():
            if not isinstance(key, str):
                ctx.add_issue(INVALID_TYPE, f"Record keys must be strings, received {type(key).__name__}", expected="string", received=type(key).__name__)
                continue
            child_ctx = ctx.child(key)
            parsed = self._value_schema._run_pipeline(value, child_ctx)
            result[key] = parsed

        return result

    def _to_node(self) -> dict[str, Any]:
        return self._add_common_node_fields({
            "kind": "record",
            "values": self._value_schema._to_node(),
        })
